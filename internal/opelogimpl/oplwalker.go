package opelogimpl

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

type OplWalker interface {
	Run() error
}

type oplWalkerImpl struct {
	mx    sync.Mutex
	lgr   *slog.Logger
	conc  int
	oplq  opelog.Queue
	oplm  opelog.OpeLogManager
	sds   dssa.Dssa
	tds   dssa.Dssa
	gErrs []error
}

func (ow *oplWalkerImpl) owErr(msg string, err error) error {
	ow.lgr.Error(msg, "err", err)
	ow.mx.Lock()
	defer ow.mx.Unlock()
	ow.gErrs = append(ow.gErrs, fmt.Errorf("%s: %v", msg, err))
	return err
}

func (ow *oplWalkerImpl) work(wkn int, wg *sync.WaitGroup) {
	defer wg.Done()
	ow.lgr.Info("oplWalkerImpl.work: start", "worker", wkn)
	for {
		rp, err := ow.oplq.Get()
		if err != nil {
			if err != common.ErrReadClosedQueue {
				ow.owErr("oplWalkerImpl.work", err)
			}
			break
		}
		ow.lgr.Debug("work", "worker", wkn, "received rp", rp)
		if err := ow.oplq.Close(); err != nil {
			ow.owErr("oplWalkerImpl.work: close queue", err)
		}
	}
	ow.lgr.Info("oplWalkerImpl.work: stop", "worker", wkn)
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
	ow.oplq.Put("")
	wg.Wait()
	if err := ow.oplm.Close(); err != nil {
		ow.lgr.Error("oplWalkerImpl.Run: close logs", "err", err)
		return err
	}
	ow.lgr.Info("oplWalkerImpl.Run: end")
	if len(ow.gErrs) > 0 {
		err := fmt.Errorf("walker %d errors occured", len(ow.gErrs))
		ow.lgr.Error("oplWalkerImpl.Run:", "err", err)
		ow.lgr.Debug("oplWalkerImpl.Run:", "err", err, "details", ow.gErrs)
		return err
	}
	return nil
}

func NewOplWalker(lgr *slog.Logger, conc int, oplq opelog.Queue, oplm opelog.OpeLogManager, sds, tds dssa.Dssa) OplWalker {
	if conc == 0 {
		conc = 1
	}
	if oplq == nil {
		oplq = NewMemQueue(conc)
	}
	return &oplWalkerImpl{
		lgr:  lgr,
		conc: conc, oplq: oplq, oplm: oplm,
		sds: sds, tds: tds,
	}
}
