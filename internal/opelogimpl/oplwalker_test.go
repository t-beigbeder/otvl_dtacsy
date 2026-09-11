package opelogimpl

import (
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
)

func TestOplWalker(t *testing.T) {
	//t.Skip("wip")
	lgr := common.DbgLogger()
	var err error
	//lgr, err = common.CliLogger("TestOplWalker", "DEBUG+2", "stderr")
	lgr = common.InfoLogger()
	//lgr = common.GetLogger()
	require.NoError(t, err)

	lgr.Debug("TestOplWalker: started")
	std := t.TempDir()
	require.NoError(t, common.FileTreeGenerate(std, 100, 3000, 2, 4096, false, 2))
	lgr.Debug("TestM2fOpeLogs: FileTreeGenerated")

	ltd := t.TempDir()
	ttd := t.TempDir()

	oplm, err := MakeM2fManager(path.Join(ltd, "m2f.opl"))
	require.NoError(t, err)
	require.NoError(t, oplm.Create(std, ttd))

	ow := NewOplWalker(
		lgr, 4, nil, oplm,
		&config.OpeLogOptionsType{
			Goals:      "load", // load, create, update/remove, verify
			SyncPeriod: int64(5 * time.Second),
		},
		localfiles.MakeLocalFilesDssa(), localfiles.MakeLocalFilesDssa(), std, ttd)
	err = ow.Run()
	require.NoError(t, err)
}
