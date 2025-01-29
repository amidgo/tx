package mocks_test

import (
	"errors"
	"testing"

	"github.com/amidgo/transaction/mocks"
	"github.com/stretchr/testify/require"
)

func Test_Transaction_Context(t *testing.T) {
	testReporter := newMockTestReporter(t, "")

	tx := mocks.ExpectNothing(testReporter)

	ctx := tx.Context()

	require.True(t, mocks.TxEnabled().Matches(ctx))
	require.False(t, mocks.TxDisabled().Matches(ctx))
}

func Test_Transaction_Commit_Valid(t *testing.T) {
	testReporter := newMockTestReporter(t, "")

	tx := mocks.ExpectCommit(testReporter)

	err := tx.Commit()
	require.NoError(t, err)
}

func Test_Transaction_Commit_CalledTwice(t *testing.T) {
	testReporter := newMockTestReporter(t, "unexpected call, tx.Commit called more than once")

	tx := mocks.ExpectCommit(testReporter)

	err := tx.Commit()
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)
}

func Test_Transaction_Commit_CalledRollback(t *testing.T) {
	testReporter := newMockTestReporter(t, "unexpected call to tx.Rollback, expected one call to tx.Commit")

	tx := mocks.ExpectCommit(testReporter)

	tx.Rollback()
}

func Test_Transaction_ExpectCommit_Expect_But_Not_Called(t *testing.T) {
	testReporter := newMockTestReporter(t, "tx assertion failed, no calls occurred")

	mocks.ExpectCommit(testReporter)
}

func Test_Transaction_ExpectRollback_Valid(t *testing.T) {
	testReporter := newMockTestReporter(t, "")

	errRollback := errors.New("rollback error")

	tx := mocks.ExpectRollback(errRollback)(testReporter)

	err := tx.Rollback()
	require.ErrorIs(t, err, errRollback)
}

func Test_Transaction_ExpectRollback_CalledTwice(t *testing.T) {
	testReporter := newMockTestReporter(t, "unexpected call, tx.Rollback called more than once")

	errRollback := errors.New("rollback error")

	tx := mocks.ExpectRollback(errRollback)(testReporter)

	err := tx.Rollback()
	require.ErrorIs(t, err, errRollback)

	err = tx.Rollback()
	require.ErrorIs(t, err, errRollback)
}

func Test_Transaction_ExpectRollback_CalledCommit(t *testing.T) {
	testReporter := newMockTestReporter(t, "unexpected call to tx.Commit, expected one call to tx.Rollback")

	errRollback := errors.New("rollback error")

	tx := mocks.ExpectRollback(errRollback)(testReporter)

	err := tx.Commit()
	require.NoError(t, err)
}

func Test_Transaction_ExpectRollback_Expected_But_Not_Called(t *testing.T) {
	testReporter := newMockTestReporter(t, "tx assertion failed, no calls occurred")

	errRollback := errors.New("rollback error")

	mocks.ExpectRollback(errRollback)(testReporter)
}

func Test_Transaction_ExpectRollbackAfterFailedCommit_Valid(t *testing.T) {
	testReporter := newMockTestReporter(t, "")

	errCommit := errors.New("failed commit")

	tx := mocks.ExpectRollbackAfterFailedCommit(errCommit)(testReporter)

	err := tx.Commit()
	require.ErrorIs(t, err, errCommit)

	err = tx.Rollback()
	require.NoError(t, err)
}

func Test_Transaction_ExpectRollbackAfterFailedCommit_RollbackFirst(t *testing.T) {
	testReporter := newMockTestReporter(t, "unexpected call, tx.Commit has not been called yet or tx.Rollback has been already called")

	errCommit := errors.New("failed commit")

	tx := mocks.ExpectRollbackAfterFailedCommit(errCommit)(testReporter)

	err := tx.Rollback()
	require.NoError(t, err)

	err = tx.Commit()
	require.ErrorIs(t, err, errCommit)
}

func Test_Transaction_ExpectRollbackAfterFailedCommit_OnlyCommit(t *testing.T) {
	testReporter := newMockTestReporter(t, "tx assertion failed, tx.Rollback not called")

	errCommit := errors.New("failed commit")

	tx := mocks.ExpectRollbackAfterFailedCommit(errCommit)(testReporter)

	err := tx.Commit()
	require.ErrorIs(t, err, errCommit)
}

func Test_Transaction_ExpectRollbackAfterFailedCommit_CommitCalledTwice(t *testing.T) {
	testReporter := newMockTestReporter(t, "unexpected call, tx.Commit has already was called, expect call tx.Rollback")

	errCommit := errors.New("failed commit")

	tx := mocks.ExpectRollbackAfterFailedCommit(errCommit)(testReporter)

	err := tx.Commit()
	require.ErrorIs(t, err, errCommit)

	err = tx.Commit()
	require.ErrorIs(t, err, errCommit)

	tx.Rollback()
}

func Test_Transaction_ExpectRollbackAfterFailedCommit_RollbackCalledTwice(t *testing.T) {
	testReporter := newMockTestReporter(t, "unexpected call, tx.Commit has not been called yet or tx.Rollback has been already called")

	errCommit := errors.New("failed commit")

	tx := mocks.ExpectRollbackAfterFailedCommit(errCommit)(testReporter)

	err := tx.Commit()
	require.ErrorIs(t, err, errCommit)

	tx.Rollback()
	tx.Rollback()
}
