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

func Test_Beginner_ExpectBeginAndReturnError_Valid(t *testing.T) {
	testReporter := newMockTestReporter(t, "")

	beginError := errors.New("begin error")

	beginner := txmocks.ExpectBeginAndReturnError(beginError)(testReporter)

	tx, err := beginner.Begin(context.Background())
	requireNil(t, tx)
	requireErrorIs(t, err, beginError)
}

func Test_Beginner_ExpectBeginAndReturnError_CalledTwice(t *testing.T) {
	testReporter := newMockTestReporter(t, "unexpected call, beginner.Begin called more than once")

	beginError := errors.New("begin error")

	beginner := txmocks.ExpectBeginAndReturnError(beginError)(testReporter)

	tx, err := beginner.Begin(context.Background())
	requireNil(t, tx)
	requireErrorIs(t, err, beginError)

	tx, err = beginner.Begin(context.Background())
	requireNil(t, tx)
	requireErrorIs(t, err, beginError)
}

func Test_Beginner_ExpectBeginAndReturnError_Expect_But_Not_Called(t *testing.T) {
	testReporter := newMockTestReporter(t, "beginner assertion failed, no calls occurred")

	beginError := errors.New("begin error")

	txmocks.ExpectBeginAndReturnError(beginError)(testReporter)
}

func Test_Beginner_ExpectBeginAndReturnError_CallBeginTx(t *testing.T) {
	testReporter := newMockTestReporter(t, "unexpected call to beginner.BeginTx, expect one call to beginner.Begin")

	beginError := errors.New("begin error")

	beginner := txmocks.ExpectBeginAndReturnError(beginError)(testReporter)

	tx, err := beginner.BeginTx(context.Background(), nil)
	requireNil(t, tx)
	requireNoError(t, err)
}

func Test_Beginner_ExpectBeginTxAndReturnError_Valid(t *testing.T) {
	testReporter := newMockTestReporter(t, "")

	beginTxError := errors.New("begin tx error")
	opts := &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}

	beginner := txmocks.ExpectBeginTxAndReturnError(beginTxError, opts)(testReporter)

	tx, err := beginner.BeginTx(context.Background(), opts)
	requireNil(t, tx)
	requireErrorIs(t, err, beginTxError)
}

func Test_Beginner_ExpectBeginTxAndReturnError_CalledTwice(t *testing.T) {
	testReporter := newMockTestReporter(t, "unexpected call, beginner.BeginTx called more than once")

	beginTxError := errors.New("begin tx error")
	opts := &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}

	beginner := txmocks.ExpectBeginTxAndReturnError(beginTxError, opts)(testReporter)
	tx, err := beginner.BeginTx(context.Background(), opts)
	requireNil(t, tx)
	requireErrorIs(t, err, beginTxError)

	tx, err = beginner.BeginTx(context.Background(), opts)
	requireNil(t, tx)
	requireErrorIs(t, err, beginTxError)
}

func Test_Beginner_ExpectBeginTxAndReturnError_Call_With_Unexpected_Opts(t *testing.T) {
	beginTxError := errors.New("begin tx error")
	expectedOpts := &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}
	callOpts := &sql.TxOptions{
		Isolation: sql.LevelDefault,
	}

	tFatalfMessage := fmt.Sprintf("unexpected call, call beginner.BeginTx with %+v opts, expected %+v", callOpts, expectedOpts)

	testReporter := newMockTestReporter(t, tFatalfMessage)

	beginner := txmocks.ExpectBeginTxAndReturnError(beginTxError, expectedOpts)(testReporter)

	tx, err := beginner.BeginTx(context.Background(), callOpts)
	requireNil(t, tx)
	requireErrorIs(t, err, beginTxError)
}

func Test_Beginner_ExpectBeginTxAndReturnError_Expect_But_Not_Called(t *testing.T) {
	testReporter := newMockTestReporter(t, "beginner assertion failed, no calls occurred")

	beginTxError := errors.New("begin tx error")
	opts := &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}

	txmocks.ExpectBeginTxAndReturnError(beginTxError, opts)(testReporter)
}

func Test_Beginner_ExpectBeginTxAndReturnError_CalledBegin(t *testing.T) {
	testReporter := newMockTestReporter(t, "unexpected call to beginner.Begin, expect one call to beginner.BeginTx")

	beginTxError := errors.New("begin tx error")
	opts := &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}

	beginner := txmocks.ExpectBeginTxAndReturnError(beginTxError, opts)(testReporter)

	tx, err := beginner.Begin(context.Background())
	requireNil(t, tx)
	requireNoError(t, err)
}

func Test_Beginner_ExpectBeginAndReturnTx_Valid(t *testing.T) {
	testReporter := newMockTestReporter(t, "")

	beginner := txmocks.ExpectBeginAndReturnTx(txmocks.NilTx)(testReporter)

	tx, err := beginner.Begin(context.Background())
	requireNoError(t, err)
	requireNotNil(t, tx)
}

func Test_Beginner_ExpectBeginAndReturnTx_CalledTwice(t *testing.T) {
	testReporter := newMockTestReporter(t, "unexpected call, beginner.Begin called more than once")

	beginner := txmocks.ExpectBeginAndReturnTx(txmocks.NilTx)(testReporter)

	tx, err := beginner.Begin(context.Background())
	requireNoError(t, err)
	requireNotNil(t, tx)

	tx, err = beginner.Begin(context.Background())
	requireNoError(t, err)
	requireNotNil(t, tx)
}

func Test_Beginner_ExpectBeginAndReturnTx_Expect_But_Not_Called(t *testing.T) {
	testReporter := newMockTestReporter(t, "beginner assertion failed, no calls occurred")

	txmocks.ExpectBeginAndReturnTx(txmocks.NilTx)(testReporter)
}

func Test_Beginner_ExpectBeginAndReturnTx_CalledBeginTx(t *testing.T) {
	testReporter := newMockTestReporter(t, "unexpected call to beginner.BeginTx, expect one call to beginner.Begin")

	beginner := txmocks.ExpectBeginAndReturnTx(txmocks.NilTx)(testReporter)

	tx, err := beginner.BeginTx(context.Background(), nil)
	requireNoError(t, err)
	requireNil(t, tx)
}

func Test_Beginner_ExpectBeginTxAndReturnTx_Valid(t *testing.T) {
	testReporter := newMockTestReporter(t, "")

	opts := &sql.TxOptions{Isolation: sql.LevelReadCommitted}

	beginner := txmocks.ExpectBeginTxAndReturnTx(txmocks.NilTx, opts)(testReporter)

	tx, err := beginner.BeginTx(context.Background(), opts)
	requireNoError(t, err)
	requireNotNil(t, tx)
}

func Test_Beginner_ExpectBeginTxAndReturnTx_CalledTwice(t *testing.T) {
	testReporter := newMockTestReporter(t, "unexpected call, beginner.BeginTx called more than once")

	opts := &sql.TxOptions{Isolation: sql.LevelReadCommitted}

	beginner := txmocks.ExpectBeginTxAndReturnTx(txmocks.NilTx, opts)(testReporter)

	tx, err := beginner.BeginTx(context.Background(), opts)
	requireNoError(t, err)
	requireNotNil(t, tx)

	tx, err = beginner.BeginTx(context.Background(), opts)
	requireNoError(t, err)
	requireNotNil(t, tx)
}

func Test_Beginner_ExpectBeginTxAndReturnTx_Call_With_Unexpected_Opts(t *testing.T) {
	expectedOpts := &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}
	callOpts := &sql.TxOptions{
		Isolation: sql.LevelDefault,
	}

	tFatalMessage := fmt.Sprintf("unexpected call, call beginner.BeginTx with %+v opts, expected %+v", callOpts, expectedOpts)

	testReporter := newMockTestReporter(t, tFatalMessage)

	beginner := txmocks.ExpectBeginTxAndReturnTx(txmocks.NilTx, expectedOpts)(testReporter)

	tx, err := beginner.BeginTx(context.Background(), callOpts)
	requireNotNil(t, tx)
	requireNoError(t, err)
}

func Test_Beginner_ExpectBeginTxAndReturnTx_Expect_But_Not_Called(t *testing.T) {
	testReporter := newMockTestReporter(t, "beginner assertion failed, no calls occurred")

	opts := &sql.TxOptions{Isolation: sql.LevelReadCommitted}

	txmocks.ExpectBeginTxAndReturnTx(txmocks.NilTx, opts)(testReporter)
}

func Test_Beginner_ExpectBeginTxAndReturnTx_CalledBegin(t *testing.T) {
	testReporter := newMockTestReporter(t, "unexpected call to beginner.Begin, expect one call to beginner.BeginTx")

	opts := &sql.TxOptions{Isolation: sql.LevelReadCommitted}

	beginner := txmocks.ExpectBeginTxAndReturnTx(txmocks.NilTx, opts)(testReporter)

	tx, err := beginner.Begin(context.Background())
	requireNil(t, tx)
	requireNoError(t, err)
}

func Test_Beginner_ExpectNothing_CalledBegin(t *testing.T) {
	testReporter := newMockTestReporter(t, "unexpected call to beginner.Begin")

	beginner := txmocks.ExpectNothing()(testReporter)

	tx, err := beginner.Begin(context.Background())
	requireNoError(t, err)
	requireNil(t, tx)
}

func Test_Beginner_ExpectNothing_CalledBeginTx(t *testing.T) {
	testReporter := newMockTestReporter(t, "unexpected call to beginner.BeginTx")

	beginner := txmocks.ExpectNothing()(testReporter)

	tx, err := beginner.BeginTx(context.Background(), nil)
	requireNoError(t, err)
	requireNil(t, tx)
}

type BeginnerJoinTest struct {
	CaseName          string
	BeginnerTemplates []txmocks.BeginnerMock
	WithBeginner      func(t *testing.T, p tx.Beginner)
	TFatalMessage     string
}

func (p *BeginnerJoinTest) Name() string {
	return p.CaseName
}

func (p *BeginnerJoinTest) Test(t *testing.T) {
	testReporter := newMockTestReporter(t, p.TFatalMessage)

	beginner := txmocks.JoinBeginners(p.BeginnerTemplates...)(testReporter)

	if p.WithBeginner != nil {
		p.WithBeginner(t, beginner)
	}
}

func runBeginnerJoinTests(t *testing.T, tests ...*BeginnerJoinTest) {
	for _, tst := range tests {
		t.Run(tst.Name(), tst.Test)
	}
}

func Test_Beginner_Join(t *testing.T) {
	errBeginTx := errors.New("begin tx")
	opts := &sql.TxOptions{
		Isolation: sql.LevelDefault,
	}

	runBeginnerJoinTests(t,
		&BeginnerJoinTest{
			CaseName:          "zero operations",
			BeginnerTemplates: nil,
			TFatalMessage:     "empty join beginner templates",
		},
		&BeginnerJoinTest{
			CaseName: "single operation, valid",
			BeginnerTemplates: []txmocks.BeginnerMock{
				txmocks.ExpectBeginAndReturnError(errBeginTx),
			},
			WithBeginner: func(t *testing.T, p tx.Beginner) {
				tx, err := p.Begin(context.Background())
				requireErrorIs(t, err, errBeginTx)
				requireNil(t, tx)
			},
		},
		&BeginnerJoinTest{
			CaseName: "single operation, not valid",
			BeginnerTemplates: []txmocks.BeginnerMock{
				txmocks.ExpectBeginAndReturnError(errBeginTx),
			},
			WithBeginner: func(t *testing.T, p tx.Beginner) {
				tx, err := p.Begin(context.Background())
				requireErrorIs(t, err, errBeginTx)
				requireNil(t, tx)

				tx, err = p.Begin(context.Background())
				requireErrorIs(t, err, errBeginTx)
				requireNil(t, tx)
			},
			TFatalMessage: "unexpected call, beginner.Begin called more than once",
		},
		&BeginnerJoinTest{
			CaseName: "two operations, valid",
			BeginnerTemplates: []txmocks.BeginnerMock{
				txmocks.ExpectBeginAndReturnError(errBeginTx),
				txmocks.ExpectBeginAndReturnError(errBeginTx),
			},
			WithBeginner: func(t *testing.T, p tx.Beginner) {
				tx, err := p.Begin(context.Background())
				requireErrorIs(t, err, errBeginTx)
				requireNil(t, tx)

				tx, err = p.Begin(context.Background())
				requireErrorIs(t, err, errBeginTx)
				requireNil(t, tx)
			},
		},
		&BeginnerJoinTest{
			CaseName: "two operations, invalid count times",
			BeginnerTemplates: []txmocks.BeginnerMock{
				txmocks.ExpectBeginAndReturnError(errBeginTx),
				txmocks.ExpectBeginAndReturnError(errBeginTx),
			},
			WithBeginner: func(t *testing.T, p tx.Beginner) {
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
			TFatalMessage: "unexpected call to beginner.Begin, no calls left",
		},
		&BeginnerJoinTest{
			CaseName: "two operations, valid order",
			BeginnerTemplates: []txmocks.BeginnerMock{
				txmocks.ExpectBeginAndReturnError(errBeginTx),
				txmocks.ExpectBeginTxAndReturnError(errBeginTx, opts),
			},
			WithBeginner: func(t *testing.T, p tx.Beginner) {
				tx, err := p.Begin(context.Background())
				requireErrorIs(t, err, errBeginTx)
				requireNil(t, tx)

				tx, err = p.BeginTx(context.Background(), opts)
				requireErrorIs(t, err, errBeginTx)
				requireNil(t, tx)
			},
		},
		&BeginnerJoinTest{
			CaseName: "two operations, invalid order",
			BeginnerTemplates: []txmocks.BeginnerMock{
				txmocks.ExpectBeginAndReturnError(errBeginTx),
				txmocks.ExpectBeginTxAndReturnError(errBeginTx, opts),
			},
			WithBeginner: func(t *testing.T, p tx.Beginner) {
				tx, err := p.BeginTx(context.Background(), opts)
				requireNil(t, err)
				requireNil(t, tx)

				tx, err = p.Begin(context.Background())
				requireNil(t, err)
				requireNil(t, tx)
			},
			TFatalMessage: "unexpected call to beginner.BeginTx, expect one call to beginner.Begin",
		},
		&BeginnerJoinTest{
			CaseName: "many operations, with transactions",
			BeginnerTemplates: []txmocks.BeginnerMock{
				txmocks.ExpectBeginAndReturnError(errBeginTx),
				txmocks.ExpectBeginTxAndReturnError(errBeginTx, opts),
				txmocks.ExpectBeginAndReturnTx(txmocks.ExpectCommit),
				txmocks.ExpectBeginTxAndReturnTx(txmocks.ExpectRollback(nil), opts),
			},
			WithBeginner: func(t *testing.T, p tx.Beginner) {
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
