package opelogimpl

import (
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

type memQ struct {
	cq chan []byte
}

// Close implements [Queue].
func (mq *memQ) Close() error {
	close(mq.cq)
	return nil
}

// Get implements [Queue].
func (mq *memQ) Get() (string, error) {
	got := <-mq.cq
	if got == nil {
		return "", common.ErrReadClosedQueue
	}
	return string(got), nil
}

// Put implements [Queue].
func (mq *memQ) Put(s string) error {
	mq.cq <- []byte(s)
	return nil
}

func NewMemQueue(conc int) opelog.Queue {
	mq := &memQ{
		cq: make(chan []byte, conc+1),
	}
	return mq
}
