package txmocks

import (
	"context"
	"sync"
	"sync/atomic"
)

type txKey struct{}

type mockTx struct{}

func startTx(ctx context.Context) context.Context {
	return context.WithValue(ctx, txKey{}, mockTx{})
}

type txAsserter interface {
	rollback() error
	commit() error
	assert()
}

type Tx struct {
	once sync.Once
	asrt txAsserter
	ctx  context.Context
}

func newTransaction(t testReporter, asrt txAsserter) *Tx {
	t.Cleanup(asrt.assert)

	ctx := startTx(context.Background())

	return &Tx{asrt: asrt, ctx: ctx}
}

func (t *Tx) Commit() error {
	t.clearTx()

	return t.asrt.commit()
}

func (t *Tx) Rollback() error {
	t.clearTx()

	return t.asrt.rollback()
}

func (t *Tx) Context() context.Context {
	return t.ctx
}

func (t *Tx) clearTx() {
	t.once.Do(func() {
		t.ctx = context.WithValue(t.ctx, txKey{}, nil)
	})
}

type TxMock func(t testReporter) *Tx

func ExpectRollback(err error) TxMock {
	return func(t testReporter) *Tx {
		asrt := &rollback{
			err: err,
			t:   t,
		}

		return newTransaction(t, asrt)
	}
}

type rollback struct {
	t      testReporter
	err    error
	called atomic.Bool
}

func (r *rollback) rollback() error {
	swapped := r.called.CompareAndSwap(false, true)
	if !swapped {
		r.t.Fatal("unexpected call, tx.Rollback called more than once")
	}

	return r.err
}

func (r *rollback) commit() error {
	r.t.Fatal("unexpected call to tx.Commit, expected one call to tx.Rollback")

	return nil
}

func (r *rollback) assert() {
	called := r.called.Load()
	if !called {
		r.t.Fatal("tx assertion failed, no calls occurred")
	}
}

func ExpectCommit(t testReporter) *Tx {
	asrt := &commit{
		t: t,
	}

	return newTransaction(t, asrt)
}

type commit struct {
	t      testReporter
	called atomic.Bool
}

func (c *commit) rollback() error {
	c.t.Fatal("unexpected call to tx.Rollback, expected one call to tx.Commit")

	return nil
}

func (c *commit) commit() error {
	swapped := c.called.CompareAndSwap(false, true)
	if !swapped {
		c.t.Fatal("unexpected call, tx.Commit called more than once")
	}

	return nil
}

func (c *commit) assert() {
	called := c.called.Load()
	if !called {
		c.t.Fatal("tx assertion failed, no calls occurred")
	}
}

func ExpectRollbackAfterFailedCommit(commitError error) TxMock {
	return func(t testReporter) *Tx {
		asrt := &rollbackAfterFailedCommit{
			t:         t,
			commitErr: commitError,
		}

		return newTransaction(t, asrt)
	}
}

const (
	notCommited int32 = iota
	commited
	rollbacked
)

type rollbackAfterFailedCommit struct {
	t           testReporter
	state       atomic.Int32
	commitErr   error
	rollbackErr error
}

func (t *rollbackAfterFailedCommit) rollback() error {
	swapped := t.state.CompareAndSwap(commited, rollbacked)
	if !swapped {
		t.t.Fatal("unexpected call, tx.Commit has not been called yet or tx.Rollback has been already called")
	}

	return nil
}

func (t *rollbackAfterFailedCommit) commit() error {
	swapped := t.state.CompareAndSwap(notCommited, commited)
	if !swapped {
		t.t.Fatal("unexpected call, tx.Commit has already was called, expect call tx.Rollback")
	}

	return t.commitErr
}

func (t *rollbackAfterFailedCommit) assert() {
	state := t.state.Load()
	switch state {
	case notCommited:
		t.t.Fatal("tx assertion failed, no calls occurred")
	case commited:
		t.t.Fatal("tx assertion failed, tx.Rollback not called")
	case rollbacked:
	}
}

func NilTx(t testReporter) *Tx {
	asrt := &nothing{
		t: t,
	}

	return newTransaction(t, asrt)
}

type nothing struct {
	t testReporter
}

func (n *nothing) rollback() error {
	n.t.Fatal("unexpected call to tx.Rollback")

	return nil
}

func (n *nothing) commit() error {
	n.t.Fatal("unexpected call to tx.Commit")

	return nil
}

func (n *nothing) assert() {}
