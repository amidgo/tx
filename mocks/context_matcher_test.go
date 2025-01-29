package transactionmocks_test

import (
	"testing"

	mocks "github.com/amidgo/transaction/mocks"
	"github.com/stretchr/testify/require"
)

func Test_ContextMatcher(t *testing.T) {
	tx := mocks.ExpectNothing(t)
	ctx := tx.Context()

	enabledMatcher := mocks.TxEnabled()
	disabledMatcher := mocks.TxDisabled()

	require.True(t, enabledMatcher.Matches(ctx))
	require.False(t, disabledMatcher.Matches(ctx))
}

func Test_Context_Disabled_After_Rollback(t *testing.T) {
	tx := mocks.ExpectRollback(nil)(t)

	require.True(t, mocks.TxEnabled().Matches(tx.Context()))
	require.False(t, mocks.TxDisabled().Matches(tx.Context()))

	err := tx.Rollback()
	require.NoError(t, err)

	require.False(t, mocks.TxEnabled().Matches(tx.Context()))
	require.True(t, mocks.TxDisabled().Matches(tx.Context()))
}

func Test_Context_Disabled_After_Commit(t *testing.T) {
	tx := mocks.ExpectCommit(t)

	require.True(t, mocks.TxEnabled().Matches(tx.Context()))
	require.False(t, mocks.TxDisabled().Matches(tx.Context()))

	err := tx.Commit()
	require.NoError(t, err)

	require.False(t, mocks.TxEnabled().Matches(tx.Context()))
	require.True(t, mocks.TxDisabled().Matches(tx.Context()))
}
