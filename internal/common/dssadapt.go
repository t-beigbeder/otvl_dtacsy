package common

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/dssagrpc"
)

func DssDte2GrpcDte(ddte *dssa.DataEntry) *dssagrpc.DataEntry {
	var sErr string
	if ddte.Error != nil {
		sErr = ddte.Error.Error()
	}
	return &dssagrpc.DataEntry{
		IsDir:         ddte.IsDir,
		Path:          ddte.Path,
		Size:          ddte.Size,
		Mtime:         ddte.Mtime,
		User:          int32(ddte.User),
		UserRights:    DssRights2GrpcRights(ddte.UserRights),
		Group:         int32(ddte.Group),
		GroupRights:   DssRights2GrpcRights(ddte.GroupRights),
		OtherRights:   DssRights2GrpcRights(ddte.OtherRights),
		IsSymLink:     ddte.IsSymLink,
		SymLinkTarget: ddte.SymLinkTarget,
		Error:         sErr,
		ErrNotExist:   ddte.ErrNotExist,
		Id:            ddte.Id,
		AddMeta:       bytes.Clone(ddte.AddMeta),
	}
}

func DssRights2GrpcRights(drts dssa.Rights) *dssagrpc.Rights {
	return &dssagrpc.Rights{Read: drts.Read, Write: drts.Write, Execute: drts.Execute}
}

func GrpcDte2DssDte(gdte *dssagrpc.DataEntry) *dssa.DataEntry {
	var err error
	if gdte.Error != "" {
		err = errors.New(gdte.Error)
	}

	return &dssa.DataEntry{
		IsDir:         gdte.IsDir,
		Path:          gdte.Path,
		Size:          gdte.Size,
		Mtime:         gdte.Mtime,
		User:          int(gdte.User),
		UserRights:    dssa.Rights{Read: gdte.UserRights.Read, Write: gdte.UserRights.Write, Execute: gdte.UserRights.Execute},
		Group:         int(gdte.Group),
		GroupRights:   dssa.Rights{Read: gdte.GroupRights.Read, Write: gdte.GroupRights.Write, Execute: gdte.GroupRights.Execute},
		OtherRights:   dssa.Rights{Read: gdte.OtherRights.Read, Write: gdte.OtherRights.Write, Execute: gdte.OtherRights.Execute},
		NoLStat:       gdte.NoLstat,
		IsSymLink:     gdte.IsSymLink,
		SymLinkTarget: gdte.SymLinkTarget,
		Error:         err,
		ErrNotExist:   gdte.ErrNotExist,
		Id:            gdte.Id,
		AddMeta:       bytes.Clone(gdte.AddMeta),
	}
}

func RelPath(fullPath, rootPath string) string {
	if fullPath == rootPath {
		return ""
	}
	if rootPath != "/" {
		rootPath += "/"
	}
	return strings.Replace(fullPath, rootPath, "", 1)
}

func ParentPath(path_ string) string {
	pp := path.Dir(path_)
	if pp == "." {
		pp = ""
	}
	return pp
}

func MakeParents(dss dssa.Dssa, path_ string) error {
	de, _ := dss.Stat(path_)
	if de.Error != nil && !de.ErrNotExist {
		return de.Error
	}
	if de.Error == nil {
		return nil
	}
	if path_ == "/" {
		return errors.New("cannot Mkdir \"/\"")
	}
	if err := MakeParents(dss, path.Dir(path_)); err != nil {
		return err
	}
	if err := dss.Mkdir(&dssa.DataEntry{Path: path_, UserRights: dssa.Rights{Read: true, Write: true, Execute: true}}); err != nil {
		// someone else did it?
		de, _ := dss.Stat(path_)
		if de.Error != nil && !de.ErrNotExist {
			return de.Error
		}
		if de.Error == nil {
			return nil
		}
		return err
	}
	return nil
}

func CopyEntry(dss dssa.Dssa, old, new_ string) error {
	rr, err := dss.GetReadCloser(old)
	if err != nil {
		return err
	}
	defer rr.Close()
	wr, err := dss.GetWriteCloser(new_)
	if err != nil {
		return err
	}
	defer wr.Close()
	_, err = io.Copy(wr, rr)
	if err != nil {
		return err
	}
	return wr.Close()
}

func ChmodRO(dss dssa.Dssa, path_ string) error {
	de, err := dss.Stat(path_)
	if err != nil {
		return err
	}
	de.UserRights.Write = false
	de.GroupRights.Write = false
	de.OtherRights.Write = false
	return dss.SetStat(de, false, false)
}

func ChmodRW(dss dssa.Dssa, path_ string) error {
	de, err := dss.Stat(path_)
	if err != nil {
		return err
	}
	de.UserRights.Write = true
	return dss.SetStat(de, false, false)
}

func DssaEntryChecksum(dss dssa.Dssa, path_ string, algos string) (string, error) {
	rc, err := dss.GetReadCloser(path_)
	if err != nil {
		return "", err
	}
	defer rc.Close()
	return ReaderChecksum(rc, algos)
}

func DataEntryMod(de *dssa.DataEntry) uint32 {
	mod := uint32(Rights2Mod([3]dssa.Rights{de.UserRights, de.GroupRights, de.OtherRights}))
	return mod
}

func rightsList(rights dssa.Rights) string {
	rs := "-"
	if rights.Read {
		rs = "r"
	}
	ws := "-"
	if rights.Write {
		ws = "w"
	}
	xs := "-"
	if rights.Execute {
		xs = "x"
	}
	return rs + ws + xs
}

func DataEntryList(de *dssa.DataEntry, isNoOwn bool, cs string) string {
	tp := "-"
	if de.IsDir {
		tp = "d"
	} else if de.IsSymLink {
		tp = "l"
	}
	ownDisp := ""
	if !isNoOwn {
		ownDisp = fmt.Sprintf(" %6d %6d ", de.User, de.Group)
	}
	lt := ""
	if de.IsSymLink {
		lt = fmt.Sprintf(" -> %s", de.SymLinkTarget)
	}
	csd := ""
	if cs != "" {
		csd = " " + cs
	}
	return fmt.Sprintf("%s%s%s%s%s%10d %21s %s%s%s", tp,
		rightsList(de.UserRights), rightsList(de.GroupRights), rightsList(de.OtherRights),
		ownDisp, de.Size, time.Unix(de.Mtime, 0).UTC().Format(time.RFC3339), de.Path,
		lt, csd,
	)
}
