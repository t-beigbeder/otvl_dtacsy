package opelogimpl

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
)

func TestOplWalker(t *testing.T) {
	lgr := common.DbgLogger()
	lgr.Debug("TestOplWalker: started")
	std := t.TempDir()
	require.NoError(t, common.FileTreeGenerate(std, 100, 3000, 2, 4096, false, 2))
	lgr.Debug("TestM2fOpeLogs: FileTreeGenerated")

	ltd := t.TempDir()
	ttd := t.TempDir()

	oplm, err := MakeM2fManager(path.Join(ltd, "m2f.opl"))
	require.NoError(t, err)
	require.NoError(t, oplm.Create(std, ttd))

	ow := NewOplWalker(lgr, 4, nil, oplm, localfiles.MakeLocalFilesDssa(), localfiles.MakeLocalFilesDssa())
	require.NoError(t, ow.Run())
}
