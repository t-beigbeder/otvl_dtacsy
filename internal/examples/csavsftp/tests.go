package csavsftp

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/pkg/sftp"
	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
	"golang.org/x/crypto/ssh"
)

func SkipIf(t *testing.T) bool {
	if true { // dev mode
		return false
	}
	if os.Getenv("OTVL_TEST_FULL") == "" {
		return true
	}
	if os.Getenv("OTVL_TEST_SFTP") == "" {
		return true
	}
	return false
}

func GetSftpEnv() (user, address, identity, root string) {
	user = os.Getenv("OTVL_TEST_SF_US")
	if user == "" {
		user = os.Getenv("SYNC_USER")
	}
	if user == "" {
		user = "sftp-user"
	}
	address = os.Getenv("OTVL_TEST_SF_AD")
	if address == "" {
		address = "t-sk3s-sv-ext:22"
	}
	identity = os.Getenv("OTVL_TEST_SF_ID")
	if identity == "" {
		identity = "/local/tmp/id_ssh_test"
	}
	root = os.Getenv("OTVL_TEST_SF_ROOT")
	if root == "" {
		root = "sftp_root"
	}
	return
}

func GetTestSftpClient(user, address, identity, _ string) (*sftp.Client, error) {
	key, err := os.ReadFile("/local/tmp/id_ssh_test")
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}
	algorithms := ssh.SupportedAlgorithms()
	khlns, err := common.FileLines("/local/tmp/known_hosts")
	if err != nil {
		return nil, err
	}
	config := &ssh.ClientConfig{
		Config: ssh.Config{
			KeyExchanges: algorithms.KeyExchanges,
			Ciphers:      algorithms.Ciphers,
			MACs:         algorithms.MACs,
		},
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			hks := key.Type() + " " + base64.StdEncoding.EncodeToString(key.Marshal())
			_ = hks
			for ix, khln := range khlns {
				marker, hosts, pubKey, comment, rest, err := ssh.ParseKnownHosts([]byte(khln))
				_, _, _, _, _, _ = marker, hosts, pubKey, comment, rest, err
				khs := pubKey.Type() + " " + base64.StdEncoding.EncodeToString(pubKey.Marshal())
				if hks == khs {
					_ = ix
					return nil
				}
			}
			return nil
		},
		HostKeyAlgorithms: algorithms.HostKeys,
	}
	sc, err := ssh.Dial("tcp", "t-sk3s-sv-ext:22", config)
	if err != nil {
		return nil, err
	}
	sfc, err := sftp.NewClient(sc)
	if err != nil {
		return nil, err
	}
	return sfc, nil
}

func GetSftpDss(t *testing.T) dssa.Dssa {
	user, address, identity, root := GetSftpEnv()
	dss, err := MakeSftpClientDssa(user, address, identity, root, 4, GetTestSftpClient, "")
	require.NoError(t, err)
	return dss
}

func Cleanup(ds dssa.Dssa) error {
	sf, ok := ds.(*sftpClient)
	if !ok {
		return errors.New("Cleanup: not a sfc")
	}
	if sf.root == "" || sf.root == "/" {
		return fmt.Errorf("Cleanup: remove %s is dangerous", sf.root)
	}
	sfc := <-sf.sfcs
	defer func() {
		sf.sfcs <- sfc
	}()
	if err := sfc.RemoveAll(sf.root); err != nil {
		return fmt.Errorf("Cleanup: %s error %v", sf.root, err)
	}
	return sfc.Mkdir(sf.root)
}
