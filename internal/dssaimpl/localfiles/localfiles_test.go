package localfiles

import (
	"io"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
)

func TestFileFunctions(t *testing.T) {
	td1 := t.TempDir()
	ft := path.Join(td1, "TestFileFunctions.dat")
	require.Nil(t, common.WriteFile(ft, []byte(t.Name())))
	lfd := MakeLocalFilesDssa()
	de, err := lfd.Stat(ft)
	require.Nil(t, err)
	require.Nil(t, common.Lutimes(ft, de.Mtime-600))
	de2, err := lfd.Stat(ft)
	require.Nil(t, err)
	require.Equal(t, de.Mtime-600, de2.Mtime)
	de2.Mtime = de.Mtime
	de2.GroupRights = dssa.Rights{}
	de2.OtherRights = dssa.Rights{}
	err = lfd.SetStat(de2, false, false)
	require.Nil(t, err)
	de3, err := lfd.Stat(ft)
	require.Nil(t, err)
	require.Equal(t, de.Mtime, de3.Mtime)

	lt := path.Join(t.TempDir(), "TestFileFunctions.symlink")
	err = os.Symlink(ft, lt)
	require.Nil(t, err)
	ltde, err := lfd.Stat(lt)
	require.Nil(t, err)
	require.True(t, ltde.IsSymLink)
	require.Equal(t, ft, ltde.SymLinkTarget)

	ldt := path.Join(t.TempDir(), "TestFileFunctions.lnsd")
	err = os.Symlink(td1, ldt)
	require.Nil(t, err)
	ldtde, err := lfd.Stat(ldt)
	require.Nil(t, err)
	require.True(t, ldtde.IsSymLink)
	require.False(t, ldtde.IsDir)
	require.Equal(t, td1, ldtde.SymLinkTarget)

	tt := path.Join(t.TempDir(), "TestFileFunctionsCopied.dat")
	rc, err := lfd.GetReadCloser(ft)
	require.Nil(t, err)
	defer rc.Close()
	wc, err := lfd.GetWriteCloser(tt)
	require.Nil(t, err)
	defer wc.Close()
	lw, err := io.Copy(wc, rc)
	require.Nil(t, err)
	require.Equal(t, de.Size, lw)

	nef := path.Join(t.TempDir(), "TestFileFunctionsNotYet.dat")
	de4, err := lfd.Stat(nef)
	require.NotNil(t, err)
	require.NotNil(t, de4)
	require.True(t, de4.ErrNotExist)
	require.Equal(t, err, de4.Error)

	td := path.Join(t.TempDir(), "TestFileFunctionsNotYetNewDir")
	err = lfd.Mkdir(&dssa.DataEntry{Path: td, UserRights: dssa.Rights{Read: true, Execute: true, Write: true}})
	require.Nil(t, err)

	err = lfd.Rm(de3.Path)
	require.Nil(t, err)
	de3ne, err := lfd.Stat(de3.Path)
	require.NotNil(t, err)
	require.NotNil(t, de3ne)
	require.True(t, de3ne.ErrNotExist)

	err = lfd.Rm(td)
	require.Nil(t, err)
	dedne, err := lfd.Stat(td)
	require.NotNil(t, err)
	require.NotNil(t, dedne)
	require.True(t, dedne.ErrNotExist)

	ft = path.Join(td1, "TestFileFunctions.dat")
	require.Nil(t, common.WriteFile(ft, []byte("TestChecksum")))
	h1, err := lfd.Checksum("sha256", ft)
	require.Nil(t, err)
	require.Equal(t, "sha256:4b86be7f5fe5776cd535cdf1e81fdd77c204df48c751f61c121b3e72f6767e1e", h1)

	des, err := lfd.List(td1)
	require.NoError(t, err)
	_ = des
}

func TestFileGetPut(t *testing.T) {
	ft := path.Join(t.TempDir(), "TestFileGetPut.dat")
	err := common.MakeTestFile(ft, 32*1024*1024)
	require.Nil(t, err)

	lfd := MakeLocalFilesDssa()
	de, err := lfd.Stat(ft)
	require.Nil(t, err)
	require.Equal(t, 32*1024*1024, int(de.Size))

	tt := path.Join(t.TempDir(), "TestFileGetPutCopied.dat")
	rc, err := lfd.GetReadCloser(ft)
	require.Nil(t, err)
	defer rc.Close()
	wc, err := lfd.GetWriteCloser(tt)
	require.Nil(t, err)
	defer wc.Close()
	written := 0
	buf := make([]byte, 31999)
	for {
		nr, err := rc.Read(buf)
		if err != nil {
			break
		}
		nw, err := wc.Write(buf[0:nr])
		if err != nil {
			break
		}
		written += nw
	}
	require.Nil(t, err)
	require.Equal(t, de.Size, int64(written))
}
