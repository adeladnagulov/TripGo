package txManager

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type contextKey string

const TxKey contextKey = "transaction"

type TransactionManager struct {
	pool      *pgxpool.Pool
	opTimeout time.Duration
}

func NewTransactionManager(pool *pgxpool.Pool, opTimeout time.Duration) *TransactionManager {
	return &TransactionManager{
		pool:      pool,
		opTimeout: opTimeout,
	}
}

func (tm *TransactionManager) Do(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	if _, ok := ctx.Value(TxKey).(pgx.Tx); ok {
		return fn(ctx)
	}

	beginCtx, beginCancel := context.WithTimeout(ctx, tm.opTimeout)
	tx, err := tm.pool.Begin(beginCtx)
	beginCancel()

	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			rbCtx, rbCancel := context.WithTimeout(context.WithoutCancel(ctx), tm.opTimeout)
			tx.Rollback(rbCtx)
			rbCancel()
			panic(p)
		} else if err != nil {
			rbCtx, rbCancel := context.WithTimeout(context.WithoutCancel(ctx), tm.opTimeout)
			if rbErr := tx.Rollback(rbCtx); rbErr != nil {
				err = fmt.Errorf("operation failed: %w, rollback failed: %v", err, rbErr)
			}
			rbCancel()
		} else {
			cmCtx, cmCancel := context.WithTimeout(ctx, tm.opTimeout)
			err = tx.Commit(cmCtx)
			cmCancel()
		}
	}()

	ctx = context.WithValue(ctx, TxKey, tx)
	return fn(ctx)
}
