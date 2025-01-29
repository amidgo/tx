package mocks_test

import (
	"testing"

	"github.com/amidgo/transaction/mocks"
)

func Test_ContextMatcher(t *testing.T) {
	tx := mocks.ExpectNothing(t)
	ctx := tx.Context()

	enabledMatcher := mocks.TxEnabled()
	disabledMatcher := mocks.TxDisabled()

	requireTrue(t, enabledMatcher.Matches(ctx))
	requireFalse(t, disabledMatcher.Matches(ctx))
}

func Test_Context_Disabled_After_Rollback(t *testing.T) {
	tx := mocks.ExpectRollback(nil)(t)

	requireTrue(t, mocks.TxEnabled().Matches(tx.Context()))
	requireFalse(t, mocks.TxDisabled().Matches(tx.Context()))

	err := tx.Rollback()
	requireNoError(t, err)

	requireFalse(t, mocks.TxEnabled().Matches(tx.Context()))
	requireTrue(t, mocks.TxDisabled().Matches(tx.Context()))
}

func Test_Context_Disabled_After_Commit(t *testing.T) {
	tx := mocks.ExpectCommit(t)

	requireTrue(t, mocks.TxEnabled().Matches(tx.Context()))
	requireFalse(t, mocks.TxDisabled().Matches(tx.Context()))

	err := tx.Commit()
	requireNoError(t, err)

	requireFalse(t, mocks.TxEnabled().Matches(tx.Context()))
	requireTrue(t, mocks.TxDisabled().Matches(tx.Context()))
}
