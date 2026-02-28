package db

import (
	"context"
	"database/sql"
)

// DBTX is satisfied by both *sql.DB and *sql.Tx, allowing repositories
// to work transparently inside or outside a transaction.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// TxManager wraps *sql.DB to provide transactional execution.
type TxManager struct {
	db *sql.DB
}

func NewTxManager(db *sql.DB) *TxManager {
	return &TxManager{db: db}
}

// RunInTx executes fn inside a transaction. If fn returns an error the
// transaction is rolled back; otherwise it is committed.
func (m *TxManager) RunInTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
