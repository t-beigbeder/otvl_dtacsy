package sftps

import "os"

func GetSftpsEnv() (user, address, identity, root string) {
	user = os.Getenv("OTVL_TEST_SF_US")
	if user == "" {
		user = "sftp-user"
	}
	address = os.Getenv("OTVL_TEST_SF_AD")
	if address == "" {
		address = "localhost:10022"
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
