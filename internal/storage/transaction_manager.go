package storage

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type contextKey string

const txKey contextKey = "transaction"

type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type TransactionManager struct {
	pool *pgxpool.Pool
}

func NewTransactionManager(pool *pgxpool.Pool) *TransactionManager {
	return &TransactionManager{
		pool: pool,
	}
}

func (tm *TransactionManager) Do(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}()

	ctx = context.WithValue(ctx, txKey, tx)
	return fn(ctx)
}
