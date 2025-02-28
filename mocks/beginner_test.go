package txmocks_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/amidgo/tx"
	txmocks "github.com/amidgo/tx/mocks"
)

type mockTestReporter struct {
	t             *testing.T
	tFatalOnce    sync.Once
	tFatalMessage string
}

func newMockTestReporter(t *testing.T, tFatalMessage string) *mockTestReporter {
	r := &mockTestReporter{t: t}

	t.Cleanup(
		func() {
			requireEqual(t, tFatalMessage, r.tFatalMessage)
		},
	)

	return r
}

func (r *mockTestReporter) Fatal(args ...any) {
	r.tFatalOnce.Do(func() {
		r.tFatalMessage = fmt.Sprint(args...)
	})
}

func (r *mockTestReporter) Fatalf(format string, args ...any) {
	r.tFatalOnce.Do(func() {
		r.tFatalMessage = fmt.Sprintf(format, args...)
	})
}

func (r *mockTestReporter) Cleanup(f func()) {
	r.t.Cleanup(f)
}

func Test_Provider_ExpectBeginAndReturnError_Valid(t *testing.T) {
	testReporter := newMockTestReporter(t, "")

	beginError := errors.New("begin error")

	provider := txmocks.ExpectBeginAndReturnError(beginError)(testReporter)

	tx, err := provider.Begin(context.Background())
	requireNil(t, tx)
	requireErrorIs(t, err, beginError)
}

func Test_Provider_ExpectBeginAndReturnError_CalledTwice(t *testing.T) {
	testReporter := newMockTestReporter(t, "unexpected call, provider.Begin called more than once")

	beginError := errors.New("begin error")

	provider := txmocks.ExpectBeginAndReturnError(beginError)(testReporter)

	tx, err := provider.Begin(context.Background())
	requireNil(t, tx)
	requireErrorIs(t, err, beginError)

	tx, err = provider.Begin(context.Background())
	requireNil(t, tx)
	requireErrorIs(t, err, beginError)
}

func Test_Provider_ExpectBeginAndReturnError_Expect_But_Not_Called(t *testing.T) {
	testReporter := newMockTestReporter(t, "provider assertion failed, no calls occurred")

	beginError := errors.New("begin error")

	txmocks.ExpectBeginAndReturnError(beginError)(testReporter)
}

func Test_Provider_ExpectBeginAndReturnError_CallBeginTx(t *testing.T) {
	testReporter := newMockTestReporter(t, "unexpected call to provider.BeginTx, expect one call to provider.Begin")

	beginError := errors.New("begin error")

	provider := txmocks.ExpectBeginAndReturnError(beginError)(testReporter)

	tx, err := provider.BeginTx(context.Background(), nil)
	requireNil(t, tx)
	requireNoError(t, err)
}

func Test_Provider_ExpectBeginTxAndReturnError_Valid(t *testing.T) {
	testReporter := newMockTestReporter(t, "")

	beginTxError := errors.New("begin tx error")
	opts := &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}

	provider := txmocks.ExpectBeginTxAndReturnError(beginTxError, opts)(testReporter)

	tx, err := provider.BeginTx(context.Background(), opts)
	requireNil(t, tx)
	requireErrorIs(t, err, beginTxError)
}

func Test_Provider_ExpectBeginTxAndReturnError_CalledTwice(t *testing.T) {
	testReporter := newMockTestReporter(t, "unexpected call, provider.BeginTx called more than once")

	beginTxError := errors.New("begin tx error")
	opts := &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}

	provider := txmocks.ExpectBeginTxAndReturnError(beginTxError, opts)(testReporter)
	tx, err := provider.BeginTx(context.Background(), opts)
	requireNil(t, tx)
	requireErrorIs(t, err, beginTxError)

	tx, err = provider.BeginTx(context.Background(), opts)
	requireNil(t, tx)
	requireErrorIs(t, err, beginTxError)
}

func Test_Provider_ExpectBeginTxAndReturnError_Call_With_Unexpected_Opts(t *testing.T) {
	beginTxError := errors.New("begin tx error")
	expectedOpts := &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}
	callOpts := &sql.TxOptions{
		Isolation: sql.LevelDefault,
	}

	tFatalfMessage := fmt.Sprintf("unexpected call, call provider.BeginTx with %+v opts, expected %+v", callOpts, expectedOpts)

	testReporter := newMockTestReporter(t, tFatalfMessage)

	provider := txmocks.ExpectBeginTxAndReturnError(beginTxError, expectedOpts)(testReporter)

	tx, err := provider.BeginTx(context.Background(), callOpts)
	requireNil(t, tx)
	requireErrorIs(t, err, beginTxError)
}

func Test_Provider_ExpectBeginTxAndReturnError_Expect_But_Not_Called(t *testing.T) {
	testReporter := newMockTestReporter(t, "provider assertion failed, no calls occurred")

	beginTxError := errors.New("begin tx error")
	opts := &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}

	txmocks.ExpectBeginTxAndReturnError(beginTxError, opts)(testReporter)
}

func Test_Provider_ExpectBeginTxAndReturnError_CalledBegin(t *testing.T) {
	testReporter := newMockTestReporter(t, "unexpected call to provider.Begin, expect one call to provider.BeginTx")

	beginTxError := errors.New("begin tx error")
	opts := &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}

	provider := txmocks.ExpectBeginTxAndReturnError(beginTxError, opts)(testReporter)

	tx, err := provider.Begin(context.Background())
	requireNil(t, tx)
	requireNoError(t, err)
}

func Test_Provider_ExpectBeginAndReturnTx_Valid(t *testing.T) {
	testReporter := newMockTestReporter(t, "")

	provider := txmocks.ExpectBeginAndReturnTx(txmocks.NilTx)(testReporter)

	tx, err := provider.Begin(context.Background())
	requireNoError(t, err)
	requireNotNil(t, tx)
}

func Test_Provider_ExpectBeginAndReturnTx_CalledTwice(t *testing.T) {
	testReporter := newMockTestReporter(t, "unexpected call, provider.Begin called more than once")

	provider := txmocks.ExpectBeginAndReturnTx(txmocks.NilTx)(testReporter)

	tx, err := provider.Begin(context.Background())
	requireNoError(t, err)
	requireNotNil(t, tx)

	tx, err = provider.Begin(context.Background())
	requireNoError(t, err)
	requireNotNil(t, tx)
}

func Test_Provider_ExpectBeginAndReturnTx_Expect_But_Not_Called(t *testing.T) {
	testReporter := newMockTestReporter(t, "provider assertion failed, no calls occurred")

	txmocks.ExpectBeginAndReturnTx(txmocks.NilTx)(testReporter)
}

func Test_Provider_ExpectBeginAndReturnTx_CalledBeginTx(t *testing.T) {
	testReporter := newMockTestReporter(t, "unexpected call to provider.BeginTx, expect one call to provider.Begin")

	provider := txmocks.ExpectBeginAndReturnTx(txmocks.NilTx)(testReporter)

	tx, err := provider.BeginTx(context.Background(), nil)
	requireNoError(t, err)
	requireNil(t, tx)
}

func Test_Provider_ExpectBeginTxAndReturnTx_Valid(t *testing.T) {
	testReporter := newMockTestReporter(t, "")

	opts := &sql.TxOptions{Isolation: sql.LevelReadCommitted}

	provider := txmocks.ExpectBeginTxAndReturnTx(txmocks.NilTx, opts)(testReporter)

	tx, err := provider.BeginTx(context.Background(), opts)
	requireNoError(t, err)
	requireNotNil(t, tx)
}

func Test_Provider_ExpectBeginTxAndReturnTx_CalledTwice(t *testing.T) {
	testReporter := newMockTestReporter(t, "unexpected call, provider.BeginTx called more than once")

	opts := &sql.TxOptions{Isolation: sql.LevelReadCommitted}

	provider := txmocks.ExpectBeginTxAndReturnTx(txmocks.NilTx, opts)(testReporter)

	tx, err := provider.BeginTx(context.Background(), opts)
	requireNoError(t, err)
	requireNotNil(t, tx)

	tx, err = provider.BeginTx(context.Background(), opts)
	requireNoError(t, err)
	requireNotNil(t, tx)
}

func Test_Provider_ExpectBeginTxAndReturnTx_Call_With_Unexpected_Opts(t *testing.T) {
	expectedOpts := &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}
	callOpts := &sql.TxOptions{
		Isolation: sql.LevelDefault,
	}

	tFatalMessage := fmt.Sprintf("unexpected call, call provider.BeginTx with %+v opts, expected %+v", callOpts, expectedOpts)

	testReporter := newMockTestReporter(t, tFatalMessage)

	provider := txmocks.ExpectBeginTxAndReturnTx(txmocks.NilTx, expectedOpts)(testReporter)

	tx, err := provider.BeginTx(context.Background(), callOpts)
	requireNotNil(t, tx)
	requireNoError(t, err)
}

func Test_Provider_ExpectBeginTxAndReturnTx_Expect_But_Not_Called(t *testing.T) {
	testReporter := newMockTestReporter(t, "provider assertion failed, no calls occurred")

	opts := &sql.TxOptions{Isolation: sql.LevelReadCommitted}

	txmocks.ExpectBeginTxAndReturnTx(txmocks.NilTx, opts)(testReporter)
}

func Test_Provider_ExpectBeginTxAndReturnTx_CalledBegin(t *testing.T) {
	testReporter := newMockTestReporter(t, "unexpected call to provider.Begin, expect one call to provider.BeginTx")

	opts := &sql.TxOptions{Isolation: sql.LevelReadCommitted}

	provider := txmocks.ExpectBeginTxAndReturnTx(txmocks.NilTx, opts)(testReporter)

	tx, err := provider.Begin(context.Background())
	requireNil(t, tx)
	requireNoError(t, err)
}

type ProviderJoinTest struct {
	CaseName          string
	ProviderTemplates []txmocks.ProviderMock
	WithProvider      func(t *testing.T, p tx.Beginner)
	TFatalMessage     string
}

func (p *ProviderJoinTest) Name() string {
	return p.CaseName
}

func (p *ProviderJoinTest) Test(t *testing.T) {
	testReporter := newMockTestReporter(t, p.TFatalMessage)

	provider := txmocks.JoinProviders(p.ProviderTemplates...)(testReporter)

	if p.WithProvider != nil {
		p.WithProvider(t, provider)
	}
}

func runProviderJoinTests(t *testing.T, tests ...*ProviderJoinTest) {
	for _, tst := range tests {
		t.Run(tst.Name(), tst.Test)
	}
}

func Test_Provider_Join(t *testing.T) {
	errBeginTx := errors.New("begin tx")
	opts := &sql.TxOptions{
		Isolation: sql.LevelDefault,
	}

	runProviderJoinTests(t,
		&ProviderJoinTest{
			CaseName:          "zero operations",
			ProviderTemplates: nil,
			TFatalMessage:     "empty join provider templates",
		},
		&ProviderJoinTest{
			CaseName: "single operation, valid",
			ProviderTemplates: []txmocks.ProviderMock{
				txmocks.ExpectBeginAndReturnError(errBeginTx),
			},
			WithProvider: func(t *testing.T, p tx.Beginner) {
				tx, err := p.Begin(context.Background())
				requireErrorIs(t, err, errBeginTx)
				requireNil(t, tx)
			},
		},
		&ProviderJoinTest{
			CaseName: "single operation, not valid",
			ProviderTemplates: []txmocks.ProviderMock{
				txmocks.ExpectBeginAndReturnError(errBeginTx),
			},
			WithProvider: func(t *testing.T, p tx.Beginner) {
				tx, err := p.Begin(context.Background())
				requireErrorIs(t, err, errBeginTx)
				requireNil(t, tx)

				tx, err = p.Begin(context.Background())
				requireErrorIs(t, err, errBeginTx)
				requireNil(t, tx)
			},
			TFatalMessage: "unexpected call, provider.Begin called more than once",
		},
		&ProviderJoinTest{
			CaseName: "two operations, valid",
			ProviderTemplates: []txmocks.ProviderMock{
				txmocks.ExpectBeginAndReturnError(errBeginTx),
				txmocks.ExpectBeginAndReturnError(errBeginTx),
			},
			WithProvider: func(t *testing.T, p tx.Beginner) {
				tx, err := p.Begin(context.Background())
				requireErrorIs(t, err, errBeginTx)
				requireNil(t, tx)

				tx, err = p.Begin(context.Background())
				requireErrorIs(t, err, errBeginTx)
				requireNil(t, tx)
			},
		},
		&ProviderJoinTest{
			CaseName: "two operations, invalid count times",
			ProviderTemplates: []txmocks.ProviderMock{
				txmocks.ExpectBeginAndReturnError(errBeginTx),
				txmocks.ExpectBeginAndReturnError(errBeginTx),
			},
			WithProvider: func(t *testing.T, p tx.Beginner) {
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
			TFatalMessage: "unexpected call to provider.Begin, no calls left",
		},
		&ProviderJoinTest{
			CaseName: "two operations, valid order",
			ProviderTemplates: []txmocks.ProviderMock{
				txmocks.ExpectBeginAndReturnError(errBeginTx),
				txmocks.ExpectBeginTxAndReturnError(errBeginTx, opts),
			},
			WithProvider: func(t *testing.T, p tx.Beginner) {
				tx, err := p.Begin(context.Background())
				requireErrorIs(t, err, errBeginTx)
				requireNil(t, tx)

				tx, err = p.BeginTx(context.Background(), opts)
				requireErrorIs(t, err, errBeginTx)
				requireNil(t, tx)
			},
		},
		&ProviderJoinTest{
			CaseName: "two operations, invalid order",
			ProviderTemplates: []txmocks.ProviderMock{
				txmocks.ExpectBeginAndReturnError(errBeginTx),
				txmocks.ExpectBeginTxAndReturnError(errBeginTx, opts),
			},
			WithProvider: func(t *testing.T, p tx.Beginner) {
				tx, err := p.BeginTx(context.Background(), opts)
				requireNil(t, err)
				requireNil(t, tx)

				tx, err = p.Begin(context.Background())
				requireNil(t, err)
				requireNil(t, tx)
			},
			TFatalMessage: "unexpected call to provider.BeginTx, expect one call to provider.Begin",
		},
		&ProviderJoinTest{
			CaseName: "many operations, with transactions",
			ProviderTemplates: []txmocks.ProviderMock{
				txmocks.ExpectBeginAndReturnError(errBeginTx),
				txmocks.ExpectBeginTxAndReturnError(errBeginTx, opts),
				txmocks.ExpectBeginAndReturnTx(txmocks.ExpectCommit),
				txmocks.ExpectBeginTxAndReturnTx(txmocks.ExpectRollback(nil), opts),
			},
			WithProvider: func(t *testing.T, p tx.Beginner) {
				tx, err := p.Begin(context.Background())
				requireErrorIs(t, err, errBeginTx)
				requireNil(t, tx)

				tx, err = p.BeginTx(context.Background(), opts)
				requireErrorIs(t, err, errBeginTx)
				requireNil(t, tx)

				tx, err = p.Begin(context.Background())
				requireNoError(t, err)
				requireNotNil(t, tx)

				err = tx.Commit()
				requireNoError(t, err)

				tx, err = p.BeginTx(context.Background(), opts)
				requireNoError(t, err)
				requireNotNil(t, tx)

				err = tx.Rollback()
				requireNoError(t, err)
			},
		},
	)
}
