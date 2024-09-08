package mocks

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/amidgo/transaction"
)

type Transaction struct {
	t    testReporter
	asrt transactionAsserter
	once sync.Once
}

func NewTransaction(t testReporter) *Transaction {
	tx := &Transaction{t: t}

	t.Cleanup(
		func() {
			if tx.asrt != nil {
				tx.asrt.assert()
			}
		},
	)

	return tx
}

func (t *Transaction) Commit(context.Context) error {
	if t.asrt == nil {
		t.t.Fatal("unexpected call to tx.Commit")

		return nil
	}

	return t.asrt.commit()
}

func (t *Transaction) Rollback(context.Context) {
	if t.asrt == nil {
		t.t.Fatal("unexpected call to tx.Rollback")

		return
	}

	t.asrt.rollback()
}

func (t *Transaction) Context() context.Context {
	return transaction.MockEnableTxContext(context.Background())
}

func (t *Transaction) ExpectRollback() {
	t.setAsserter(
		&rollback{
			t: t.t,
		},
	)
}

type rollback struct {
	t      testReporter
	called atomic.Bool
}

func (r *rollback) rollback() {
	swapped := r.called.CompareAndSwap(false, true)
	if !swapped {
		r.t.Fatal("unexpected call, tx.Rollback called more than once")
	}
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

func (t *Transaction) ExpectCommit() {
	t.setAsserter(
		&commit{
			t: t.t,
		},
	)
}

type commit struct {
	t      testReporter
	called atomic.Bool
}

func (c *commit) rollback() {
	c.t.Fatal("unexpected call to tx.Rollback, expected one call to tx.Commit")
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

func (t *Transaction) ExpectRollbackAfterFailedCommit(commitErr error) {
	t.setAsserter(
		&rollbackAfterFailedCommit{
			t:   t.t,
			err: commitErr,
		},
	)
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

func (t *rollbackAfterFailedCommit) rollback() {
	swapped := t.state.CompareAndSwap(commited, rollbacked)
	if !swapped {
		t.t.Fatal("unexpected call, tx.Commit has not been called yet or tx.Rollback has been already called")
	}
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

func (t *Transaction) setAsserter(asrt transactionAsserter) {
	t.once.Do(func() { t.asrt = asrt })
}

type transactionAsserter interface {
	rollback()
	commit() error
	assert()
}
