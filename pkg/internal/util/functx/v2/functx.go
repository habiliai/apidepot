package functx

import (
	"context"
	"sync"
)

type funcTx struct {
	children []Tx
	f        RollbackFunc
	parent   Tx

	lck  sync.Mutex
	done bool
}

func (t *funcTx) Add(tx Tx) {
	t.children = append(t.children, tx)
}

func (t *funcTx) Rollback(ctx context.Context) {
	if t.done {
		return
	}

	t.lck.Lock()
	defer t.lck.Unlock()

	if t.done {
		return
	}

	defer func() {
		t.done = true
	}()

	t.f(ctx)

	for i := len(t.children) - 1; i >= 0; i-- {
		t.children[i].Rollback(ctx)
	}
}

func (t *funcTx) Done() {
	defer func() {
		t.done = true
	}()
	for _, child := range t.children {
		child.Done()
	}
}
