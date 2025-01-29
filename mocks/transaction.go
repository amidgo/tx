package transactionmocks

import (
	"context"
	"sync"
	"sync/atomic"
)

type txKey struct{}

type tx struct{}

func startTx(ctx context.Context) context.Context {
	return context.WithValue(ctx, txKey{}, tx{})
}

type transactionAsserter interface {
	rollback() error
	commit() error
	assert()
}

type Transaction struct {
	once sync.Once
	asrt transactionAsserter
	ctx  context.Context
}

func newTransaction(t testReporter, asrt transactionAsserter) *Transaction {
	t.Cleanup(asrt.assert)

	ctx := startTx(context.Background())

	return &Transaction{asrt: asrt, ctx: ctx}
}

func (t *Transaction) Commit() error {
	t.clearTx()

	return t.asrt.commit()
}

func (t *Transaction) Rollback() error {
	t.clearTx()

	return t.asrt.rollback()
}

func (t *Transaction) Context() context.Context {
	return t.ctx
}

func (t *Transaction) clearTx() {
	t.once.Do(func() {
		t.ctx = context.WithValue(t.ctx, txKey{}, nil)
	})
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

const (
	notCommited int32 = iota
	commited
	rollbacked
)

type rollbackAfterFailedCommit struct {
	t     testReporter
	state atomic.Int32
	err   error
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

	return t.err
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
