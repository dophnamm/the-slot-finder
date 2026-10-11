package db

import (
	"context"

	"gorm.io/gorm"
)

// txKey stores the open transaction in a context.
type txKey struct{}

// TxManager runs work in one database transaction. Services own the boundary; repositories
// join it by reading their handle with PostgresDB.From(ctx).
type TxManager interface {
	// Do begins a transaction, stores it in ctx and runs fn. It commits when fn returns nil
	// and rolls back on an error or a panic. A nested Do reuses the open transaction.
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type txManager struct {
	db *PostgresDB
}

// ProvideTxManager creates the transaction manager.
func ProvideTxManager(db *PostgresDB) TxManager {
	return &txManager{db: db}
}

func (m *txManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return fn(ctx)
	}
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}

// From returns the transaction open in ctx, or the plain database, bound to ctx.
// Repositories call it instead of WithContext so they work inside and outside TxManager.Do.
func (r *PostgresDB) From(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return r.WithContext(ctx)
}
