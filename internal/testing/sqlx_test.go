package tx_test

import (
	context "context"
	sql "database/sql"
	"errors"
	"testing"

	postgrescontainer "github.com/amidgo/containers/postgres"
	"github.com/amidgo/tx"
	sqlxtx "github.com/amidgo/tx/sqlx"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func Test_SqlxProvider_Begin_BeginTx(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db := postgrescontainer.RunForTesting(t, postgrescontainer.EmptyMigrations{})
	sqlxDB := sqlx.NewDb(db, "pgx")

	provider := sqlxtx.NewProvider(sqlxDB)

	tx, err := provider.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable, ReadOnly: false})
	require.NoError(t, err)

	assertSqlxTransactionEnabled(t, provider, tx, "serializable", false)

	tx, err = provider.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	require.NoError(t, err)

	assertSqlxTransactionEnabled(t, provider, tx, "repeatable read", true)

	tx, err = provider.Begin(ctx)
	require.NoError(t, err)

	assertSqlxTransactionEnabled(t, provider, tx, "read committed", false)
}

func assertSqlxTransactionEnabled(t *testing.T, provider *sqlxtx.Provider, tx tx.Tx, expectedIsolationLevel string, readOnly bool) {
	enabled := provider.TxEnabled(tx.Context())
	require.True(t, enabled)

	exec := provider.Executor(tx.Context())
	_, ok := exec.(*sqlx.Tx)
	require.True(t, ok)

	assertSQLTransactionLevel(t, exec, expectedIsolationLevel, readOnly)

	err := tx.Rollback()
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

	provider := sqlxtx.NewProvider(sqlxDB)

	tx, err := provider.Begin(ctx)
	require.NoError(t, err)

	assertTxCommit(t, provider, provider.Executor(tx.Context()), tx, db)

	tx, err = provider.Begin(ctx)
	require.NoError(t, err)

	assertTxRollback(t, provider, provider.Executor(tx.Context()), tx, db)

	opts := &sql.TxOptions{Isolation: sql.LevelReadCommitted}

	tx, err = provider.BeginTx(ctx, opts)
	require.NoError(t, err)

	assertTxCommit(t, provider, provider.Executor(tx.Context()), tx, db)

	tx, err = provider.BeginTx(ctx, opts)
	require.NoError(t, err)

	assertTxRollback(t, provider, provider.Executor(tx.Context()), tx, db)
}

func Test_SqlxProvider_WithTx(t *testing.T) {
	t.Parallel()

	const createUsersTableQuery = `
		CREATE TABLE users (
			id uuid primary key,
			age smallint not null
		)
	`

	errStub := errors.New("stub err")

	db := postgrescontainer.RunForTesting(t, postgrescontainer.EmptyMigrations{}, createUsersTableQuery)

	sqlxDB := sqlx.NewDb(db, "pgx")

	provider := sqlxtx.NewProvider(sqlxDB)

	t.Run("no external tx, execution failed, rollback expected", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		userID := uuid.New()

		err := provider.WithTx(ctx,
			func(ctx context.Context, exec sqlxtx.Executor) error {
				_, err := exec.ExecContext(ctx, "INSERT INTO users (id, age) VALUES ($1, $2)", userID, 100)
				require.NoError(t, err)

				return errStub
			},
			&sql.TxOptions{
				Isolation: sql.LevelReadCommitted,
			},
		)
		require.ErrorIs(t, err, errStub)

		assertUserNotFound(t, db, userID)
	})

	t.Run("no external tx, execution success, commit expected", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		userID := uuid.New()
		userAge := 100

		err := provider.WithTx(ctx,
			func(ctx context.Context, exec sqlxtx.Executor) error {
				_, err := exec.ExecContext(ctx, "INSERT INTO users (id, age) VALUES ($1, $2)", userID, userAge)
				require.NoError(t, err)

				return nil
			},
			&sql.TxOptions{
				Isolation: sql.LevelReadCommitted,
			},
		)
		require.NoError(t, err)

		assertUserExists(t, db, userID, userAge)
	})
}
