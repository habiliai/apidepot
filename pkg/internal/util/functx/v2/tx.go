package functx

import (
	"context"
	"sync"
)

type (
	RollbackFunc func(context.Context)
	DoneFunc     func(ctx context.Context, rollback bool)
	Tx           interface {
		Rollback(ctx context.Context)
		Add(tx Tx)
		Done()
	}
	contextKey string
)

const (
	contextKeyFuncTx contextKey = "functx"
)

func getFuncTx(ctx context.Context) Tx {
	tx, ok := ctx.Value(contextKeyFuncTx).(Tx)
	if !ok {
		return nil
	}
	return tx
}

func WithFuncTx(ctx context.Context) (context.Context, DoneFunc) {
	parentTx := getFuncTx(ctx)
	tx := &empty{
		parent: parentTx,
	}
	if parentTx != nil {
		parentTx.Add(tx)
	}

	once := &sync.Once{}
	return context.WithValue(ctx, contextKeyFuncTx, tx), func(ctx context.Context, rollback bool) {
		once.Do(func() {
			if parentTx != nil {
				return
			}
			if rollback {
				tx.Rollback(ctx)
			} else {
				tx.Done()
			}
		})
	}
}

func AddRollback(ctx context.Context, f RollbackFunc) {
	parent := getFuncTx(ctx)
	if parent == nil {
		return
	}

	tx := &funcTx{
		f:      f,
		parent: parent,
	}

	parent.Add(tx)
}
