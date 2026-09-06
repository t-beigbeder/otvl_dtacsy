package opelogimpl

import (
	"errors"
	"log/slog"
	"path"
	"strings"

	"github.com/t-beigbeder/vdasync/opelog"
)

type oplLogicalEntry struct {
	relPath string
	plgr    *slog.Logger
	owi     *oplWalkerImpl
	le      *opelog.LogicalEntry
}

func (ole *oplLogicalEntry) lgr() *slog.Logger {
	return ole.plgr.With("relPath", ole.relPath)
}

func (ole *oplLogicalEntry) load() error {
	ole.lgr().Debug("load: start")

	ole.lgr().Debug("load: stop")
	return errors.ErrUnsupported
}

func (ole *oplLogicalEntry) process() error {
	ole.lgr().Debug("process: start")
	var (
		err error
	)
	for _, goal := range strings.Split("load,create,update,verify", ",") {
		if !ole.owi.hasGoal(goal) {
			continue
		}
		switch goal {
		case "load":
			err = ole.load()
		default:
			err = errors.ErrUnsupported
		}
		if err != nil {
			break
		}
	}
	ole.lgr().Debug("process: stop")
	return err
}

type oplStoredEntry struct {
	oplLogicalEntry
	isTarget bool
}

func (ose *oplStoredEntry) pfx() string {
	if ose.isTarget {
		return "{T}"
	} else {
		return "{S}"
	}
}

func (ose *oplStoredEntry) lgr() *slog.Logger {
	return ose.plgr.With("path", ose.pfx()+ose.relPath)
}

func (ose *oplStoredEntry) root() string {
	if ose.isTarget {
		return ose.owi.tRoot
	} else {
		return ose.owi.sRoot
	}
}

func (ose *oplStoredEntry) fullPath() string {
	return path.Join(ose.root(), ose.relPath)
}
