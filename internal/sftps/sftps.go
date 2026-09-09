package sftps

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// Server implementation adapted from sftp/examples/go-sftp-server/main.go
// See https://pkg.go.dev/github.com/pkg/sftp for license and copyright

func loadPriK(ident string) (ssh.Signer, error) {
	key, err := os.ReadFile(ident)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}
	return signer, nil
}

func insecConfig(lgr *slog.Logger, user string, prik ssh.Signer) *ssh.ServerConfig {
	config := &ssh.ServerConfig{
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			if conn.User() == user && string(password) == user {
				return nil, nil
			}
			lgr.Error("password callback rejected", "user", conn.User(), "password", string(password))
			return nil, errors.New("sftp user/password rejected")
		},
	}
	config.AddHostKey(prik)
	return config
}

type ShutdownCb func() error

func RunInsecureSftpServer(lgr *slog.Logger, user, address, identity, root string) (ShutdownCb, error) {
	lgr.Info("RunInsecureSftpServer: start", "user", user, "address", address, "identity", identity, "root", root)
	prik, err := loadPriK(identity)
	if err != nil {
		return nil, err
	}
	config := insecConfig(lgr, user, prik)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("listen: %v", err)
	}
	scb := func() error {
		err := listener.Close()
		if err != nil {
			return fmt.Errorf("shutdown sftp server: close %v", err)
		}
		return err
	}
	for {
		nConn, err := listener.Accept()
		if err != nil {
			lgr.Error("accept", "err", err)
			continue
		}
		_, chans, reqs, err := ssh.NewServerConn(nConn, config)
		if err != nil {
			lgr.Error("NewServerConn", "err", err)
			continue
		}
		go ssh.DiscardRequests(reqs)
		for newChannel := range chans {
			lgr.Debug("incoming channel", "channel", newChannel.ChannelType())
			if newChannel.ChannelType() != "session" {
				newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
				continue
			}
			channel, requests, err := newChannel.Accept()
			if err != nil {
				lgr.Error("accept", "err", err)
				continue
			}
			lgr.Debug("channel accepted")
			go func(in <-chan *ssh.Request) {
				for req := range in {
					lgr.Debug("request", "req", req.Type)
					ok := false
					switch req.Type {
					case "subsystem":
						if string(req.Payload[4:]) == "sftp" {
							ok = true
						}
					}
					lgr.Debug("accepted", "ok", ok)
					req.Reply(ok, nil)
				}
			}(requests)
			server, err := sftp.NewServer(channel)
			if err != nil {
				lgr.Error("NewServer", "err", err)
				continue
			}
			if err := server.Serve(); err != nil {
				lgr.Error("server.Serve()", "err", err)
				continue
			}
			server.Close()
			lgr.Debug("server closed")
		}
	}
	return scb, errors.ErrUnsupported
}
