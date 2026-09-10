package opelogimpl

import (
	"errors"
	"sync"

	"github.com/gammazero/deque"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

type memQ struct {
	mx       sync.Mutex
	entries  deque.Deque[string]
	closed   bool
	prodSubs []chan bool
}

func (mq *memQ) prodEvent() {
	if mq.entries.Len() == 0 {
		for _, prodSub := range mq.prodSubs {
			close(prodSub)
		}
		mq.prodSubs = make([]chan bool, 0)
	}
}

// Close implements [Queue].
func (mq *memQ) Close() error {
	mq.mx.Lock()
	defer mq.mx.Unlock()
	if mq.closed {
		return errors.New("memQ.Close: already done")
	}
	mq.closed = true
	mq.prodEvent()
	return nil
}

// Get implements [Queue].
func (mq *memQ) Get() (string, error) {
	mq.mx.Lock()
	for mq.entries.Len() == 0 {
		if mq.closed {
			mq.mx.Unlock()
			return "", common.ErrReadClosedQueue
		}
		prodSub := make(chan bool)
		mq.prodSubs = append(mq.prodSubs, prodSub)
		mq.mx.Unlock()
		<-prodSub
		mq.mx.Lock()
	}
	defer mq.mx.Unlock()
	return mq.entries.PopFront(), nil
}

// Put implements [Queue].
func (mq *memQ) Put(s string) error {
	mq.mx.Lock()
	defer mq.mx.Unlock()
	if mq.closed {
		return errors.New("memQ.Put: write on closed queue")
	}
	mq.prodEvent()
	mq.entries.PushBack(s)
	return nil
}

func NewMemQueue() opelog.Queue {
	return &memQ{prodSubs: []chan bool{}}
}
