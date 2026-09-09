package config

import (
	"path"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
)

const CliConfigSample1Yaml string = `
plugins:
- name: localFilesSample
  type: localFiles
  port: 10314
  addArgs:
  - "-is-plugin"
`

const CliConfigSample2Yaml string = `
plugins:
- name: localFilesSample
  type: localFiles
  port: 10314
  addArgs:
  - "-is-plugin"
vdaServers:
- name: localFileSystem
  type: localFiles
- name: pluginSample
  type: plugin
  pluginName: localFilesSample
- name: remoteSample
  type: remote
  host: localhost
  port: 10443
  tls: true
  caCertPath: x509/ca_cert.pem
  clientCertPath: x509/client_cert.pem
  clientKeyPath: x509/client_key.pem
sftpServers:
- name: sftp1
  address: sftp1-server:22
  user: sftp-user
  identity: /home/guest/.ssh/id_ssh_sync
  root: /home/sftp-user/sftp-server-root
  concurrency: 5
  knownHostsFile: /home/guest/.ssh/sftp1_known
`

const CliConfigSample3Yaml string = `
plugins:
- executablePath: /usr/bin/echo
`

func TestLoadConfig(t *testing.T) {
	config1, err := Load(CliConfigSample1Yaml)
	if err != nil {
		t.Error(err)
	}
	require.Nil(t, err)
	require.Equal(t, "0.1", config1.Version)
	require.Equal(t, 1, len(config1.Plugins))

	config2, err := Load(CliConfigSample2Yaml)
	require.Nil(t, err)
	require.Equal(t, 3, len(config2.VdaServers))

	config3, err := Load(CliConfigSample3Yaml)
	require.Nil(t, err)
	require.Equal(t, 1, len(config3.Plugins))
	require.Equal(t, "default", config3.Plugins[0].Name)
	require.Equal(t, DefaultPluginType, config3.Plugins[0].Type)
}

func Test2seeOplConfig(t *testing.T) {
	oplc := &OpeLogOptionsType{
		SyncOptionsType: SyncOptionsType{
			Dryrun:  true,
			Check:   true,
			CsAlgos: "md5,sha256",
		},
		Goals: "one,two",
	}
	oply, err := yaml.Marshal(&oplc)
	require.NoError(t, err)
	td := t.TempDir()
	common.WriteFile(path.Join(td, "sy.yaml"), oply)
}
