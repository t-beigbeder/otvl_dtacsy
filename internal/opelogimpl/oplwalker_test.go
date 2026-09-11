package opelogimpl

import (
	"os"
	"path"
	"slices"
	"strings"
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
	lgr = common.InfoLogger()
	//lgr = common.GetLogger()
	//lgr, _ = common.CliLogger("TestOplWalker", "DEBUG+2", "")
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
	owi, ok := ow.(*oplWalkerImpl)
	require.True(t, ok)
	oplmi, ok := owi.oplm.(*m2fMng)
	ks := make([]string, 0, len(oplmi.les))
	for k := range oplmi.les {
		ks = append(ks, path.Join(std, k))
	}
	slices.Sort(ks)
	if len(oplmi.les) > 3101 {
		common.WriteFile(path.Join(os.TempDir(), "3102.keys"),[]byte(strings.Join(ks, "\n")))
		lgr.Debug("here")
	} else {
		common.WriteFile(path.Join(os.TempDir(), "3101.keys"),[]byte(strings.Join(ks, "\n")))
		lgr.Debug("there")
	}
	require.LessOrEqual(t, 3000+100+1, len(oplmi.les))
}
