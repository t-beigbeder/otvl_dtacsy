package opelogimpl

import (
	"errors"
	"log/slog"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

type oplLogicalEntry struct {
	hasChanges bool
	relPath    string
	plgr       *slog.Logger
	owi        *oplWalkerImpl
	le         *opelog.LogicalEntry
	sChildrenQ []string
	tChildrenQ []string
	sParentQ   bool
	tParentQ   bool
}

func (ole *oplLogicalEntry) lgr() *slog.Logger {
	return ole.plgr.With("relPath", ole.relPath)
}

func (ole *oplLogicalEntry) source() *oplStoredEntry {
	return &oplStoredEntry{oplLogicalEntry: ole}
}

func (ole *oplLogicalEntry) target() *oplStoredEntry {
	return &oplStoredEntry{oplLogicalEntry: ole, isTarget: true}
}

func (ole *oplLogicalEntry) load() error {
	ole.lgr().Debug("load: start")
	if err := ole.source().load(); err != nil {
		return err
	}
	if err := ole.target().load(); err != nil {
		return err
	}
	return nil
}

func (ole *oplLogicalEntry) queueChildren() error {
	merged := slices.Clone(ole.sChildrenQ)
	for _, childRp := range ole.tChildrenQ {
		if !slices.Contains(ole.sChildrenQ, childRp) {
			merged = append(merged, childRp)
		}
	}
	ole.sChildrenQ, ole.tChildrenQ = nil, nil
	for _, childRp := range merged {
		if err := ole.owi.oplq.Put(path.Join(ole.relPath, childRp)); err != nil {
			ole.owi.owErr(ole.lgr(), "oplq.Put error", err)
			return err
		}
	}
	return nil
}

func (ole *oplLogicalEntry) queueParent() error {
	var err error
	if ole.sParentQ || ole.tParentQ {
		err = ole.owi.oplq.Put(common.ParentPath(ole.relPath))
		ole.sParentQ, ole.tParentQ = false, false
	}
	if err != nil {
		ole.owi.owErr(ole.lgr(), "oplq.Put error", err)

	}
	return err
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
		if err = ole.queueChildren(); err != nil {
			break
		}
		if err = ole.queueParent(); err != nil {
			break
		}
	}
	ole.lgr().Debug("process: stop")
	return err
}

type oplStoredEntry struct {
	*oplLogicalEntry
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
	return ose.plgr.With("path", path.Join(ose.pfx(), ose.relPath))
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

func (ose *oplStoredEntry) events() (evs *[]*opelog.Event) {
	if ose.isTarget {
		evs = &ose.le.TargetEvents
	} else {
		evs = &ose.le.SourceEvents
	}
	return
}

func (ose *oplStoredEntry) existenceEv() (ev *opelog.Event) {
	evs := ose.events()
	for i := len(*evs) - 1; i >= 0; i-- {
		if (*evs)[i].Kind == opelog.EVT_ABS || (*evs)[i].Kind == opelog.EVT_EXIST {
			return (*evs)[i]
		}
	}
	return nil
}

func (ose *oplStoredEntry) states() (sts *[]*opelog.StoredEntry) {
	if ose.isTarget {
		sts = &ose.le.TargetStates
	} else {
		sts = &ose.le.SourceStates
	}
	return
}

func (ose *oplStoredEntry) currentState() *opelog.StoredEntry {
	sts := ose.states()
	if len(*sts) == 0 {
		return nil
	}
	return (*sts)[len(*sts)-1]
}

func (ose *oplStoredEntry) dss() (ds dssa.Dssa) {
	if ose.isTarget {
		ds = ose.owi.tds
	} else {
		ds = ose.owi.sds
	}
	return
}

func (ose *oplStoredEntry) newState(se *opelog.StoredEntry) {
	sts := ose.states()
	if len(*sts) == 0 {
		*sts = append(*sts, se)
		return
	}
	if se.Equal((*sts)[len(*sts)-1]) {
		return
	}
	*sts = append(*sts, se)
}

func (ose *oplStoredEntry) newEvent(kind opelog.EventCode, origin opelog.OriginCode, sErr string) {
	evs := ose.events()
	*evs = append(*evs,
		&opelog.Event{
			Kind: kind, Origin: origin, TimeStamp: time.Now().Unix(),
			StateIndex: int32(len(*ose.states()) - 1), Error: sErr})
	ose.hasChanges = true
}

func (ose *oplStoredEntry) queueChild(child string) error {
	if err := ose.owi.oplq.Put(path.Join(ose.relPath, child)); err != nil {
		ose.owi.owErr(ose.lgr(), "oplq.Put error", err)
		return err
	}
	return nil
}

func (ose *oplStoredEntry) setChildrenQ(children []string) {
	if ose.isTarget {
		ose.tChildrenQ = slices.Clone(children)
	} else {
		ose.sChildrenQ = slices.Clone(children)
	}
}

func (ose *oplStoredEntry) setParentQ() {
	if ose.isTarget {
		ose.tParentQ = true
	} else {
		ose.sParentQ = true
	}
}

func (ose *oplStoredEntry) load() error {
	eev := ose.existenceEv()
	if eev != nil && eev.Error == "" {
		return nil
	}
	ose.lgr().Debug("load: start")
	de, err := ose.dss().Stat(ose.fullPath())
	if err != nil && !de.ErrNotExist {
		ose.newState(&opelog.StoredEntry{})
		ose.newEvent(opelog.EVT_ABS, opelog.ORI_STAT, err.Error())
		return err
	}
	if de.ErrNotExist {
		ose.newState(&opelog.StoredEntry{})
		ose.newEvent(opelog.EVT_ABS, opelog.ORI_STAT, "")
		return nil
	}
	var children, fCn []string
	var cdes []*dssa.DataEntry
	if de.IsDir {
		cdes, err = ose.dss().List(ose.fullPath())
		if err != nil {
			se := opelog.FromDataEntry(de, nil)
			se.IsPresent = true
			ose.newState(se)
			ose.newEvent(opelog.EVT_EXIST, opelog.ORI_STAT, err.Error())
			return err
		}

		for _, cde := range cdes {
			if cde.IsDir {
				children = append(children, path.Base(cde.Path))
			} else {
				fCn = append(fCn, path.Base(cde.Path))
			}
			children = slices.Concat(children, fCn)
			ose.setChildrenQ(children)
		}
	} else {
		ose.setParentQ()
	}
	se := opelog.FromDataEntry(de, children)
	se.IsPresent = true
	ose.newState(se)
	ose.newEvent(opelog.EVT_EXIST, opelog.ORI_STAT, "")

	return nil
}
