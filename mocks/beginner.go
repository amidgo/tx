package txmocks

import (
	"context"
	"database/sql"
	"sync"
	"sync/atomic"

	"github.com/amidgo/tx"
)

var _ tx.Beginner = (*Beginner)(nil)

type beginnerAsserter interface {
	assert()
	begin(ctx context.Context) (tx.Tx, error)
	beginTx(ctx context.Context, opts *sql.TxOptions) (tx.Tx, error)
}

type Beginner struct {
	asrt beginnerAsserter
}

func newBeginner(t testReporter, asrt beginnerAsserter) *Beginner {
	t.Cleanup(asrt.assert)

	return &Beginner{asrt: asrt}
}

func (p *Beginner) Begin(ctx context.Context) (tx.Tx, error) {
	return p.asrt.begin(ctx)
}

func (p *Beginner) BeginTx(ctx context.Context, opts *sql.TxOptions) (tx.Tx, error) {
	return p.asrt.beginTx(ctx, opts)
}

func (b *Beginner) TxEnabled(ctx context.Context) bool {
	return txEnabled(ctx)
}

func txEnabled(ctx context.Context) bool {
	_, ok := ctx.Value(txKey{}).(mockTx)

	return ok
}

type BeginnerMock func(t testReporter) *Beginner

func ExpectBeginAndReturnError(beginError error) BeginnerMock {
	return func(t testReporter) *Beginner {
		asrt := &beginAndReturnError{
			t:   t,
			err: beginError,
		}

		return newBeginner(t, asrt)
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
		b.t.Fatal("unexpected call, beginner.Begin called more than once")
	}

	return nil, b.err
}

func (b *beginAndReturnError) beginTx(context.Context, *sql.TxOptions) (tx.Tx, error) {
	b.t.Fatal("unexpected call to beginner.BeginTx, expect one call to beginner.Begin")

	return nil, nil
}

func (b *beginAndReturnError) assert() {
	called := b.called.Load()
	if !called {
		b.t.Fatal("beginner assertion failed, no calls occurred")
	}
}

func ExpectBeginTxAndReturnError(beginError error, expectedOpts *sql.TxOptions) BeginnerMock {
	return func(t testReporter) *Beginner {
		asrt := &beginTxAndReturnError{
			t:            t,
			err:          beginError,
			expectedOpts: expectedOpts,
		}

		return newBeginner(t, asrt)
	}
}

type beginTxAndReturnError struct {
	t            testReporter
	err          error
	expectedOpts *sql.TxOptions
	called       atomic.Bool
}

func (b *beginTxAndReturnError) begin(context.Context) (tx.Tx, error) {
	b.t.Fatal("unexpected call to beginner.Begin, expect one call to beginner.BeginTx")

	return nil, nil
}

func (b *beginTxAndReturnError) beginTx(_ context.Context, opts *sql.TxOptions) (tx.Tx, error) {
	swapped := b.called.CompareAndSwap(false, true)
	if !swapped {
		b.t.Fatal("unexpected call, beginner.BeginTx called more than once")
	}

	sqlOptsEqual(b.t, b.expectedOpts, opts)

	return nil, b.err
}

func (b *beginTxAndReturnError) assert() {
	called := b.called.Load()
	if !called {
		b.t.Fatal("beginner assertion failed, no calls occurred")
	}
}

func ExpectBeginAndReturnTx(txMock TxMock) BeginnerMock {
	return func(t testReporter) *Beginner {
		asrt := &beginAndReturnTx{
			t:  t,
			tx: txMock(t),
		}

		return newBeginner(t, asrt)
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
		b.t.Fatal("unexpected call, beginner.Begin called more than once")
	}

	b.tx.ctx = startTx(ctx)

	return b.tx, nil
}

func (b *beginAndReturnTx) beginTx(ctx context.Context, opts *sql.TxOptions) (tx.Tx, error) {
	b.t.Fatal("unexpected call to beginner.BeginTx, expect one call to beginner.Begin")

	return nil, nil
}

func (b *beginAndReturnTx) assert() {
	called := b.called.Load()
	if !called {
		b.t.Fatal("beginner assertion failed, no calls occurred")
	}
}

func ExpectBeginTxAndReturnTx(tx TxMock, opts *sql.TxOptions) BeginnerMock {
	return func(t testReporter) *Beginner {
		asrt := &beginTxAndReturnTx{
			t:            t,
			tx:           tx(t),
			expectedOpts: opts,
		}

		return newBeginner(t, asrt)
	}
}

type beginTxAndReturnTx struct {
	t            testReporter
	tx           *Tx
	expectedOpts *sql.TxOptions
	called       atomic.Bool
}

func (b *beginTxAndReturnTx) begin(context.Context) (tx.Tx, error) {
	b.t.Fatal("unexpected call to beginner.Begin, expect one call to beginner.BeginTx")

	return nil, nil
}

func (b *beginTxAndReturnTx) beginTx(ctx context.Context, opts *sql.TxOptions) (tx.Tx, error) {
	swapped := b.called.CompareAndSwap(false, true)
	if !swapped {
		b.t.Fatal("unexpected call, beginner.BeginTx called more than once")
	}

	sqlOptsEqual(b.t, b.expectedOpts, opts)

	b.tx.ctx = startTx(ctx)

	return b.tx, nil
}

func (b *beginTxAndReturnTx) assert() {
	called := b.called.Load()
	if !called {
		b.t.Fatal("beginner assertion failed, no calls occurred")
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
	t.Fatalf("unexpected call, call beginner.BeginTx with %+v opts, expected %+v", actual, expected)
}

func JoinBeginners(beginners ...BeginnerMock) BeginnerMock {
	return func(t testReporter) *Beginner {
		switch len(beginners) {
		case 0:
			t.Fatal("empty join beginner templates")

			return nil
		case 1:
			return beginners[0](t)
		}

		asrts := make([]beginnerAsserter, len(beginners))

		for i := range beginners {
			index := len(beginners) - 1 - i

			prv := beginners[index](t)

			if prv.asrt == nil {
				t.Fatalf("invalid beginner by index %d, Beginner.asrt is nil", index)

				return nil
			}

			asrts[index] = prv.asrt
		}

		asrt := &beginnerAsserterJoin{
			t:     t,
			asrts: asrts,
		}

		return newBeginner(t, asrt)
	}
}

type beginnerAsserterJoin struct {
	t            testReporter
	asrts        []beginnerAsserter
	currentIndex int
	mu           sync.Mutex
}

func (p *beginnerAsserterJoin) begin(ctx context.Context) (tx.Tx, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	asrt, expected := p.currentAsserter()
	if !expected {
		p.t.Fatal("unexpected call to beginner.Begin, no calls left")

		return nil, nil
	}

	tx, err := asrt.begin(ctx)

	p.currentIndex++

	return tx, err
}

func (p *beginnerAsserterJoin) beginTx(ctx context.Context, opts *sql.TxOptions) (tx.Tx, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	asrt, expected := p.currentAsserter()
	if !expected {
		p.t.Fatal("unexpected call to beginner.BeginTx, no calls left")

		return nil, nil
	}

	tx, err := asrt.beginTx(ctx, opts)

	p.currentIndex++

	return tx, err
}

func (p *beginnerAsserterJoin) assert() {
	for _, asrt := range p.asrts {
		asrt.assert()
	}
}

func (p *beginnerAsserterJoin) currentAsserter() (beginnerAsserter, bool) {
	if p.currentIndex > len(p.asrts)-1 {
		return nil, false
	}

	return p.asrts[p.currentIndex], true
}
