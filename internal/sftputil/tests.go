package sftputil

import (
	"fmt"
	"os"

	"github.com/t-beigbeder/vdasync/internal/common"
)

func GetSftpsEnv() (user, address, identity, root string, err error) {
	user = os.Getenv("OTVL_TEST_SF_US")
	if user == "" {
		user = "sftp-user"
	}
	address = os.Getenv("OTVL_TEST_SF_AD")
	if address == "" {
		var port int
		port, err = common.GetFreePort()
		if err != nil {
			return
		}
		address = fmt.Sprintf("localhost:%d", port)
	}
	identity = os.Getenv("OTVL_TEST_SF_ID")
	if identity == "" {
		identity = "/local/tmp/id_ssh_test"
	}
	root = os.Getenv("OTVL_TEST_SF_ROOT")
	if root == "" {
		root = "/local/tmp"
	}
	return
}
