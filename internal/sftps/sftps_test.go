package sftps

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
)

func TestSftps(t *testing.T) {
	lgr := common.DbgLogger()
	user, address, identity, root := GetSftpsEnv()
	var cb ShutdownCb
	var gerr error
	go func() {
		cb, gerr = RunInsecureSftpServer(lgr, user, address, identity, root)
		if gerr != nil {
			return
		}
		gerr = errors.New("RunInsecureSftpServer terminated")
	}()
	time.Sleep(time.Second)
	if gerr != nil {
		lgr.Error(t.Name(), "err", gerr)
	}
	if cb != nil {
		cb()
	}
	require.NoError(t, gerr)
}
