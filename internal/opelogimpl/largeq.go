package opelogimpl

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/gammazero/deque"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

type largeQ struct {
	lgr      *slog.Logger
	dir      string
	segSize  int
	mx       sync.Mutex
	closed   bool
	prodSubs []chan bool
	cOff     int
	pOff     int
	cEntries *deque.Deque[string]
	pEntries *deque.Deque[string]
}

func (lq *largeQ) saveEntries() error {
	fp := path.Join(lq.dir, fmt.Sprintf(".largeQ-%d.txt", lq.pOff/lq.segSize-1))
	out := make([]string, 0, lq.pEntries.Len())
	out = lq.pEntries.AppendToSlice(out)
	if err := common.WriteFile(fp, []byte(strings.Join(out, "\n"))); err != nil {
		return fmt.Errorf("largeQ.saveEntries: write %s error %v", fp, err)
	}
	lq.pEntries.Clear()
	return nil
}

func (lq *largeQ) prodEvent() {
	if lq.cOff == lq.pOff {
		for _, prodSub := range lq.prodSubs {
			close(prodSub)
		}
		lq.prodSubs = make([]chan bool, 0)
	}
}

// Close implements [opelog.Queue].
func (lq *largeQ) Close() error {
	lq.mx.Lock()
	defer lq.mx.Unlock()
	if lq.closed {
		return errors.New("largeQ.Close: already done")
	}
	lq.closed = true
	if lq.pEntries.Len() > 0 {
		if err := lq.saveEntries(); err != nil {
			return err
		}
	}
	lq.prodEvent()
	return nil
}

// Get implements [opelog.Queue].
func (lq *largeQ) Get() (string, error) {
	lq.mx.Lock()
	for lq.cOff == lq.pOff {
		if lq.closed {
			lq.mx.Unlock()
			return "", common.ErrReadClosedQueue
		}
		prodSub := make(chan bool)
		lq.prodSubs = append(lq.prodSubs, prodSub)
		lq.mx.Unlock()
		<-prodSub // FIXME: can block
		lq.mx.Lock()
	}
	defer lq.mx.Unlock()

	if lq.cOff/lq.segSize == lq.pOff/lq.segSize {
		s := lq.pEntries.At(lq.cOff % lq.segSize)
		lq.cOff++
		return s, nil
	}
	if lq.cEntries.Len() == 0 {
		fp := path.Join(lq.dir, fmt.Sprintf(".largeQ-%d.txt", lq.cOff/lq.segSize))
		bs, err := common.UnsafeLoadFile(fp)
		if err != nil {
			return "", fmt.Errorf("largeQ.Get: load error %v pOff %d cOff %d", err, lq.pOff, lq.cOff)
		}
		if err := os.Remove(fp); err != nil {
			return "", fmt.Errorf("largeQ.Get: remove error %v", err)
		}
		lns := strings.Split(string(bs), "\n")
		if len(lns) > lq.segSize {
			return "", fmt.Errorf("largeQ.Get: read %s len %d", fp, len(lns))
		}
		lq.cEntries.CopyInSlice(lns)
	}
	s := lq.cEntries.At(lq.cOff % lq.segSize)
	lq.cOff++
	if lq.cOff%lq.segSize == 0 {
		lq.cEntries.Clear()
	}
	return s, nil
}

// Put implements [opelog.Queue].
func (lq *largeQ) Put(s string) error {
	if strings.Contains(s, "\n") {
		return errors.New("largeQ.Put optimization forbids \\n character")
	}
	lq.mx.Lock()
	defer lq.mx.Unlock()
	if lq.closed {
		return errors.New("largeQ.Put: write on closed queue")
	}
	lq.prodEvent()
	lq.pEntries.PushBack(s)
	lq.pOff++
	if lq.pEntries.Len() == lq.segSize {
		if err := lq.saveEntries(); err != nil {
			return err
		}
	}
	return nil
}

func NewLargeQ(lgr *slog.Logger, dir string, segSize int) (opelog.Queue, error) {
	var pEntries deque.Deque[string]
	pEntries.SetBaseCap(segSize)
	var cEntries deque.Deque[string]
	cEntries.SetBaseCap(segSize)
	return &largeQ{
		lgr: lgr, dir: dir, segSize: segSize,
		prodSubs: []chan bool{},
		pEntries: &pEntries, cEntries: &cEntries,
	}, nil
}
