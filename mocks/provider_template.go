package mocks

import (
	"database/sql"
)

type ProviderTemplate func(t testReporter) *Provider

func ExpectBeginAndReturnError(beginError error) ProviderTemplate {
	return func(t testReporter) *Provider {
		provider := NewProvider(t)

		provider.ExpectBeginAndReturnError(beginError)

		return provider
	}
}

func ExpectBeginTxAndReturnError(beginError error, expectedOpts sql.TxOptions) ProviderTemplate {
	return func(t testReporter) *Provider {
		provider := NewProvider(t)

		provider.ExpectBeginTxAndReturnError(beginError, expectedOpts)

		return provider
	}
}

func ExpectBeginAndReturnTx(tx TransactionTemplate) ProviderTemplate {
	return func(t testReporter) *Provider {
		provider := NewProvider(t)

		provider.ExpectBeginAndReturnTx(tx(t))

		return provider
	}
}

func ExpectBeginTxAndReturnTx(tx TransactionTemplate, opts sql.TxOptions) ProviderTemplate {
	return func(t testReporter) *Provider {
		provider := NewProvider(t)

		provider.ExpectBeginTxAndReturnTx(tx(t), opts)

		return provider
	}
}
