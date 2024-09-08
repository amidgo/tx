package mocks

import (
	"context"
	"strconv"

	"github.com/amidgo/transaction"
)

const (
	TxEnabled  TxEnabledContextMatcher = true
	TxDisabled TxEnabledContextMatcher = false
)

type TxEnabledContextMatcher bool

func (m TxEnabledContextMatcher) Matches(x any) bool {
	ctx, ok := x.(context.Context)
	if !ok {
		return false
	}

	return transaction.TxEnabled(ctx) == bool(m)
}

func (m TxEnabledContextMatcher) String() string {
	return "context should contain txEnabled{} flag that equal " + strconv.FormatBool(bool(m))
}
