package dssaimpl

import (
	"fmt"
	"log/slog"
	"path"
	"runtime"

	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
	"github.com/t-beigbeder/vdasync/internal/plugin"
)

type TestDss struct {
	kind string
	dss  dssa.Dssa
	rps  []*plugin.RunningPlugin
}

type TestDssOptions struct {
	Kind          string // "localFiles", "sftp"
	Lgr           *slog.Logger
	PluginType    string // defaults to vda + Kind
	PluginAddArgs string // defaults to [-notls, -log, stderr, -level, <LEVEL>]
	PluginLevel   string // defaults to ERROR
	PluginGetArgs func(*TestDssOptions) string
	SftpUser      string
	SftpHost      string
	SftpPort      string
	SftpIdent     string
	SftpRoot      string
	SftpKHFile    string
	SftpHasServer bool
}

func testDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return path.Dir(filename)
}

func (tdo *TestDssOptions) getPluginConfig() (*config.CliConfig, error) {
	pt := tdo.PluginType
	if pt == "" {
		pt = "vda" + tdo.Kind
	}
	pl := tdo.PluginLevel
	if pl == "" {
		pl = "ERROR"
	}
	paa := tdo.PluginAddArgs
	if paa == "" {
		paa = "[-notls, -log, stderr, -level, %s%s]"
	}
	pag := ""
	if tdo.PluginGetArgs != nil {
		pag = tdo.PluginGetArgs(tdo)
	}
	paa = fmt.Sprintf(paa, pl, pag)
	conf := `
pluginsOptions:
  noTls: true
plugins:
- name: %s
  type: %s
  executablePath: %s
  addArgs: %s
`
	exep := path.Clean(testDir() + "/../../bin/lamd64/" + pt)
	conf = fmt.Sprintf(conf, pt, pt, exep, paa)
	return config.Load(conf)
}

func (tdo *TestDssOptions) GetPluginTestDss() (*TestDss, error) {
	cf, err := tdo.getPluginConfig()
	if err != nil {
		return nil, err
	}
	rps, err := plugin.RunCliConfig(tdo.Lgr, cf, nil)
	if err != nil {
		return nil, err
	}
	return &TestDss{rps: rps}, nil
}

func GetSftpArgs(tdo *TestDssOptions) string {
	tpl := ", -sftpuser, %s"
	return fmt.Sprintf(tpl, "tsftpuser")
}

func (tdo *TestDssOptions) getSftpTestDss() (*TestDss, error) {
	if tdo.PluginGetArgs == nil {
		tdo.PluginGetArgs = GetSftpArgs
	}
	testDss, err := tdo.GetPluginTestDss()
	if err != nil {
		return nil, err
	}
	return testDss, nil
}

func NewTestDss(tdo *TestDssOptions) (*TestDss, error) {
	var (
		testDss *TestDss
		err     error
	)
	switch tdo.Kind {
	case "localFiles":
		testDss = &TestDss{dss: localfiles.MakeLocalFilesDssa()}
	case "sftp":
		testDss, err = tdo.getSftpTestDss()
	default:
		err = fmt.Errorf("kind for TestDssOptions %s unknown", tdo.Kind)
	}
	if err != nil {
		return nil, err
	}
	testDss.kind = tdo.Kind
	return testDss, nil
}

func (testDss *TestDss) Close() error {
	if testDss.rps != nil {
		plugin.Shutdown(testDss.rps)
		plugin.WaitFor(testDss.rps)
		if len(plugin.Errors(testDss.rps)) > 0 {
			return fmt.Errorf("plugin error %v", testDss.rps[0].Err)
		}
	}
	return nil
}
