package txmocks_test

import (
	"testing"

	txmocks "github.com/amidgo/tx/mocks"
)

func Test_ContextMatcher(t *testing.T) {
	tx := txmocks.NilTx(t)
	ctx := tx.Context()

	enabledMatcher := txmocks.TxEnabled()
	disabledMatcher := txmocks.TxDisabled()

	requireTrue(t, enabledMatcher.Matches(ctx))
	requireFalse(t, disabledMatcher.Matches(ctx))
}

func Test_Context_Disabled_After_Rollback(t *testing.T) {
	tx := txmocks.ExpectRollback(nil)(t)

	requireTrue(t, txmocks.TxEnabled().Matches(tx.Context()))
	requireFalse(t, txmocks.TxDisabled().Matches(tx.Context()))

	err := tx.Rollback()
	requireNoError(t, err)

	requireFalse(t, txmocks.TxEnabled().Matches(tx.Context()))
	requireTrue(t, txmocks.TxDisabled().Matches(tx.Context()))
}

func Test_Context_Disabled_After_Commit(t *testing.T) {
	tx := txmocks.ExpectCommit(t)

	requireTrue(t, txmocks.TxEnabled().Matches(tx.Context()))
	requireFalse(t, txmocks.TxDisabled().Matches(tx.Context()))

	err := tx.Commit()
	requireNoError(t, err)

	requireFalse(t, txmocks.TxEnabled().Matches(tx.Context()))
	requireTrue(t, txmocks.TxDisabled().Matches(tx.Context()))
}
