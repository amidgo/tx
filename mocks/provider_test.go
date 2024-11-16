package mocks_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/amidgo/tester"
	"github.com/amidgo/transaction"
	"github.com/amidgo/transaction/mocks"
)

type mockTestReporter struct {
	t      *testing.T
	called bool
}

func newMockTestReporter(t *testing.T, expectCalled bool) *mockTestReporter {
	r := &mockTestReporter{t: t}

	t.Cleanup(
		func() {
			requireEqual(t, expectCalled, r.called)
		},
	)

	return r
}

func (r *mockTestReporter) Fatal(...any) {
	r.called = true
}

func (r *mockTestReporter) Fatalf(string, ...any) {
	r.called = true
}

func (r *mockTestReporter) Cleanup(f func()) {
	r.t.Cleanup(f)
}

func Test_Provider_ExpectBeginAndReturnError_Valid(t *testing.T) {
	testReporter := newMockTestReporter(t, false)

	beginError := errors.New("begin error")

	provider := mocks.ExpectBeginAndReturnError(beginError)(testReporter)

	tx, err := provider.Begin(context.Background())
	requireNil(t, tx)
	requireErrorIs(t, err, beginError)
}

func Test_Provider_ExpectBeginAndReturnError_CalledTwice(t *testing.T) {
	testReporter := newMockTestReporter(t, true)

	beginError := errors.New("begin error")

	provider := mocks.ExpectBeginAndReturnError(beginError)(testReporter)

	tx, err := provider.Begin(context.Background())
	requireNil(t, tx)
	requireErrorIs(t, err, beginError)

	tx, err = provider.Begin(context.Background())
	requireNil(t, tx)
	requireErrorIs(t, err, beginError)
}

func Test_Provider_ExpectBeginAndReturnError_Expect_But_Not_Called(t *testing.T) {
	testReporter := newMockTestReporter(t, true)

	beginError := errors.New("begin error")

	mocks.ExpectBeginAndReturnError(beginError)(testReporter)
}

func Test_Provider_ExpectBeginAndReturnError_CallBeginTx(t *testing.T) {
	testReporter := newMockTestReporter(t, true)

	beginError := errors.New("begin error")

	provider := mocks.ExpectBeginAndReturnError(beginError)(testReporter)

	tx, err := provider.BeginTx(context.Background(), sql.TxOptions{})
	requireNil(t, tx)
	requireNoError(t, err)
}

func Test_Provider_ExpectBeginTxAndReturnError_Valid(t *testing.T) {
	testReporter := newMockTestReporter(t, false)

	beginTxError := errors.New("begin tx error")
	opts := sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}

	provider := mocks.ExpectBeginTxAndReturnError(beginTxError, opts)(testReporter)

	tx, err := provider.BeginTx(context.Background(), opts)
	requireNil(t, tx)
	requireErrorIs(t, err, beginTxError)
}

func Test_Provider_ExpectBeginTxAndReturnError_CalledTwice(t *testing.T) {
	testReporter := newMockTestReporter(t, true)

	beginTxError := errors.New("begin tx error")
	opts := sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}

	provider := mocks.ExpectBeginTxAndReturnError(beginTxError, opts)(testReporter)

	tx, err := provider.BeginTx(context.Background(), opts)
	requireNil(t, tx)
	requireErrorIs(t, err, beginTxError)

	tx, err = provider.BeginTx(context.Background(), opts)
	requireNil(t, tx)
	requireErrorIs(t, err, beginTxError)
}

func Test_Provider_ExpectBeginTxAndReturnError_Call_With_Unexpected_Opts(t *testing.T) {
	testReporter := newMockTestReporter(t, true)

	beginTxError := errors.New("begin tx error")
	opts := sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}

	provider := mocks.ExpectBeginTxAndReturnError(beginTxError, opts)(testReporter)

	tx, err := provider.BeginTx(context.Background(), sql.TxOptions{Isolation: sql.LevelDefault})
	requireNil(t, tx)
	requireErrorIs(t, err, beginTxError)
}

func Test_Provider_ExpectBeginTxAndReturnError_Expect_But_Not_Called(t *testing.T) {
	testReporter := newMockTestReporter(t, true)

	beginTxError := errors.New("begin tx error")
	opts := sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}

	mocks.ExpectBeginTxAndReturnError(beginTxError, opts)(testReporter)
}

func Test_Provider_ExpectBeginTxAndReturnError_CalledBegin(t *testing.T) {
	testReporter := newMockTestReporter(t, true)

	beginTxError := errors.New("begin tx error")
	opts := sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}

	provider := mocks.ExpectBeginTxAndReturnError(beginTxError, opts)(testReporter)

	tx, err := provider.Begin(context.Background())
	requireNil(t, tx)
	requireNoError(t, err)
}

func Test_Provider_ExpectBeginAndReturnTx_Valid(t *testing.T) {
	testReporter := newMockTestReporter(t, false)

	provider := mocks.ExpectBeginAndReturnTx(mocks.ExpectNothing())(testReporter)

	tx, err := provider.Begin(context.Background())
	requireNoError(t, err)
	requireNotNil(t, tx)
}

func Test_Provider_ExpectBeginAndReturnTx_CalledTwice(t *testing.T) {
	testReporter := newMockTestReporter(t, true)

	provider := mocks.ExpectBeginAndReturnTx(mocks.ExpectNothing())(testReporter)

	tx, err := provider.Begin(context.Background())
	requireNoError(t, err)
	requireNotNil(t, tx)

	tx, err = provider.Begin(context.Background())
	requireNoError(t, err)
	requireNotNil(t, tx)
}

func Test_Provider_ExpectBeginAndReturnTx_Expect_But_Not_Called(t *testing.T) {
	testReporter := newMockTestReporter(t, true)

	mocks.ExpectBeginAndReturnTx(mocks.ExpectNothing())(testReporter)
}

func Test_Provider_ExpectBeginAndReturnTx_CalledBeginTx(t *testing.T) {
	testReporter := newMockTestReporter(t, true)

	provider := mocks.ExpectBeginAndReturnTx(mocks.ExpectNothing())(testReporter)

	tx, err := provider.BeginTx(context.Background(), sql.TxOptions{})
	requireNoError(t, err)
	requireNil(t, tx)
}

func Test_Provider_ExpectBeginTxAndReturnTx_Valid(t *testing.T) {
	testReporter := newMockTestReporter(t, false)

	opts := sql.TxOptions{Isolation: sql.LevelReadCommitted}

	provider := mocks.ExpectBeginTxAndReturnTx(mocks.ExpectNothing(), opts)(testReporter)

	tx, err := provider.BeginTx(context.Background(), opts)
	requireNoError(t, err)
	requireNotNil(t, tx)
}

func Test_Provider_ExpectBeginTxAndReturnTx_CalledTwice(t *testing.T) {
	testReporter := newMockTestReporter(t, true)

	opts := sql.TxOptions{Isolation: sql.LevelReadCommitted}

	provider := mocks.ExpectBeginTxAndReturnTx(mocks.ExpectNothing(), opts)(testReporter)

	tx, err := provider.BeginTx(context.Background(), opts)
	requireNoError(t, err)
	requireNotNil(t, tx)

	tx, err = provider.BeginTx(context.Background(), opts)
	requireNoError(t, err)
	requireNotNil(t, tx)
}

func Test_Provider_ExpectBeginTxAndReturnTx_Call_With_Unexpected_Opts(t *testing.T) {
	testReporter := newMockTestReporter(t, true)

	opts := sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}

	provider := mocks.ExpectBeginTxAndReturnTx(mocks.ExpectNothing(), opts)(testReporter)

	tx, err := provider.BeginTx(context.Background(), sql.TxOptions{Isolation: sql.LevelDefault})
	requireNotNil(t, tx)
	requireNoError(t, err)
}

func Test_Provider_ExpectBeginTxAndReturnTx_Expect_But_Not_Called(t *testing.T) {
	testReporter := newMockTestReporter(t, true)

	opts := sql.TxOptions{Isolation: sql.LevelReadCommitted}

	mocks.ExpectBeginTxAndReturnTx(mocks.ExpectNothing(), opts)(testReporter)
}

func Test_Provider_ExpectBeginTxAndReturnTx_CalledBegin(t *testing.T) {
	testReporter := newMockTestReporter(t, true)

	opts := sql.TxOptions{Isolation: sql.LevelReadCommitted}

	provider := mocks.ExpectBeginTxAndReturnTx(mocks.ExpectNothing(), opts)(testReporter)

	tx, err := provider.Begin(context.Background())
	requireNil(t, tx)
	requireNoError(t, err)
}

type ProviderJoinTest struct {
	CaseName          string
	ProviderTemplates []mocks.ProviderTemplate
	WithProvider      func(p transaction.Provider)
	ExpectReport      bool
}

func (p *ProviderJoinTest) Name() string {
	return p.CaseName
}

func (p *ProviderJoinTest) Test(t *testing.T) {
	testReporter := newMockTestReporter(t, p.ExpectReport)

	provider := mocks.ProviderJoin(p.ProviderTemplates...)(testReporter)

	if p.WithProvider != nil {
		p.WithProvider(provider)
	}
}

func Test_Provider_Join(t *testing.T) {
	errBeginTx := errors.New("begin tx")
	opts := sql.TxOptions{
		Isolation: sql.LevelDefault,
	}

	tester.RunNamedTesters(t,
		&ProviderJoinTest{
			CaseName:          "zero operations",
			ProviderTemplates: nil,
			ExpectReport:      true,
		},
		&ProviderJoinTest{
			CaseName: "single operation, valid",
			ProviderTemplates: []mocks.ProviderTemplate{
				mocks.ExpectBeginAndReturnError(errBeginTx),
			},
			WithProvider: func(p transaction.Provider) {
				tx, err := p.Begin(context.Background())
				requireErrorIs(t, err, errBeginTx)
				requireNil(t, tx)
			},
			ExpectReport: false,
		},
		&ProviderJoinTest{
			CaseName: "single operation, not valid",
			ProviderTemplates: []mocks.ProviderTemplate{
				mocks.ExpectBeginAndReturnError(errBeginTx),
			},
			WithProvider: func(p transaction.Provider) {
				tx, err := p.Begin(context.Background())
				requireErrorIs(t, err, errBeginTx)
				requireNil(t, tx)

				tx, err = p.Begin(context.Background())
				requireErrorIs(t, err, errBeginTx)
				requireNil(t, tx)
			},
			ExpectReport: true,
		},
		&ProviderJoinTest{
			CaseName: "two operations, valid",
			ProviderTemplates: []mocks.ProviderTemplate{
				mocks.ExpectBeginAndReturnError(errBeginTx),
				mocks.ExpectBeginAndReturnError(errBeginTx),
			},
			WithProvider: func(p transaction.Provider) {
				tx, err := p.Begin(context.Background())
				requireErrorIs(t, err, errBeginTx)
				requireNil(t, tx)

				tx, err = p.Begin(context.Background())
				requireErrorIs(t, err, errBeginTx)
				requireNil(t, tx)
			},
			ExpectReport: false,
		},
		&ProviderJoinTest{
			CaseName: "two operations, invalid count times",
			ProviderTemplates: []mocks.ProviderTemplate{
				mocks.ExpectBeginAndReturnError(errBeginTx),
				mocks.ExpectBeginAndReturnError(errBeginTx),
			},
			WithProvider: func(p transaction.Provider) {
				tx, err := p.Begin(context.Background())
				requireErrorIs(t, err, errBeginTx)
				requireNil(t, tx)

				tx, err = p.Begin(context.Background())
				requireErrorIs(t, err, errBeginTx)
				requireNil(t, tx)

				tx, err = p.Begin(context.Background())
				requireNoError(t, err)
				requireNil(t, tx)
			},
			ExpectReport: true,
		},
		&ProviderJoinTest{
			CaseName: "two operations, valid order",
			ProviderTemplates: []mocks.ProviderTemplate{
				mocks.ExpectBeginAndReturnError(errBeginTx),
				mocks.ExpectBeginTxAndReturnError(errBeginTx, opts),
			},
			WithProvider: func(p transaction.Provider) {
				tx, err := p.Begin(context.Background())
				requireErrorIs(t, err, errBeginTx)
				requireNil(t, tx)

				tx, err = p.BeginTx(context.Background(), opts)
				requireErrorIs(t, err, errBeginTx)
				requireNil(t, tx)
			},
			ExpectReport: false,
		},
		&ProviderJoinTest{
			CaseName: "two operations, invalid order",
			ProviderTemplates: []mocks.ProviderTemplate{
				mocks.ExpectBeginAndReturnError(errBeginTx),
				mocks.ExpectBeginTxAndReturnError(errBeginTx, opts),
			},
			WithProvider: func(p transaction.Provider) {
				tx, err := p.BeginTx(context.Background(), opts)
				requireNil(t, err)
				requireNil(t, tx)

				tx, err = p.Begin(context.Background())
				requireNil(t, err)
				requireNil(t, tx)
			},
			ExpectReport: true,
		},
		&ProviderJoinTest{
			CaseName: "to many operations, with transactions",
			ProviderTemplates: []mocks.ProviderTemplate{
				mocks.ExpectBeginAndReturnError(errBeginTx),
				mocks.ExpectBeginTxAndReturnError(errBeginTx, opts),
				mocks.ExpectBeginAndReturnTx(mocks.ExpectCommit()),
				mocks.ExpectBeginTxAndReturnTx(mocks.ExpectRollback(nil), opts),
			},
			WithProvider: func(p transaction.Provider) {
				tx, err := p.Begin(context.Background())
				requireErrorIs(t, err, errBeginTx)
				requireNil(t, tx)

				tx, err = p.BeginTx(context.Background(), opts)
				requireErrorIs(t, err, errBeginTx)
				requireNil(t, tx)

				tx, err = p.Begin(context.Background())
				requireNoError(t, err)
				requireNotNil(t, tx)

				err = tx.Commit(context.Background())
				requireNoError(t, err)

				tx, err = p.BeginTx(context.Background(), opts)
				requireNoError(t, err)
				requireNotNil(t, tx)

				tx.Rollback(context.Background())
			},
			ExpectReport: false,
		},
	)
}
