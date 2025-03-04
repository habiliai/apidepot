package functx

import "context"

type empty struct {
	children []Tx
	parent   Tx
}

func (e *empty) Add(tx Tx) {
	e.children = append(e.children, tx)
}

func (e *empty) Rollback(ctx context.Context) {
	for i := len(e.children) - 1; i >= 0; i-- {
		e.children[i].Rollback(ctx)
	}
}

func (e *empty) Done() {
	for _, child := range e.children {
		child.Done()
	}
}
