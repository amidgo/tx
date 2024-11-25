package mocks

import (
	"database/sql"
	"sync"

	"github.com/amidgo/transaction"
)

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

func ExpectBeginTxAndReturnError(beginError error, expectedOpts sql.TxOptions) ProviderMock {
	return func(t testReporter) *Provider {
		asrt := &beginTxAndReturnError{
			t:            t,
			err:          beginError,
			expectedOpts: expectedOpts,
		}

		return newProvider(t, asrt)
	}
}

func ExpectBeginAndReturnTx(tx TransactionMock) ProviderMock {
	return func(t testReporter) *Provider {
		asrt := &beginAndReturnTx{
			t:  t,
			tx: tx(t),
		}

		return newProvider(t, asrt)
	}
}

func ExpectBeginTxAndReturnTx(tx TransactionMock, opts sql.TxOptions) ProviderMock {
	return func(t testReporter) *Provider {
		asrt := &beginTxAndReturnTx{
			t:            t,
			tx:           tx(t),
			expectedOpts: opts,
		}

		return newProvider(t, asrt)
	}
}

func ProviderJoin(tmpls ...ProviderMock) ProviderMock {
	return func(t testReporter) *Provider {
		switch len(tmpls) {
		case 0:
			t.Fatal("empty join provider templates")

			return nil
		case 1:
			return tmpls[0](t)
		}

		asrts := make([]providerAsserter, len(tmpls))

		for i := range tmpls {
			index := len(tmpls) - 1 - i

			prv := tmpls[index](t)

			if prv.asrt == nil {
				t.Fatalf("invalid provider template by index %d, Provider.asrt is nil", index)

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

func (p *providerAsserterJoin) begin() (transaction.Transaction, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	asrt, expected := p.currentAsserter()
	if !expected {
		p.t.Fatal("unexpected call to provider.Begin, no calls left")

		return nil, nil
	}

	tx, err := asrt.begin()

	p.currentIndex++

	return tx, err
}

func (p *providerAsserterJoin) beginTx(opts sql.TxOptions) (transaction.Transaction, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	asrt, expected := p.currentAsserter()
	if !expected {
		p.t.Fatal("unexpected call to provider.BeginTx, no calls left")

		return nil, nil
	}

	tx, err := asrt.beginTx(opts)

	p.currentIndex++

	return tx, err
}

func (p *providerAsserterJoin) assert() {}

func (p *providerAsserterJoin) currentAsserter() (providerAsserter, bool) {
	if p.currentIndex > len(p.asrts)-1 {
		return nil, false
	}

	return p.asrts[p.currentIndex], true
}
