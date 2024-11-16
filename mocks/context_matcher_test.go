package mocks_test

import (
	"context"
	"testing"

	"github.com/amidgo/transaction"
	"github.com/amidgo/transaction/mocks"
)

func Test_ContextMatcher(t *testing.T) {
	ctx := transaction.StartTx(context.Background())

	requireTrue(t, mocks.TxEnabled.Matches(ctx))
	requireFalse(t, mocks.TxDisabled.Matches(ctx))
}
