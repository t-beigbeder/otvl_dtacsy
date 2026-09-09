package dssaimpl

import (
	"errors"
	"fmt"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
)

type TestDss struct {
	kind string
	dss  dssa.Dssa
}

type TestDssOptions struct {
	Kind          string // "lf", "sftp"
	SftpUser      string
	SftpHost      string
	SftpPort      string
	SftpIdent     string
	SftpRoot      string
	SftpKHFile    string
	SftpHasServer bool
}

func NewTestDss(tdo *TestDssOptions) (*TestDss, error) {
	var (
		err     error
		testDss *TestDss
	)
	switch tdo.Kind {
	case "lf":
		testDss = &TestDss{dss: localfiles.MakeLocalFilesDssa()}
	case "sftp":
		err = errors.ErrUnsupported
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
	switch testDss.kind {
	case "lf":
		return nil
	default:
		return fmt.Errorf("TestDss.Close: kind %s unknown", testDss.kind)
	}
}
