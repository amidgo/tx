package transaction_test

import (
	"context"
	"database/sql"
	"testing"

	postgrescontainer "github.com/amidgo/containers/postgres"
	"github.com/amidgo/transaction"
	stdlibtransaction "github.com/amidgo/transaction/stdlib"
	"github.com/stretchr/testify/require"
)

func Test_SQLProvider_Begin_BeginTx(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db := postgrescontainer.RunForTesting(t, postgrescontainer.EmptyMigrations{})

	provider := stdlibtransaction.NewProvider(db)

	exec := provider.Executor(ctx)
	_, ok := exec.(*sql.DB)
	require.True(t, ok)

	tx, err := provider.BeginTx(ctx, sql.TxOptions{Isolation: sql.LevelSerializable, ReadOnly: false})
	require.NoError(t, err)

	assertSQLTransactionEnabled(t, provider, tx, "serializable", false)

	tx, err = provider.BeginTx(ctx, sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	require.NoError(t, err)

	assertSQLTransactionEnabled(t, provider, tx, "repeatable read", true)

	tx, err = provider.Begin(ctx)
	require.NoError(t, err)

	assertSQLTransactionEnabled(t, provider, tx, "read committed", false)
}

func assertSQLTransactionEnabled(t *testing.T, provider *stdlibtransaction.Provider, tx transaction.Transaction, expectedIsolationLevel string, readOnly bool) {
	enabled := provider.TxEnabled(tx.Context())
	require.True(t, enabled)

	exec := provider.Executor(tx.Context())
	_, ok := exec.(*sql.Tx)
	require.True(t, ok)

	assertSQLTransactionLevel(t, exec, expectedIsolationLevel, readOnly)

	err := tx.Rollback(tx.Context())
	require.NoError(t, err)

	enabled = provider.TxEnabled(tx.Context())
	require.False(t, enabled)
}

func assertSQLTransactionLevel(t *testing.T, exec executor, expectedIsolationLevel string, readOnly bool) {
	var isolationLevel string

	err := exec.QueryRowContext(context.Background(), "SHOW transaction isolation level").Scan(&isolationLevel)
	require.NoError(t, err)

	require.Equal(t, expectedIsolationLevel, isolationLevel)

	var txReadOnly transactionReadOnly

	err = exec.QueryRowContext(context.Background(), "SHOW transaction_read_only").Scan(&txReadOnly)
	require.NoError(t, err)

	require.Equal(t, readOnly, txReadOnly.readOnly)
}

func Test_SQLProvider_Rollback_Commit(t *testing.T) {
	t.Parallel()

	const createUsersTableQuery = `
		CREATE TABLE users (
			id uuid primary key,
			age smallint not null
		)
	`

	ctx := context.Background()

	db := postgrescontainer.RunForTesting(t, postgrescontainer.EmptyMigrations{}, createUsersTableQuery)

	provider := stdlibtransaction.NewProvider(db)

	tx, err := provider.Begin(ctx)
	require.NoError(t, err)

	assertTxCommit(t, provider.Executor(tx.Context()), tx, db)

	tx, err = provider.Begin(ctx)
	require.NoError(t, err)

	assertTxRollback(t, provider.Executor(tx.Context()), tx, db)

	opts := sql.TxOptions{Isolation: sql.LevelReadCommitted}

	tx, err = provider.BeginTx(ctx, opts)
	require.NoError(t, err)

	assertTxCommit(t, provider.Executor(tx.Context()), tx, db)

	tx, err = provider.BeginTx(ctx, opts)
	require.NoError(t, err)

	assertTxRollback(t, provider.Executor(tx.Context()), tx, db)
}
