package mocks

import (
	"context"
	"database/sql"
	"sync"
	"sync/atomic"

	"github.com/amidgo/transaction"
)

type testReporter interface {
	Fatalf(format string, args ...any)
	Fatal(args ...any)
	Cleanup(func())
}

type Provider struct {
	t    testReporter
	once sync.Once
	asrt providerAsserter
}

func NewProvider(t testReporter) *Provider {
	p := &Provider{t: t}

	t.Cleanup(
		func() {
			if p.asrt != nil {
				p.asrt.assert()
			}
		},
	)

	return p
}

func (p *Provider) Begin(context.Context) (transaction.Transaction, error) {
	if p.asrt == nil {
		p.t.Fatal("unexpected call to provider.Begin")

		return nil, nil
	}

	return p.asrt.begin()
}

func (p *Provider) BeginTx(_ context.Context, opts sql.TxOptions) (transaction.Transaction, error) {
	if p.asrt == nil {
		p.t.Fatal("unexpected call to provider.BeginTx")

		return nil, nil
	}

	return p.asrt.beginTx(opts)
}

func (p *Provider) ExpectBeginAndReturnError(beginError error) {
	p.setAsserter(
		&beginAndReturnError{
			t:   p.t,
			err: beginError,
		},
	)
}

type beginAndReturnError struct {
	t      testReporter
	err    error
	called atomic.Bool
}

func (b *beginAndReturnError) begin() (transaction.Transaction, error) {
	swapped := b.called.CompareAndSwap(false, true)
	if !swapped {
		b.t.Fatal("unexpected call, provider.Begin called more than once")
	}

	return nil, b.err
}

func (b *beginAndReturnError) beginTx(sql.TxOptions) (transaction.Transaction, error) {
	b.t.Fatal("unexpected call to provider.BeginTx, expect one call to provider.Begin")

	return nil, nil
}

func (b *beginAndReturnError) assert() {
	called := b.called.Load()
	if !called {
		b.t.Fatal("provider assertion failed, no calls occurred")
	}
}

func (p *Provider) ExpectBeginTxAndReturnError(beginError error, expectedOpts sql.TxOptions) {
	p.setAsserter(
		&beginTxAndReturnError{
			t:            p.t,
			err:          beginError,
			expectedOpts: expectedOpts,
		},
	)
}

type beginTxAndReturnError struct {
	t            testReporter
	err          error
	expectedOpts sql.TxOptions
	called       atomic.Bool
}

func (b *beginTxAndReturnError) begin() (transaction.Transaction, error) {
	b.t.Fatal("unexpected call to provider.Begin, expect one call to provider.BeginTx")

	return nil, nil
}

func (b *beginTxAndReturnError) beginTx(opts sql.TxOptions) (transaction.Transaction, error) {
	swapped := b.called.CompareAndSwap(false, true)
	if !swapped {
		b.t.Fatal("unexpected call, provider.BeginTx called more than once")
	}

	if b.expectedOpts != opts {
		b.t.Fatalf("unexpected call, call provider.BeginTx with %v opts, expected %v", opts, b.expectedOpts)
	}

	return nil, b.err
}

func (b *beginTxAndReturnError) assert() {
	called := b.called.Load()
	if !called {
		b.t.Fatal("provider assertion failed, no calls occurred")
	}
}

func (p *Provider) ExpectBeginAndReturnTx(tx transaction.Transaction) {
	p.setAsserter(
		&beginAndReturnTx{
			t:  p.t,
			tx: tx,
		},
	)
}

type beginAndReturnTx struct {
	t      testReporter
	tx     transaction.Transaction
	called atomic.Bool
}

func (b *beginAndReturnTx) begin() (transaction.Transaction, error) {
	swapped := b.called.CompareAndSwap(false, true)
	if !swapped {
		b.t.Fatal("unexpected call, provider.Begin called more than once")
	}

	return b.tx, nil
}

func (b *beginAndReturnTx) beginTx(sql.TxOptions) (transaction.Transaction, error) {
	b.t.Fatal("unexpected call to provider.BeginTx, expect one call to provider.Begin")

	return nil, nil
}

func (b *beginAndReturnTx) assert() {
	called := b.called.Load()
	if !called {
		b.t.Fatal("provider assertion failed, no calls occurred")
	}
}

func (p *Provider) ExpectBeginTxAndReturnTx(tx transaction.Transaction, opts sql.TxOptions) {
	p.setAsserter(
		&beginTxAndReturnTx{
			t:            p.t,
			tx:           tx,
			expectedOpts: opts,
		},
	)
}

type beginTxAndReturnTx struct {
	t            testReporter
	tx           transaction.Transaction
	expectedOpts sql.TxOptions
	called       atomic.Bool
}

func (b *beginTxAndReturnTx) begin() (transaction.Transaction, error) {
	b.t.Fatal("unexpected call to provider.Begin, expect one call to provider.BeginTx")

	return nil, nil
}

func (b *beginTxAndReturnTx) beginTx(opts sql.TxOptions) (transaction.Transaction, error) {
	swapped := b.called.CompareAndSwap(false, true)
	if !swapped {
		b.t.Fatal("unexpected call, provider.BeginTx called more than once")
	}

	if b.expectedOpts != opts {
		b.t.Fatalf("unexpected call, call provider.BeginTx with %v opts, expected %v", opts, b.expectedOpts)
	}

	return b.tx, nil
}

func (b *beginTxAndReturnTx) assert() {
	called := b.called.Load()
	if !called {
		b.t.Fatal("provider assertion failed, no calls occurred")
	}
}

type providerAsserter interface {
	assert()
	begin() (transaction.Transaction, error)
	beginTx(sql.TxOptions) (transaction.Transaction, error)
}

func (p *Provider) setAsserter(asrt providerAsserter) {
	p.once.Do(func() { p.asrt = asrt })
}
