package config

import (
	"github.com/goccy/go-yaml"
)

type PluginsOptionsType struct {
	NoTls          bool   `yaml:"noTls"`
	Insecure       bool   `yaml:"insecure"`
	CertPath       string `yaml:"certPath"`
	KeyPath        string `yaml:"keyPath"`
	CaCertPath     string `yaml:"caCertPath"`
	ClientCertPath string `yaml:"clientCertPath"`
	ClientKeyPath  string `yaml:"clientKeyPath"`
}

type PluginType struct {
	Name           string   `yaml:"name"`
	Type           string   `yaml:"type"`
	ExecutablePath string   `yaml:"executablePath"`
	AddArgs        []string `yaml:"addArgs"`
	Port           int      `yaml:"port"`
}

type VdaServerType struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	Insecure       bool   `yaml:"insecure"`
	NoTls          bool   `yaml:"noTls"`
	CaCertPath     string `yaml:"caCertPath"`
	ClientCertPath string `yaml:"clientCertPath"`
	ClientKeyPath  string `yaml:"clientKeyPath"`
}

type SyncOptionsType struct {
	Dryrun       bool   `yaml:"dryrun"`
	Check        bool   `yaml:"check"`
	CsAlgos      string `yaml:"csAlgos"`
	NoPerm       bool   `yaml:"noPerm"`
	NoMtime      bool   `yaml:"noMtime"`
	NoMtLink     bool   `yaml:"noMtLink"`
	Rm           bool   `yaml:"rm"`
	Force        bool   `yaml:"force"`
	IgnoreIrreg  bool   `yaml:"ignoreIrreg"`
	ExclListPath string `yaml:"exclListPath"`
	InclListPath string `yaml:"inclListPath"`
}

type OpeLogOptionsType struct {
	SyncOptionsType `yaml:"syncOptionsType"`
	Goals           string `yaml:"goals"`
	StatsTime       int64  `yaml:"statsTime"`
}

type SftpServerType struct {
	Name           string `yaml:"name"`
	User           string `yaml:"user"`
	Address        string `yaml:"address"`
	Identity       string `yaml:"identity"`
	Root           string `yaml:"root"`
	Concurrency    int    `yaml:"concurrency"`
	KnownHostsFile string `yaml:"knownHostsFile"`
}

type CliConfig struct {
	Version            string              `yaml:"version"`
	PluginsOptions     *PluginsOptionsType `yaml:"pluginsOptions"`
	Plugins            []*PluginType       `yaml:"plugins"`
	PluginReadyRetries int                 `yaml:"pluginReadyRetries"`
	PluginReadyTimeout string              `yaml:"pluginReadyTimeout"`
	PluginAddress      string              `yaml:"pluginAddress"`
	VdaServers         []*VdaServerType    `yaml:"vdaServers"`
	SftpServers        []*SftpServerType   `yaml:"sftpServers"`
	SyncOptions        *SyncOptionsType    `yaml:"syncOptions"`
}

const CliConfigDefaultYaml string = `
version: "0.1"
pluginsOptions:
  noTls: false
  insecure: false
plugins:
pluginReadyRetries: 4
pluginReadyTimeout: "100ms"
pluginAddress: "localhost"
`

const PluginTypeDefaultYaml string = `
name: "default"
type: "localFiles"
`

var defaultPluginTypeValues = &PluginType{}

var DefaultPluginType = "localFiles"

func init() {
	yaml.Unmarshal([]byte(PluginTypeDefaultYaml), &defaultPluginTypeValues)
}

func umarshalPlugin(op *PluginType, b []byte) error {
	tp := &PluginType{}
	*tp = *defaultPluginTypeValues
	if err := yaml.Unmarshal(b, tp); err != nil {
		return err
	}
	*op = *tp
	return nil
}

func Load(config string) (*CliConfig, error) {
	conf := CliConfig{}
	if err := yaml.Unmarshal([]byte(CliConfigDefaultYaml), &conf); err != nil {
		return nil, err
	}
	if err := yaml.UnmarshalWithOptions([]byte(config), &conf, yaml.CustomUnmarshaler(umarshalPlugin)); err != nil {
		return nil, err
	}
	return &conf, nil
}

func VdaServer(cfg *CliConfig, host string, port int) *VdaServerType {
	if cfg == nil {
		cfg = &CliConfig{}
	}
	for _, vs := range cfg.VdaServers {
		if (vs.Host == "" || vs.Host == host) && vs.Port == port {
			return vs
		}
	}
	return nil
}

func SftpServer(cfg *CliConfig, name string) (*SftpServerType, []any) {
	if cfg == nil {
		cfg = &CliConfig{}
	}
	for _, sfs := range cfg.SftpServers {
		if sfs.Name == name {
			return sfs, []any{sfs.User, sfs.Address, sfs.Identity, sfs.Root, sfs.Concurrency, sfs.KnownHostsFile}
		}
	}
	return nil, nil
}
