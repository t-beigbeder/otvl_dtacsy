package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/internal/cli"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/sftpc"
	"github.com/t-beigbeder/vdasync/internal/remote"
	"google.golang.org/grpc"
)

func RunSftpPlugin() {
	var (
		hostFlag    = flag.String("host", "localhost", "host/address to listen, defaults to localhost")
		portFlag    = flag.Int("port", 0, "port to listen")
		nameFlag    = flag.String("name", "", "plugin name")
		typeFlag    = flag.String("type", "", "plugin type")
		sftpUser    = flag.String("sftpuser", "", "SFTP server login")
		sftpAddress = flag.String("sftpaddress", "localhost:22", "SFTP server address")
		sftpIdent   = flag.String("sftpident", "", "SSH identity file to authenticate")
		sftpRoot    = flag.String("sftproot", "", "root path from SFTP server root where files are served")
		sftpKHFile  = flag.String("sftpkhfile", "", "known_hosts file, defaults to $HOME/.ssh/known_hosts")
		sftpNoHKC   = flag.Bool("sftpnohkc", false, "ignore host key, insecure, equivalent of ssh StrictHostKeychecking=no")
	)
	cf := cli.CommonFlags()
	flag.Parse()
	exe, err := os.Executable()
	if err != nil {
		common.Fatal(nil, fmt.Errorf("os.Executable: %v", err))
	}
	cmd := path.Base(exe)
	lgr, err := common.CliLogger(cmd, *cf.LogLevelFlag, *cf.LogFlag)
	if err != nil {
		common.Fatal(lgr, fmt.Errorf("path.Base: %s: %v", exe, err))
	}

	knownHostsFile := ""
	if !*sftpNoHKC {
		if *sftpKHFile == "" {
			knownHostsFile = path.Join(os.Getenv("HOME"), ".ssh", "known_hosts")
		} else {
			knownHostsFile = *sftpKHFile
		}
		if !common.FileExists(knownHostsFile) {
			common.Fatal(lgr, fmt.Errorf("sftpkhfile: %s does not exist", knownHostsFile))
		}
	}
	if *sftpUser == "" {
		common.Fatal(lgr, errors.New("sftuser empty"))
	}
	if *sftpIdent == "" {
		common.Fatal(lgr, errors.New("sftpident empty"))
	}
	if *sftpRoot == "" {
		common.Fatal(lgr, errors.New("sftproot empty"))
	}
	dss, err := sftpc.MakeSftpClientDssa(*sftpUser, *sftpAddress, *sftpIdent, *sftpRoot, *cf.ConcurrencyFlag, sftpc.GetSftpClient, knownHostsFile)
	if err != nil {
		common.Fatal(lgr, fmt.Errorf("sftpc.MakeSftpClientDssa: %s: %v", exe, err))
	}

	sop, err := cli.GetServerOrPluginTls(cf)
	if err != nil {
		common.Fatal(lgr, err)
	}
	var sops []grpc.ServerOption
	if sop != nil {
		sops = []grpc.ServerOption{sop}
	}

	lgr.Info(fmt.Sprintf("%s.main starting", cmd), "version", config.GetVersion(), "name", *nameFlag, "type", *typeFlag, "host", *hostFlag, "port", *portFlag)
	done := make(chan bool)
	cb := func() {
		lgr.Debug("shutdownCb called, closing done")
		close(done)
	}
	_, _, err = remote.RunOpeDssaServer(
		lgr, context.Background(), *hostFlag, *portFlag,
		sops, dss, cb, nil)
	<-done
	if err != nil {
		common.Fatal(lgr, fmt.Errorf("RunOpeDssaServer failed %s", err))
	}
	lgr.Info(fmt.Sprintf("%s.main done", cmd), "host", *hostFlag, "port", *portFlag)
}

func main() {
	RunSftpPlugin()
}
