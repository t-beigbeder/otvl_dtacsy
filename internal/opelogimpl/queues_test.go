package opelogimpl

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

type productionDesc struct {
	initDelay time.Duration
	smpNum    int
	smpGen    func(int) string
	endDelay  time.Duration
}

type baseTestDesc struct {
	label         string
	unSkipped     bool
	lgr           *slog.Logger
	conc          int
	qType         string
	segSize       int
	production    []*productionDesc
	consInitDelay time.Duration
	endFlag       string
}

func smpGen(sn int, tp *productionDesc) string {
	if tp.smpGen != nil {
		return tp.smpGen(sn)
	}
	return fmt.Sprintf("%d", sn)
}

func runBaseTest(_ *testing.T, bt *baseTestDesc, tq opelog.Queue) error {
	var wg sync.WaitGroup
	var total int
	hasErrors := false
	bt.lgr = bt.lgr.With("label", bt.label)
	wg.Add(1)
	go func() {
		var err error
		defer wg.Done()
		bt.lgr.Debug("runBaseTest.producer started")
	DONE:
		for _, tp := range bt.production {
			if tp.initDelay != 0 {
				time.Sleep(tp.initDelay)
			}
			for sn := range tp.smpNum {
				if err = tq.Put(smpGen(sn, tp)); err != nil {
					break DONE
				}
			}
			total += tp.smpNum
			if tp.endDelay != 0 {
				time.Sleep(tp.endDelay)
			}
		}
		if err == nil {
			if err = tq.Put(bt.endFlag); err == nil {
				total++
			}
		}
		if err != nil {
			bt.lgr.Error("runBaseTest.producer ended", "err", err)
			hasErrors = true
		} else {
			bt.lgr.Debug("runBaseTest.producer ended")
		}
	}()
	if bt.consInitDelay != 0 {
		time.Sleep(bt.consInitDelay)
	}
	counts := make([]int, bt.conc)
	for cons := range bt.conc {
		wg.Add(1)
		go func(cons int) {
			var err error
			var s string
			defer wg.Done()
			bt.lgr.Debug("runBaseTest.consumer started", "cons", cons)
			for {
				s, err = tq.Get()
				if err != nil {
					if err == common.ErrReadClosedQueue {
						err = nil
					}
					break
				}
				counts[cons]++
				if s == bt.endFlag {
					err = tq.Close()
					break
				}
			}
			if err != nil {
				bt.lgr.Error("runBaseTest.consumer ended", "cons", cons, "err", err)
				hasErrors = true
			} else {
				bt.lgr.Debug("runBaseTest.consumer ended", "cons", cons)
			}
		}(cons)
	}

	wg.Wait()
	var err error
	if hasErrors {
		err = fmt.Errorf("[%s] subroutine error, consult logs for details", bt.label)
	} else {
		sum := 0
		for cons := range bt.conc {
			sum += counts[cons]
		}
		if sum != total {
			err = fmt.Errorf("[%s] produced %d consumed %d", bt.label, total, sum)
		}
	}
	if err != nil {
		bt.lgr.Error("runBaseTest failed", "err", err, "counts", counts)
		return err
	}
	bt.lgr.Debug("runBaseTest", "counts", counts)
	return nil
}

func TestQueuesSimple(t *testing.T) {
	dbgLog := common.DbgLogger()
	nullLog := common.GetNullLogger()
	defLog := common.GetLogger()
	_, _ = dbgLog, defLog
	skipped := false
	tests := []*baseTestDesc{
		{
			label: "memq prod then cons",
			production: []*productionDesc{
				{
					smpNum: 1000000,
				},
			},
		},
		{
			label: "memq prod then cons 4",
			conc:  4,
			production: []*productionDesc{
				{
					smpNum: 1000000,
				},
			},
		},
		{
			label:   "largeq prod then cons",
			qType:   "LargeQueue",
			segSize: 10000,
			production: []*productionDesc{
				{
					smpNum: 1000000,
				},
			},
		},
		{
			label:   "largeq prod then cons 4",
			conc:    4,
			qType:   "LargeQueue",
			segSize: 10000,
			production: []*productionDesc{
				{
					smpNum: 1000000,
				},
			},
		},
	}
	for _, test := range tests {
		var tq opelog.Queue
		var err error
		td := t.TempDir()
		if test.lgr == nil {
			test.lgr = nullLog
		}
		if test.conc <= 0 {
			test.conc = 1
		}
		if test.endFlag == "" {
			test.endFlag = "EOF"
		}
		switch test.qType {
		default:
			tq = NewMemQueue(test.conc)
		case "MemQueue":
			tq = NewMemQueue(test.conc)
		case "LargeQueue":
			tq, err = NewLargeQ(test.lgr, td, test.segSize)
		}
		require.NoError(t, err)
		if skipped && !test.unSkipped {
			defLog.Info(t.Name(), "label", test.label, "skipped", !test.unSkipped)
			continue
		}
		require.NoError(t, runBaseTest(t, test, tq))
	}
}
