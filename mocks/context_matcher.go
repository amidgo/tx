package txmocks

import (
	"context"
	"strconv"
)

type Matcher interface {
	Matches(x any) bool
	String() string
}

func TxEnabled() Matcher {
	return &txMatcher{
		enabled: true,
	}
}

func TxDisabled() Matcher {
	return &txMatcher{
		enabled: false,
	}
}

type txMatcher struct {
	enabled bool
}

func (t txMatcher) Matches(x any) bool {
	ctx, ok := x.(context.Context)
	if !ok {
		return false
	}

	return txEnabled(ctx) == t.enabled
}

func (t txMatcher) String() string {
	return "context should contain txEnabled{} flag that equal " + strconv.FormatBool(t.enabled)
}
