package mocks_test

import (
	"context"
	"testing"

	"github.com/amidgo/transaction"
	"github.com/amidgo/transaction/mocks"
	"github.com/stretchr/testify/require"
)

func Test_ContextMatcher(t *testing.T) {
	ctx := transaction.StartTx(context.Background())

	require.True(t, mocks.TxEnabled.Matches(ctx))
	require.False(t, mocks.TxDisabled.Matches(ctx))
}
