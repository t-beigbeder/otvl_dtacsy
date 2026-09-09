package opelogimpl

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

func TestM2fOpeLogs(t *testing.T) {
	lgr := common.DbgLogger()
	lgr.Debug("TestM2fOpeLogs: start")

	std := t.TempDir()
	require.NoError(t, common.FileTreeGenerate(std, 250, 25000, 2, 32768, false, 0))
	lgr.Debug("TestM2fOpeLogs: FileTreeGenerated")

	ltd := t.TempDir()
	ttd := t.TempDir()

	olm, err := MakeM2fManager(path.Join(ltd, "m2f.opl"))
	require.NoError(t, err)
	require.NoError(t, olm.Create(std, ttd))
	require.NoError(t, olm.Open(false))
	olm.PutLogicalEntry("", &opelog.LogicalEntry{})
	require.NoError(t, olm.Sync())
	olm.PutLogicalEntry("a", &opelog.LogicalEntry{})
	require.NoError(t, olm.Sync())
	olm2, err := MakeM2fManager(path.Join(ltd, "m2f.opl"))
	require.NoError(t, err)
	require.Error(t, olm2.Open(false))
	require.NoError(t, olm2.Open(true))
	require.NoError(t, olm.Close())
	require.NoError(t, olm2.Close())
	lgr.Debug("TestM2fOpeLogs: end")
}
