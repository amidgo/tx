package transactionmocks

import (
	"context"
	"strconv"

	"github.com/amidgo/transaction"
)

type Matcher interface {
	Matches(x any) bool
	String() string
}

func TxEnabled() Matcher {
	return &txMatcher{
		provider: &Provider{},
		enabled:  true,
	}
}

func TxDisabled() Matcher {
	return &txMatcher{
		provider: &Provider{},
		enabled:  false,
	}
}

type txMatcher struct {
	provider transaction.Provider
	enabled  bool
}

func (t txMatcher) Matches(x any) bool {
	ctx, ok := x.(context.Context)
	if !ok {
		return false
	}

	return t.provider.TxEnabled(ctx) == t.enabled
}

func (t txMatcher) String() string {
	return "context should contain txEnabled{} flag that equal " + strconv.FormatBool(t.enabled)
}
