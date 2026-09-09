package dssaimpl

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
)

func TestAll(t *testing.T) {
	t.Skip("wip")
	tds, err := NewTestDss(&TestDssOptions{
		Lgr: common.DbgLogger(),
		Kind: "sftp",
		PluginLevel: "DEBUG",
	})
	require.NoError(t, err)
	require.NoError(t, tds.Close())
}
