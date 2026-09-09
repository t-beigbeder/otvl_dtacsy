package opelogimpl

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
)

func TestMemqSimple(t *testing.T) {
	const conc = 4
	lgr := common.DbgLogger()
	mq := NewMemQueue(conc)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		lgr.Debug("push started")
		for i := range 100000 {
			require.NoError(t, mq.Put(fmt.Sprintf("%7d", i)))
		}
		require.NoError(t, mq.Put("EOF"))
		lgr.Debug("push done")
		wg.Done()
	}()
	subTotals := [conc]int{}
	for cons := range conc {
		wg.Add(1)
		go func(cons int) {
			lgr.Debug("pull started", "cons", cons)
		DONE:
			for {
				si, err := mq.Get()
				if err != nil {
					lgr.Error("pull", "err", err)
					break DONE
				}
				if si == "EOF" {
					if err := mq.Close(); err != nil {
						lgr.Error("pull", "err", err)
					}
					break DONE
				} else {
					subTotals[cons]++
				}
			}
			lgr.Debug("pull done", "cons", cons)
			wg.Done()
		}(cons)
	}
	wg.Wait()
	total := 0
	for cons := range conc {
		total += subTotals[cons]
	}
	require.Equal(t, 100000, total)
}
