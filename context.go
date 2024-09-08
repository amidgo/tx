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

func MockEnableTxContext(ctx context.Context) context.Context {
	return startTx(ctx)
}

func startTx(ctx context.Context) context.Context {
	return context.WithValue(ctx, txEnabled{}, true)
}

func clearTx(ctx context.Context) context.Context {
	return context.WithValue(ctx, txEnabled{}, false)
}
