package transactionmocks

import (
	"context"
	"database/sql"
	"sync/atomic"

	"github.com/amidgo/transaction"
)

var _ transaction.Provider = (*Provider)(nil)

type providerAsserter interface {
	assert()
	begin(ctx context.Context) (transaction.Transaction, error)
	beginTx(ctx context.Context, opts *sql.TxOptions) (transaction.Transaction, error)
}

type Provider struct {
	asrt providerAsserter
}

func newProvider(t testReporter, asrt providerAsserter) *Provider {
	t.Cleanup(asrt.assert)

	return &Provider{asrt: asrt}
}

func (p *Provider) Begin(ctx context.Context) (transaction.Transaction, error) {
	return p.asrt.begin(ctx)
}

func (p *Provider) BeginTx(ctx context.Context, opts *sql.TxOptions) (transaction.Transaction, error) {
	return p.asrt.beginTx(ctx, opts)
}

func (b *Provider) TxEnabled(ctx context.Context) bool {
	_, ok := ctx.Value(txKey{}).(tx)

	return ok
}

type beginAndReturnError struct {
	t      testReporter
	err    error
	called atomic.Bool
}

func (b *beginAndReturnError) begin(context.Context) (transaction.Transaction, error) {
	swapped := b.called.CompareAndSwap(false, true)
	if !swapped {
		b.t.Fatal("unexpected call, provider.Begin called more than once")
	}

	return nil, b.err
}

func (b *beginAndReturnError) beginTx(context.Context, *sql.TxOptions) (transaction.Transaction, error) {
	b.t.Fatal("unexpected call to provider.BeginTx, expect one call to provider.Begin")

	return nil, nil
}

func (b *beginAndReturnError) assert() {
	called := b.called.Load()
	if !called {
		b.t.Fatal("provider assertion failed, no calls occurred")
	}
}

type beginTxAndReturnError struct {
	t            testReporter
	err          error
	expectedOpts *sql.TxOptions
	called       atomic.Bool
}

func (b *beginTxAndReturnError) begin(context.Context) (transaction.Transaction, error) {
	b.t.Fatal("unexpected call to provider.Begin, expect one call to provider.BeginTx")

	return nil, nil
}

func (b *beginTxAndReturnError) beginTx(_ context.Context, opts *sql.TxOptions) (transaction.Transaction, error) {
	swapped := b.called.CompareAndSwap(false, true)
	if !swapped {
		b.t.Fatal("unexpected call, provider.BeginTx called more than once")
	}

	sqlOptsEqual(b.t, b.expectedOpts, opts)

	return nil, b.err
}

func (b *beginTxAndReturnError) assert() {
	called := b.called.Load()
	if !called {
		b.t.Fatal("provider assertion failed, no calls occurred")
	}
}

type beginAndReturnTx struct {
	t      testReporter
	tx     *Transaction
	called atomic.Bool
}

func (b *beginAndReturnTx) begin(ctx context.Context) (transaction.Transaction, error) {
	swapped := b.called.CompareAndSwap(false, true)
	if !swapped {
		b.t.Fatal("unexpected call, provider.Begin called more than once")
	}

	b.tx.ctx = startTx(ctx)

	return b.tx, nil
}

func (b *beginAndReturnTx) beginTx(ctx context.Context, opts *sql.TxOptions) (transaction.Transaction, error) {
	b.t.Fatal("unexpected call to provider.BeginTx, expect one call to provider.Begin")

	return nil, nil
}

func (b *beginAndReturnTx) assert() {
	called := b.called.Load()
	if !called {
		b.t.Fatal("provider assertion failed, no calls occurred")
	}
}

type beginTxAndReturnTx struct {
	t            testReporter
	tx           *Transaction
	expectedOpts *sql.TxOptions
	called       atomic.Bool
}

func (b *beginTxAndReturnTx) begin(context.Context) (transaction.Transaction, error) {
	b.t.Fatal("unexpected call to provider.Begin, expect one call to provider.BeginTx")

	return nil, nil
}

func (b *beginTxAndReturnTx) beginTx(ctx context.Context, opts *sql.TxOptions) (transaction.Transaction, error) {
	swapped := b.called.CompareAndSwap(false, true)
	if !swapped {
		b.t.Fatal("unexpected call, provider.BeginTx called more than once")
	}

	sqlOptsEqual(b.t, b.expectedOpts, opts)

	b.tx.ctx = startTx(ctx)

	return b.tx, nil
}

func (b *beginTxAndReturnTx) assert() {
	called := b.called.Load()
	if !called {
		b.t.Fatal("provider assertion failed, no calls occurred")
	}
}

func sqlOptsEqual(t testReporter, expected, actual *sql.TxOptions) {
	switch expected {
	case nil:
		if actual == nil {
			return
		}

		tFatalUnexpectedOpts(t, expected, actual)
	default:
		if actual == nil {
			tFatalUnexpectedOpts(t, expected, actual)

			return
		}

		if *expected != *actual {
			tFatalUnexpectedOpts(t, expected, actual)
		}
	}
}

func tFatalUnexpectedOpts(t testReporter, expected, actual *sql.TxOptions) {
	t.Fatalf("unexpected call, call provider.BeginTx with %+v opts, expected %+v", actual, expected)
}
