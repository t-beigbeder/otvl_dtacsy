package opelogimpl

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

type OplWalker interface {
	Run() error
}

type oplWalkerImpl struct {
	mx         sync.Mutex
	lgr        *slog.Logger
	conc       int
	oplq       opelog.Queue
	oplm       opelog.OpeLogManager
	owo        *config.OpeLogOptionsType
	sds        dssa.Dssa
	tds        dssa.Dssa
	sRoot      string
	tRoot      string
	gErrs      []error
	syncTicker *time.Ticker
	bg context.Context
}

func (ow *oplWalkerImpl) owErr(lgr *slog.Logger, msg string, err error) error {
	lgr.Error(msg, "err", err)
	ow.mx.Lock()
	defer ow.mx.Unlock()
	ow.gErrs = append(ow.gErrs, fmt.Errorf("%s: %v", msg, err))
	return err
}

func (ow *oplWalkerImpl) detail(lgr *slog.Logger, msg string, args ...any) {
	lgr.Log(ow.bg, slog.LevelDebug+2, msg, args...)
}

func (ow *oplWalkerImpl) hasGoal(goal string) bool {
	for _, owg := range strings.Split(ow.owo.Goals, ",") {
		if owg == goal {
			return true
		}
	}
	return false
}

func (ow *oplWalkerImpl) oplmSync() {
	lgr := ow.lgr.With("worker", "oplmSync")
	lgr.Debug("oplWalkerImpl.oplmSync: start")

	for tick := range ow.syncTicker.C {
		lgr.Info("oplWalkerImpl.oplmSync: tick", "tick", tick)
		if err := ow.oplm.Sync(); err != nil {
			ow.owErr(lgr, "failed to synchronize logs", err)
		}
	}
}

func (ow *oplWalkerImpl) work(wkn int, wg *sync.WaitGroup) {
	defer wg.Done()
	lgr := ow.lgr.With("worker", wkn)
	lgr.Debug("oplWalkerImpl.work: start", "worker", wkn)
	for {
		relPath, err := ow.oplq.Get()
		if err != nil {
			if err != common.ErrReadClosedQueue {
				ow.owErr(lgr, "oplWalkerImpl.work", err)
			}
			break
		}
		ow.detail(ow.lgr, "oplWalkerImpl.work", "worker", wkn, "readPathFromQueue", relPath)
		le, err := ow.oplm.GetLogicalEntry(relPath)
		if err != nil {
			ow.owErr(lgr, "oplWalkerImpl.work: GetLogicalEntry", err)
			continue
		}
		ole := &oplLogicalEntry{plgr: lgr, relPath: relPath, owi: ow, le: le}
		if le == nil {
			ole.le = &opelog.LogicalEntry{}
			ole.hasChanges = true
		}
		if err := ole.process(); err != nil {
			ow.owErr(lgr, "oplWalkerImpl.work: process entry", err)
			continue
		}
		if ole.hasChanges {
			ow.oplm.PutLogicalEntry(relPath, ole.le)
		}
	}
	ow.lgr.Debug("oplWalkerImpl.work: stop", "worker", wkn)
}

func (ow *oplWalkerImpl) Run() error {
	ow.lgr.Info("oplWalkerImpl.Run: start")
	if err := ow.oplm.Open(false); err != nil {
		ow.lgr.Error("oplWalkerImpl.Run: open logs", "err", err)
		return err
	}
	var wg sync.WaitGroup
	for wkn := range ow.conc {
		wg.Add(1)
		go ow.work(wkn, &wg)
	}
	if ow.owo.SyncPeriod != 0 {
		ow.syncTicker = time.NewTicker(time.Duration(ow.owo.SyncPeriod))
		go ow.oplmSync()
	}
	ow.oplq.Put("")
	wg.Wait()
	if ow.owo.SyncPeriod != 0 {
		ow.syncTicker.Stop()
	}
	if err := ow.oplm.Close(); err != nil {
		ow.lgr.Error("oplWalkerImpl.Run: close logs", "err", err)
		return err
	}
	ow.lgr.Info("oplWalkerImpl.Run: end")
	if len(ow.gErrs) > 0 {
		err := fmt.Errorf("walker %d errors occured", len(ow.gErrs))
		ow.lgr.Error("oplWalkerImpl.Run:", "err", err)
		ow.detail(ow.lgr, "oplWalkerImpl.Run:", "err", err, "details", ow.gErrs)
		return err
	}
	return nil
}

func NewOplWalker(lgr *slog.Logger, conc int, oplq opelog.Queue, oplm opelog.OpeLogManager, owo *config.OpeLogOptionsType, sds, tds dssa.Dssa, sRoot, tRoot string) OplWalker {
	if conc == 0 {
		conc = 1
	}
	if oplq == nil {
		oplq = NewMemQueue()
	}
	return &oplWalkerImpl{
		lgr:  lgr,
		conc: conc, oplq: oplq, oplm: oplm, owo: owo,
		sds: sds, tds: tds, sRoot: sRoot, tRoot: tRoot,
		bg: context.Background(),
	}
}
