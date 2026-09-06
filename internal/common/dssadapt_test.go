package common

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRP(t *testing.T) {
	td := t.TempDir()
	dt := path.Join(td, "d1")
	ft := path.Join(dt, "f1")
	rdn := RelPath(dt, td)
	require.Equal(t, rdn, "d1")
	rfn := RelPath(ft, td)
	require.Equal(t, rfn, "d1/f1")
	td2 := path.Join(td, "")
	require.Equal(t, td, td2)
	require.Equal(t, "", RelPath(td2, td))
	require.Equal(t, "path1", RelPath("/path/to/path1", "/path/to"))
	require.Equal(t, "path2", RelPath("/path2", "/"))
}
