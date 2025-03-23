package tx_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	postgrescontainer "github.com/amidgo/containers/postgres"
	"github.com/amidgo/tx"
	sqltx "github.com/amidgo/tx/sql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test_SQLBeginner_Begin_BeginTx(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db := postgrescontainer.RunForTesting(t, postgrescontainer.EmptyMigrations{})

	beginner := sqltx.NewBeginner(db)

	exec := beginner.Executor(ctx)
	_, ok := exec.(*sql.DB)
	require.True(t, ok)

	tx, err := beginner.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable, ReadOnly: false})
	require.NoError(t, err)

	assertSQLTransactionEnabled(t, beginner, tx, "serializable", false)

	tx, err = beginner.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	require.NoError(t, err)

	assertSQLTransactionEnabled(t, beginner, tx, "repeatable read", true)

	tx, err = beginner.Begin(ctx)
	require.NoError(t, err)

	assertSQLTransactionEnabled(t, beginner, tx, "read committed", false)
}

func assertSQLTransactionEnabled(
	t *testing.T,
	beginner *sqltx.Beginner,
	tx tx.Tx,
	expectedIsolationLevel string,
	readOnly bool,
) {
	enabled := beginner.TxEnabled(tx.Context())
	require.True(t, enabled)

	exec := beginner.Executor(tx.Context())
	_, ok := exec.(*sql.Tx)
	require.True(t, ok)

	assertSQLTransactionLevel(t, exec, expectedIsolationLevel, readOnly)

	err := tx.Rollback()
	require.NoError(t, err)

	enabled = beginner.TxEnabled(tx.Context())
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

func Test_SQLBeginner_Rollback_Commit(t *testing.T) {
	t.Parallel()

	const createUsersTableQuery = `
		CREATE TABLE users (
			id uuid primary key,
			age smallint not null
		)
	`

	ctx := context.Background()

	db := postgrescontainer.RunForTesting(t, postgrescontainer.EmptyMigrations{}, createUsersTableQuery)

	beginner := sqltx.NewBeginner(db)

	tx, err := beginner.Begin(ctx)
	require.NoError(t, err)

	assertTxCommit(t, beginner, beginner.Executor(tx.Context()), tx, db)

	tx, err = beginner.Begin(ctx)
	require.NoError(t, err)

	assertTxRollback(t, beginner, beginner.Executor(tx.Context()), tx, db)

	opts := &sql.TxOptions{Isolation: sql.LevelReadCommitted}

	tx, err = beginner.BeginTx(ctx, opts)
	require.NoError(t, err)

	assertTxCommit(t, beginner, beginner.Executor(tx.Context()), tx, db)

	tx, err = beginner.BeginTx(ctx, opts)
	require.NoError(t, err)

	assertTxRollback(t, beginner, beginner.Executor(tx.Context()), tx, db)
}

func Test_SQLBeginner_WithTx(t *testing.T) {
	t.Parallel()

	const createUsersTableQuery = `
		CREATE TABLE users (
			id uuid primary key,
			age smallint not null
		)
	`

	errStub := errors.New("stub err")

	db := postgrescontainer.RunForTesting(t, postgrescontainer.EmptyMigrations{}, createUsersTableQuery)

	beginner := sqltx.NewBeginner(db)

	t.Run("no external tx, execution failed, rollback expected", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		userID := uuid.New()

		err := beginner.WithTx(ctx,
			func(ctx context.Context, exec sqltx.Executor) error {
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

		err := beginner.WithTx(ctx,
			func(ctx context.Context, exec sqltx.Executor) error {
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
