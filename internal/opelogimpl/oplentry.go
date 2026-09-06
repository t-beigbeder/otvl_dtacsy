package opelogimpl

import (
	"path"

	"github.com/t-beigbeder/vdasync/opelog"
)

type oplLogicalEntry struct {
	relPath string
	owi     *oplWalkerImpl
	le      *opelog.LogicalEntry
}

type oplStoredEntry struct {
	oplLogicalEntry
	isTarget bool
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
