package mocks

type TransactionMock func(t testReporter) *Transaction

func ExpectRollback(err error) TransactionMock {
	return func(t testReporter) *Transaction {
		asrt := &rollback{
			err: err,
			t:   t,
		}

		return newTransaction(t, asrt)
	}
}

func ExpectCommit(t testReporter) *Transaction {
	asrt := &commit{
		t: t,
	}

	return newTransaction(t, asrt)
}

func ExpectRollbackAfterFailedCommit(commitError error) TransactionMock {
	return func(t testReporter) *Transaction {
		asrt := &rollbackAfterFailedCommit{
			t:   t,
			err: commitError,
		}

		return newTransaction(t, asrt)
	}
}

func ExpectNothing(t testReporter) *Transaction {
	asrt := &nothing{
		t: t,
	}

	return newTransaction(t, asrt)
}
