package sftputil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
)

func TestSftps(t *testing.T) {
	lgr := common.GetLogger()
	user, address, identity, root, err := GetSftpsEnv()
	require.NoError(t, err)
	var cb ShutdownCb
	var gerr error
	go func() {
		cb, gerr = RunInsecureSftpServer(lgr, user, address, identity, root)
		if gerr != nil {
			return
		}
		lgr.Debug(t.Name(), "terminated", true)
	}()
	time.Sleep(100 * time.Millisecond)
	if gerr != nil {
		lgr.Error(t.Name(), "err", gerr)
	}
	sftc, err := GetSftpClient(user, address, identity, "")
	require.NoError(t, err)
	fi, err := sftc.Lstat(".")
	require.NoError(t, err)
	lgr.Debug(t.Name(), "fi", fi.Name())
	// time.Sleep(100*time.Millisecond)
	if cb != nil {
		cb()
	}
	require.NoError(t, gerr)
}
