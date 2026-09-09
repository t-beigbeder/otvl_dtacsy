package sftputil

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/pkg/sftp"
	"github.com/t-beigbeder/vdasync/internal/common"
	"golang.org/x/crypto/ssh"
)

func GetSftpClient(user, address, identity, knownHostsFile string) (*sftp.Client, error) {
	if identity == "" {
		return nil, errors.New("GetSftpClient: missing identity file")
	}
	key, err := os.ReadFile(identity)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}
	algorithms := ssh.SupportedAlgorithms()
	var khss []string
	if knownHostsFile != "" {
		khlns, err := common.FileLines(knownHostsFile)
		if err != nil {
			return nil, err
		}
		for _, khln := range khlns {
			_, _, pubKey, _, _, err := ssh.ParseKnownHosts([]byte(khln))
			if err != nil {
				return nil, err
			}
			khss = append(khss, pubKey.Type()+" "+base64.StdEncoding.EncodeToString(pubKey.Marshal()))
		}
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
			ssh.Password(string(key)), // impractical for user, just for testing
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			if knownHostsFile == "" {
				return nil
			}
			hks := key.Type() + " " + base64.StdEncoding.EncodeToString(key.Marshal())
			for _, khs := range khss {
				if hks == khs {
					return nil
				}
			}
			return fmt.Errorf("unkown host key %s for %s", hks, hostname)
		},
		HostKeyAlgorithms: algorithms.HostKeys,
	}
	sc, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, err
	}
	sfc, err := sftp.NewClient(sc)
	if err != nil {
		return nil, err
	}
	return sfc, nil
}
