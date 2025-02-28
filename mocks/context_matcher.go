package txmocks

import (
	"context"
	"strconv"

	"github.com/amidgo/tx"
)

type Matcher interface {
	Matches(x any) bool
	String() string
}

func TxEnabled() Matcher {
	return &txMatcher{
		beginner: &Beginner{},
		enabled:  true,
	}
}

func TxDisabled() Matcher {
	return &txMatcher{
		beginner: &Beginner{},
		enabled:  false,
	}
}

type txMatcher struct {
	beginner tx.Beginner
	enabled  bool
}

func (t txMatcher) Matches(x any) bool {
	ctx, ok := x.(context.Context)
	if !ok {
		return false
	}

	return t.beginner.TxEnabled(ctx) == t.enabled
}

func (t txMatcher) String() string {
	return "context should contain txEnabled{} flag that equal " + strconv.FormatBool(t.enabled)
}
