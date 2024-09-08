package transaction

import (
	context "context"
	sql "database/sql"
	"sync"

	"github.com/uptrace/bun"
)

type bunTxKey struct{}

type BunTransaction struct {
	tx   *bun.Tx
	ctx  context.Context
	once sync.Once
}

func (s *BunTransaction) Context() context.Context {
	return s.ctx
}

func (s *BunTransaction) Commit(ctx context.Context) error {
	s.clearTx()

	return s.tx.Commit()
}

func (s *BunTransaction) Rollback(ctx context.Context) error {
	s.clearTx()

	return s.tx.Rollback()
}

func (s *BunTransaction) clearTx() {
	s.once.Do(func() { s.ctx = ClearTx(s.ctx) })
}

type BunProvider struct {
	db *bun.DB
}

func NewBunProvider(db *bun.DB) *BunProvider {
	return &BunProvider{
		db: db,
	}
}

func (s *BunProvider) Begin(ctx context.Context) (Transaction, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}

	return &BunTransaction{tx: &tx, ctx: s.transactionContext(ctx, &tx)}, nil
}

func (s *BunProvider) BeginTx(ctx context.Context, opts sql.TxOptions) (Transaction, error) {
	tx, err := s.db.BeginTx(ctx, &opts)
	if err != nil {
		return nil, err
	}

	return &BunTransaction{tx: &tx, ctx: s.transactionContext(ctx, &tx)}, nil
}

func (s *BunProvider) transactionContext(ctx context.Context, tx *bun.Tx) context.Context {
	return context.WithValue(StartTx(ctx), bunTxKey{}, tx)
}

func (s *BunProvider) Executor(ctx context.Context) bun.IDB {
	executor, _ := s.executor(ctx)

	return executor
}

func (s *BunProvider) TxEnabled(ctx context.Context) bool {
	_, ok := s.executor(ctx)

	return ok
}

func (s *BunProvider) executor(ctx context.Context) (bun.IDB, bool) {
	if !TxEnabled(ctx) {
		return s.db, false
	}

	tx, ok := ctx.Value(bunTxKey{}).(*bun.Tx)
	if !ok {
		return s.db, false
	}

	return tx, true
}
