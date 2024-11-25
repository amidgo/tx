package mocks

import "github.com/amidgo/transaction"

type TransactionMock func(t testReporter) transaction.Transaction

func ExpectRollback(err error) TransactionMock {
	return func(t testReporter) transaction.Transaction {
		asrt := &rollback{
			err: err,
			t:   t,
		}

		return newTransaction(t, asrt)
	}
}

func ExpectCommit() TransactionMock {
	return func(t testReporter) transaction.Transaction {
		asrt := &commit{
			t: t,
		}

		return newTransaction(t, asrt)
	}
}

func ExpectRollbackAfterFailedCommit(commitError error) TransactionMock {
	return func(t testReporter) transaction.Transaction {
		asrt := &rollbackAfterFailedCommit{
			t:   t,
			err: commitError,
		}

		return newTransaction(t, asrt)
	}
}

func ExpectNothing() TransactionMock {
	return func(t testReporter) transaction.Transaction {
		asrt := &nothing{
			t: t,
		}

		return newTransaction(t, asrt)
	}
}
