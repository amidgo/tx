package txmocks

import (
	"context"
	"database/sql"
	"sync"
	"sync/atomic"

	"github.com/amidgo/tx"
)

var _ tx.Provider = (*Provider)(nil)

type providerAsserter interface {
	assert()
	begin(ctx context.Context) (tx.Tx, error)
	beginTx(ctx context.Context, opts *sql.TxOptions) (tx.Tx, error)
}

type Provider struct {
	asrt providerAsserter
}

func newProvider(t testReporter, asrt providerAsserter) *Provider {
	t.Cleanup(asrt.assert)

	return &Provider{asrt: asrt}
}

func (p *Provider) Begin(ctx context.Context) (tx.Tx, error) {
	return p.asrt.begin(ctx)
}

func (p *Provider) BeginTx(ctx context.Context, opts *sql.TxOptions) (tx.Tx, error) {
	return p.asrt.beginTx(ctx, opts)
}

func (b *Provider) TxEnabled(ctx context.Context) bool {
	_, ok := ctx.Value(txKey{}).(mockTx)

	return ok
}

type ProviderMock func(t testReporter) *Provider

func ExpectBeginAndReturnError(beginError error) ProviderMock {
	return func(t testReporter) *Provider {
		asrt := &beginAndReturnError{
			t:   t,
			err: beginError,
		}

		return newProvider(t, asrt)
	}
}

type beginAndReturnError struct {
	t      testReporter
	err    error
	called atomic.Bool
}

func (b *beginAndReturnError) begin(context.Context) (tx.Tx, error) {
	swapped := b.called.CompareAndSwap(false, true)
	if !swapped {
		b.t.Fatal("unexpected call, provider.Begin called more than once")
	}

	return nil, b.err
}

func (b *beginAndReturnError) beginTx(context.Context, *sql.TxOptions) (tx.Tx, error) {
	b.t.Fatal("unexpected call to provider.BeginTx, expect one call to provider.Begin")

	return nil, nil
}

func (b *beginAndReturnError) assert() {
	called := b.called.Load()
	if !called {
		b.t.Fatal("provider assertion failed, no calls occurred")
	}
}

func ExpectBeginTxAndReturnError(beginError error, expectedOpts *sql.TxOptions) ProviderMock {
	return func(t testReporter) *Provider {
		asrt := &beginTxAndReturnError{
			t:            t,
			err:          beginError,
			expectedOpts: expectedOpts,
		}

		return newProvider(t, asrt)
	}
}

type beginTxAndReturnError struct {
	t            testReporter
	err          error
	expectedOpts *sql.TxOptions
	called       atomic.Bool
}

func (b *beginTxAndReturnError) begin(context.Context) (tx.Tx, error) {
	b.t.Fatal("unexpected call to provider.Begin, expect one call to provider.BeginTx")

	return nil, nil
}

func (b *beginTxAndReturnError) beginTx(_ context.Context, opts *sql.TxOptions) (tx.Tx, error) {
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

func ExpectBeginAndReturnTx(txMock TxMock) ProviderMock {
	return func(t testReporter) *Provider {
		asrt := &beginAndReturnTx{
			t:  t,
			tx: txMock(t),
		}

		return newProvider(t, asrt)
	}
}

type beginAndReturnTx struct {
	t      testReporter
	tx     *Tx
	called atomic.Bool
}

func (b *beginAndReturnTx) begin(ctx context.Context) (tx.Tx, error) {
	swapped := b.called.CompareAndSwap(false, true)
	if !swapped {
		b.t.Fatal("unexpected call, provider.Begin called more than once")
	}

	b.tx.ctx = startTx(ctx)

	return b.tx, nil
}

func (b *beginAndReturnTx) beginTx(ctx context.Context, opts *sql.TxOptions) (tx.Tx, error) {
	b.t.Fatal("unexpected call to provider.BeginTx, expect one call to provider.Begin")

	return nil, nil
}

func (b *beginAndReturnTx) assert() {
	called := b.called.Load()
	if !called {
		b.t.Fatal("provider assertion failed, no calls occurred")
	}
}

func ExpectBeginTxAndReturnTx(tx TxMock, opts *sql.TxOptions) ProviderMock {
	return func(t testReporter) *Provider {
		asrt := &beginTxAndReturnTx{
			t:            t,
			tx:           tx(t),
			expectedOpts: opts,
		}

		return newProvider(t, asrt)
	}
}

type beginTxAndReturnTx struct {
	t            testReporter
	tx           *Tx
	expectedOpts *sql.TxOptions
	called       atomic.Bool
}

func (b *beginTxAndReturnTx) begin(context.Context) (tx.Tx, error) {
	b.t.Fatal("unexpected call to provider.Begin, expect one call to provider.BeginTx")

	return nil, nil
}

func (b *beginTxAndReturnTx) beginTx(ctx context.Context, opts *sql.TxOptions) (tx.Tx, error) {
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

func JoinProviders(providers ...ProviderMock) ProviderMock {
	return func(t testReporter) *Provider {
		switch len(providers) {
		case 0:
			t.Fatal("empty join provider templates")

			return nil
		case 1:
			return providers[0](t)
		}

		asrts := make([]providerAsserter, len(providers))

		for i := range providers {
			index := len(providers) - 1 - i

			prv := providers[index](t)

			if prv.asrt == nil {
				t.Fatalf("invalid provider by index %d, Provider.asrt is nil", index)

				return nil
			}

			asrts[index] = prv.asrt
		}

		asrt := &providerAsserterJoin{
			t:     t,
			asrts: asrts,
		}

		return newProvider(t, asrt)
	}
}

type providerAsserterJoin struct {
	t            testReporter
	asrts        []providerAsserter
	currentIndex int
	mu           sync.Mutex
}

func (p *providerAsserterJoin) begin(ctx context.Context) (tx.Tx, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	asrt, expected := p.currentAsserter()
	if !expected {
		p.t.Fatal("unexpected call to provider.Begin, no calls left")

		return nil, nil
	}

	tx, err := asrt.begin(ctx)

	p.currentIndex++

	return tx, err
}

func (p *providerAsserterJoin) beginTx(ctx context.Context, opts *sql.TxOptions) (tx.Tx, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	asrt, expected := p.currentAsserter()
	if !expected {
		p.t.Fatal("unexpected call to provider.BeginTx, no calls left")

		return nil, nil
	}

	tx, err := asrt.beginTx(ctx, opts)

	p.currentIndex++

	return tx, err
}

func (p *providerAsserterJoin) assert() {
	for _, asrt := range p.asrts {
		asrt.assert()
	}
}

func (p *providerAsserterJoin) currentAsserter() (providerAsserter, bool) {
	if p.currentIndex > len(p.asrts)-1 {
		return nil, false
	}

	return p.asrts[p.currentIndex], true
}
