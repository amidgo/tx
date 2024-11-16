package buntransaction

import (
	context "context"
	sql "database/sql"
	"sync"

	"github.com/amidgo/transaction"
	"github.com/uptrace/bun"
)

type txKey struct{}

type Transaction struct {
	tx   *bun.Tx
	ctx  context.Context
	once sync.Once
}

func (s *Transaction) Context() context.Context {
	return s.ctx
}

func (s *Transaction) Commit(ctx context.Context) error {
	s.clearTx()

	return s.tx.Commit()
}

func (s *Transaction) Rollback(ctx context.Context) error {
	s.clearTx()

	return s.tx.Rollback()
}

func (s *Transaction) clearTx() {
	s.once.Do(func() { s.ctx = transaction.ClearTx(s.ctx) })
}

type Provider struct {
	db *bun.DB
}

func NewProvider(db *bun.DB) *Provider {
	return &Provider{
		db: db,
	}
}

func (s *Provider) Begin(ctx context.Context) (transaction.Transaction, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}

	return &Transaction{tx: &tx, ctx: s.transactionContext(ctx, &tx)}, nil
}

func (s *Provider) BeginTx(ctx context.Context, opts sql.TxOptions) (transaction.Transaction, error) {
	tx, err := s.db.BeginTx(ctx, &opts)
	if err != nil {
		return nil, err
	}

	return &Transaction{tx: &tx, ctx: s.transactionContext(ctx, &tx)}, nil
}

func (s *Provider) transactionContext(ctx context.Context, tx *bun.Tx) context.Context {
	return context.WithValue(transaction.StartTx(ctx), txKey{}, tx)
}

func (s *Provider) Executor(ctx context.Context) bun.IDB {
	executor, _ := s.executor(ctx)

	return executor
}

func (s *Provider) TxEnabled(ctx context.Context) bool {
	_, ok := s.executor(ctx)

	return ok
}

func (s *Provider) executor(ctx context.Context) (bun.IDB, bool) {
	if !transaction.TxEnabled(ctx) {
		return s.db, false
	}

	tx, ok := ctx.Value(txKey{}).(*bun.Tx)
	if !ok {
		return s.db, false
	}

	return tx, true
}
