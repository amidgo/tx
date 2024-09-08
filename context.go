package transaction

import context "context"

type txEnabled struct{}

func TxEnabled(txContext context.Context) bool {
	enabled, ok := txContext.Value(txEnabled{}).(bool)
	if !ok {
		return false
	}

	return enabled
}

func StartTx(ctx context.Context) context.Context {
	return context.WithValue(ctx, txEnabled{}, true)
}

func ClearTx(ctx context.Context) context.Context {
	return context.WithValue(ctx, txEnabled{}, false)
}
