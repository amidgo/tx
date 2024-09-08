package mocks

import "github.com/amidgo/transaction"

type TransactionTemplate func(t testReporter) transaction.Transaction

func ExpectRollback(err error) TransactionTemplate {
	return func(t testReporter) transaction.Transaction {
		asrt := &rollback{
			err: err,
			t:   t,
		}

		return newTransaction(t, asrt)
	}
}

func ExpectCommit() TransactionTemplate {
	return func(t testReporter) transaction.Transaction {
		asrt := &commit{
			t: t,
		}

		return newTransaction(t, asrt)
	}
}

func ExpectRollbackAfterFailedCommit(commitError error) TransactionTemplate {
	return func(t testReporter) transaction.Transaction {
		asrt := &rollbackAfterFailedCommit{
			t:   t,
			err: commitError,
		}

		return newTransaction(t, asrt)
	}
}

func ExpectNothing() TransactionTemplate {
	return func(t testReporter) transaction.Transaction {
		asrt := &nothing{
			t: t,
		}

		return newTransaction(t, asrt)
	}
}
