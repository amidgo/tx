package transaction_test

import (
	context "context"
	sql "database/sql"
	"testing"

	postgrescontainer "github.com/amidgo/containers/postgres"
	"github.com/amidgo/transaction"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func Test_SqlxProvider_Begin_BeginTx(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db := postgrescontainer.RunForTesting(t, postgrescontainer.EmptyMigrations{})
	sqlxDB := sqlx.NewDb(db, "pgx")

	provider := transaction.NewSqlxProvider(sqlxDB)

	tx, err := provider.BeginTx(ctx, sql.TxOptions{Isolation: sql.LevelSerializable, ReadOnly: false})
	require.NoError(t, err)

	assertSqlxTransactionEnabled(t, provider, tx, "serializable", false)

	tx, err = provider.BeginTx(ctx, sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	require.NoError(t, err)

	assertSqlxTransactionEnabled(t, provider, tx, "repeatable read", true)

	tx, err = provider.Begin(ctx)
	require.NoError(t, err)

	assertSqlxTransactionEnabled(t, provider, tx, "read committed", false)
}

func assertSqlxTransactionEnabled(t *testing.T, provider *transaction.SqlxProvider, tx transaction.Transaction, expectedIsolationLevel string, readOnly bool) {
	enabled := provider.TxEnabled(tx.Context())
	require.True(t, enabled)

	exec := provider.Executor(tx.Context())
	_, ok := exec.(*sqlx.Tx)
	require.True(t, ok)

	assertSQLTransactionLevel(t, exec, expectedIsolationLevel, readOnly)

	err := tx.Rollback(tx.Context())
	require.NoError(t, err)

	enabled = provider.TxEnabled(tx.Context())
	require.False(t, enabled)
}

func Test_SqlxProvider_Rollback_Commit(t *testing.T) {
	t.Parallel()

	const createUsersTableQuery = `
		CREATE TABLE users (
			id uuid primary key,
			age smallint not null
		)
	`

	ctx := context.Background()

	db := postgrescontainer.RunForTesting(t, postgrescontainer.EmptyMigrations{}, createUsersTableQuery)

	sqlxDB := sqlx.NewDb(db, "pgx")

	provider := transaction.NewSqlxProvider(sqlxDB)

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
