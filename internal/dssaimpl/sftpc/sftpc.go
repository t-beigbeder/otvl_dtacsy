package sftpc

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"time"

	"github.com/pkg/sftp"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/sftputil"
)

type sftpClient struct {
	sfcs chan *sftp.Client
	root string
}

// EndSession implements [dssa.Dssa].
func (sf *sftpClient) EndSession() error {
	return nil
}

// GetReadCloser implements [dssa.Dssa].
func (sf *sftpClient) GetReadCloser(path_ string) (io.ReadCloser, error) {
	sfc := <-sf.sfcs
	rr, err := sfc.Open(path.Join(sf.root, path_))
	if err != nil {
		sf.sfcs <- sfc
		return nil, err
	}
	return &sftpReader{reader: rr, cb: func() { sf.sfcs <- sfc }}, nil
}

// GetWriteCloser implements [dssa.Dssa].
func (sf *sftpClient) GetWriteCloser(path_ string) (io.WriteCloser, error) {
	sfc := <-sf.sfcs
	wr, err := sfc.Create(path.Join(sf.root, path_))
	if err != nil {
		sf.sfcs <- sfc
		return nil, err
	}
	return &sftpWriter{writer: wr, cb: func() { sf.sfcs <- sfc }}, nil
}

// List implements [dssa.Dssa].
func (sf *sftpClient) List(path_ string) ([]*dssa.DataEntry, error) {
	sfc := <-sf.sfcs
	defer func() {
		sf.sfcs <- sfc
	}()
	fis, err := sfc.ReadDir(path.Join(sf.root, path_))
	if err != nil {
		return nil, err
	}
	des := []*dssa.DataEntry{}
	for _, fi := range fis {
		sfi, ok := fi.Sys().(*sftp.FileStat)
		if !ok {
			return nil, errors.New("sftpClient.List: not a sftp.FileStat")
		}
		rights := common.Perm2Rights(sfi.FileMode())
		des = append(des, &dssa.DataEntry{
			Path:        path.Join(path_, fi.Name()),
			IsDir:       fi.IsDir(),
			NoLStat:     true,
			Size:        fi.Size(),
			Mtime:       fi.ModTime().Unix(),
			User:        int(sfi.UID),
			UserRights:  rights[0],
			Group:       int(sfi.GID),
			GroupRights: rights[1],
			OtherRights: rights[2],
		})
	}
	return des, nil
}

// Mkdir implements [dssa.Dssa].
func (sf *sftpClient) Mkdir(de *dssa.DataEntry) error {
	sfc := <-sf.sfcs
	defer func() {
		sf.sfcs <- sfc
	}()
	return sfc.Mkdir(path.Join(sf.root, de.Path))
}

// NewSession implements [dssa.Dssa].
func (sf *sftpClient) NewSession() error {
	return nil
}

// Rm implements [dssa.Dssa].
func (sf *sftpClient) Rm(path_ string) error {
	sfc := <-sf.sfcs
	defer func() {
		sf.sfcs <- sfc
	}()
	return sfc.Remove(path.Join(sf.root, path_))
}

// SetStat implements [dssa.Dssa].
func (sf *sftpClient) SetStat(de *dssa.DataEntry, noPerm bool, noMtime bool) error {
	sfc := <-sf.sfcs
	defer func() {
		sf.sfcs <- sfc
	}()
	fp := path.Join(sf.root, de.Path)
	if !noPerm && !de.IsSymLink {
		ugor := []dssa.Rights{de.UserRights, de.GroupRights, de.OtherRights}
		if err := sfc.Chmod(fp, common.Rights2Mod([3]dssa.Rights(ugor))); err != nil {
			return err
		}
	}
	// No Lutime equivalent with sftp client
	if !noMtime && !de.IsSymLink {
		t := time.Unix(de.Mtime, 0)
		if err := sfc.Chtimes(fp, t, t); err != nil {
			return err
		}
	}
	return nil
}

// Stat implements [dssa.Dssa].
func (sf *sftpClient) Stat(path_ string) (*dssa.DataEntry, error) {
	sfc := <-sf.sfcs
	defer func() {
		sf.sfcs <- sfc
	}()
	fp := path.Join(sf.root, path_)
	fi, err := sfc.Lstat(fp)
	if err != nil {
		return &dssa.DataEntry{Path: path_, Error: err, ErrNotExist: os.IsNotExist(err)}, err
	}
	isSymlink := fi.Mode().Type()&fs.ModeSymlink != 0
	var linkTarget string
	if isSymlink {
		linkTarget, err = sfc.ReadLink(fp)
		if err != nil {
			return nil, err
		}
	}
	sfi, ok := fi.Sys().(*sftp.FileStat)
	if !ok {
		return nil, errors.New("sftpClient.Stat: not a sftp.FileStat")
	}
	ugoRights := common.Perm2Rights(sfi.FileMode().Perm())
	return &dssa.DataEntry{
		IsDir:         fi.IsDir(),
		Path:          path_,
		Size:          fi.Size(),
		Mtime:         fi.ModTime().Unix(),
		User:          int(sfi.UID),
		UserRights:    ugoRights[0],
		Group:         int(sfi.GID),
		GroupRights:   ugoRights[1],
		OtherRights:   ugoRights[2],
		IsSymLink:     isSymlink,
		SymLinkTarget: linkTarget,
	}, nil
}

// Checksum implements [dssa.Dssa].
func (sf *sftpClient) Checksum(algos string, path_ string) (string, error) {
	return common.DssaEntryChecksum(sf, path_, algos)
}

// Symlink implements [dssa.Dssa].
func (sf *sftpClient) Symlink(old string, new_ string) error {
	sfc := <-sf.sfcs
	defer func() {
		sf.sfcs <- sfc
	}()
	return sfc.Symlink(old, path.Join(sf.root, new_))
}

type SftpClientFactory func(user, address, identity, knownHostsFile string) (*sftp.Client, error)

func MakeSftpClientDssa(
	user, address, identity, root string, concurrency int,
	factory SftpClientFactory, knownHostsFile string) (dssa.Dssa, error) {
	sfcs := make(chan *sftp.Client, concurrency+1)
	for i := 0; i <= concurrency; i++ {
		sfc, err := factory(user, address, identity, knownHostsFile)
		if err != nil {
			return nil, fmt.Errorf("MakeSftpClientDssa: factory error count #%d: %v", i, err)
		}
		sfcs <- sfc
	}
	return &sftpClient{sfcs: sfcs, root: root}, nil
}

type DssaMaker struct{}

// MakeDssa implements [dssa.DssaMaker].
func (dmk *DssaMaker) MakeDssa(args ...any) (dssa.Dssa, error) {
	if len(args) != 6 {
		return nil, fmt.Errorf("sftpc.MakeDssa: 6 arguments required, %d given", len(args))
	}
	var (
		user           string
		address        string
		identity       string
		root           string
		concurrency    int
		knownHostsFile string
		ok             bool
	)
	if user, ok = args[0].(string); !ok {
		return nil, errors.New("sftpc.MakeDssa: user incorrect type")
	}
	if address, ok = args[1].(string); !ok {
		return nil, errors.New("sftpc.MakeDssa: address incorrect type")
	}
	if identity, ok = args[2].(string); !ok {
		return nil, errors.New("sftpc.MakeDssa: identity incorrect type")
	}
	if root, ok = args[3].(string); !ok {
		return nil, errors.New("sftpc.MakeDssa: root incorrect type")
	}
	if concurrency, ok = args[4].(int); !ok {
		return nil, errors.New("sftpc.MakeDssa: concurrency incorrect type")
	}
	if knownHostsFile, ok = args[5].(string); !ok {
		return nil, errors.New("sftpc.MakeDssa: knownHostsFile incorrect type")
	}
	return MakeSftpClientDssa(user, address, identity, root, concurrency, sftputil.GetSftpClient, knownHostsFile)
}

var _ dssa.DssaMaker = &DssaMaker{}
