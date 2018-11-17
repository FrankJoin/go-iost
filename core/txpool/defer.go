package txpool

import (
	"bytes"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/iost-official/go-iost/core/tx"

	"github.com/emirpasic/gods/trees/redblacktree"
	"github.com/uber-go/atomic"
)

var (
	minTickerTime = time.Second
)

func compareDeferTx(a, b interface{}) int {
	txa := a.(*tx.Tx)
	txb := b.(*tx.Tx)
	if txa.Time == txb.Time {
		return bytes.Compare(txa.Hash(), txb.Hash())
	}
	return int(txa.Time - txb.Time)
}

// DeferServer manages defer transaction and sends them to txpool on time.
type DeferServer struct {
	pool             *redblacktree.Tree
	rw               *sync.RWMutex
	nextScheduleTime atomic.Int64

	txpool *TxPImpl

	quitCh chan struct{}
}

// NewDeferServer returns a new DeferServer instance.
func NewDeferServer(txpool *TxPImpl) (*DeferServer, error) {
	deferServer := &DeferServer{
		pool:   redblacktree.NewWith(compareDeferTx),
		rw:     new(sync.RWMutex),
		txpool: txpool,
		quitCh: make(chan struct{}),
	}
	err := deferServer.buildIndex()
	if err != nil {
		return nil, fmt.Errorf("build defertx index error, %v", err)
	}

	return deferServer, nil
}

func (d *DeferServer) buildIndex() error {
	txs, err := d.txpool.global.BlockChain().AllDelaytx()
	if err != nil {
		return err
	}
	for _, t := range txs {
		d.pool.Put(d.toIndex(t), true)
	}
	return nil
}

func (d *DeferServer) toIndex(delayTx *tx.Tx) *tx.Tx {
	return &tx.Tx{
		ReferredTx: delayTx.Hash(),
		Time:       delayTx.Time + delayTx.Delay,
	}
}

// DelDeferTx deletes a tx in defer server.
func (d *DeferServer) DelDeferTx(deferTx *tx.Tx) error {
	idx := &tx.Tx{
		ReferredTx: deferTx.ReferredTx,
		Time:       deferTx.Time,
	}
	d.rw.Lock()
	d.pool.Remove(idx)
	d.rw.Unlock()
	return nil
}

// StoreDeferTx stores a tx in defer server.
func (d *DeferServer) StoreDeferTx(delayTx *tx.Tx) {
	idx := d.toIndex(delayTx)
	d.rw.Lock()
	d.pool.Put(idx, true)
	d.rw.Unlock()
	if idx.Time < d.nextScheduleTime.Load() {
		d.nextScheduleTime.Store(idx.Time)
		d.restartDeferTicker()
	}
}

// DumpDeferTx dumps all defer transactions for debug.
func (d *DeferServer) DumpDeferTx() []*tx.Tx {
	ret := make([]*tx.Tx, 0)
	iter := d.pool.Iterator()
	d.rw.RLock()
	ok := iter.Next()
	for ok {
		deferTx := iter.Key().(*tx.Tx)
		ret = append(ret, deferTx)
		ok = iter.Next()
	}
	d.rw.RUnlock()
	return ret
}

// Start starts the defer server.
func (d *DeferServer) Start() error {
	go d.deferTicker()
	return nil
}

// Stop stops the defer server.
func (d *DeferServer) Stop() {
	d.stopDeferTicker()
}

func (d *DeferServer) stopDeferTicker() {
	d.quitCh <- struct{}{}
	<-d.quitCh
}

func (d *DeferServer) restartDeferTicker() {
	d.stopDeferTicker()
	go d.deferTicker()
}

func (d *DeferServer) deferTicker() {
	for {
		scheduled := time.Duration(d.nextScheduleTime.Load() - time.Now().UnixNano())
		if scheduled < minTickerTime {
			scheduled = minTickerTime
		}
		select {
		case <-d.quitCh:
			d.quitCh <- struct{}{}
			return
		case <-time.After(scheduled):
			iter := d.pool.Iterator()
			d.rw.RLock()
			ok := iter.Next()
			d.rw.RUnlock()
			for ok {
				deferTx := iter.Key().(*tx.Tx)
				if deferTx.Time > time.Now().UnixNano() {
					d.nextScheduleTime.Store(deferTx.Time)
					break
				}
				err := d.txpool.AddDefertx(deferTx.ReferredTx)
				if err == ErrCacheFull {
					d.nextScheduleTime.Store(deferTx.Time)
					break
				}
				if err == nil || err == ErrDupChainTx || err == ErrDupPendingTx {
					d.rw.Lock()
					d.pool.Remove(deferTx)
					d.rw.Unlock()
				}
				d.rw.RLock()
				ok = iter.Next()
				d.rw.RUnlock()
			}
			if !ok {
				d.nextScheduleTime.Store(math.MaxInt64)
			}
		}
	}
}
