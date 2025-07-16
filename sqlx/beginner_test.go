package sqlxtx_test

import (
	context "context"
	sql "database/sql"
	"errors"
	"testing"

	postgrescontainer "github.com/amidgo/containers/postgres"
	"github.com/amidgo/containers/postgres/migrations"
	pgrunner "github.com/amidgo/containers/postgres/runner"
	"github.com/amidgo/tx"
	"github.com/amidgo/tx/internal/reusable"
	txtest "github.com/amidgo/tx/internal/testing"
	sqlxtx "github.com/amidgo/tx/sqlx"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func Test_SqlxBeginner_Begin_BeginTx(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db := postgrescontainer.ReuseForTesting(t,
		reusable.Postgres(),
		migrations.Nil,
	)

	sqlxDB := sqlx.NewDb(db, "pgx")

	beginner := sqlxtx.NewBeginner(sqlxDB)

	tx, err := beginner.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable, ReadOnly: false})
	require.NoError(t, err)

	assertSqlxTransactionEnabled(t, beginner, tx, "serializable", false)

	tx, err = beginner.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	require.NoError(t, err)

	assertSqlxTransactionEnabled(t, beginner, tx, "repeatable read", true)

	tx, err = beginner.Begin(ctx)
	require.NoError(t, err)

	assertSqlxTransactionEnabled(t, beginner, tx, "read committed", false)
}

func assertSqlxTransactionEnabled(t *testing.T, beginner *sqlxtx.Beginner, tx tx.Tx, expectedIsolationLevel string, readOnly bool) {
	enabled := beginner.TxEnabled(tx.Context())
	require.True(t, enabled)

	exec := beginner.Executor(tx.Context())
	_, ok := exec.(*sqlx.Tx)
	require.True(t, ok)

	txtest.AssertSQLTransactionLevel(t, exec, expectedIsolationLevel, readOnly)

	err := tx.Rollback()
	require.NoError(t, err)

	enabled = beginner.TxEnabled(tx.Context())
	require.False(t, enabled)
}

func Test_SqlxBeginner_Rollback_Commit(t *testing.T) {
	t.Parallel()

	const createUsersTableQuery = `
		CREATE TABLE users (
			id uuid primary key,
			age smallint not null
		)
	`

	ctx := context.Background()

	db := postgrescontainer.ReuseForTesting(t,
		reusable.Postgres(),
		migrations.Nil,
		createUsersTableQuery,
	)

	sqlxDB := sqlx.NewDb(db, "pgx")

	beginner := sqlxtx.NewBeginner(sqlxDB)

	tx, err := beginner.Begin(ctx)
	require.NoError(t, err)

	txtest.AssertTxCommit(t, beginner, beginner.Executor(tx.Context()), tx, db)

	tx, err = beginner.Begin(ctx)
	require.NoError(t, err)

	txtest.AssertTxRollback(t, beginner, beginner.Executor(tx.Context()), tx, db)

	opts := &sql.TxOptions{Isolation: sql.LevelReadCommitted}

	tx, err = beginner.BeginTx(ctx, opts)
	require.NoError(t, err)

	txtest.AssertTxCommit(t, beginner, beginner.Executor(tx.Context()), tx, db)

	tx, err = beginner.BeginTx(ctx, opts)
	require.NoError(t, err)

	txtest.AssertTxRollback(t, beginner, beginner.Executor(tx.Context()), tx, db)
}

func Test_SqlxBeginner_WithTx(t *testing.T) {
	t.Parallel()

	const createUsersTableQuery = `
		CREATE TABLE users (
			id uuid primary key,
			age smallint not null
		)
	`

	errStub := errors.New("stub err")

	db := postgrescontainer.ReuseForTesting(t,
		reusable.Postgres(),
		migrations.Nil,
		createUsersTableQuery,
	)

	sqlxDB := sqlx.NewDb(db, "pgx")

	beginner := sqlxtx.NewBeginner(sqlxDB)

	t.Run("no external tx, execution failed, rollback expected", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		userID := uuid.New()

		err := beginner.WithTx(ctx,
			func(ctx context.Context, exec sqlxtx.Executor) error {
				_, err := exec.ExecContext(ctx, "INSERT INTO users (id, age) VALUES ($1, $2)", userID, 100)
				require.NoError(t, err)

				enabled := beginner.TxEnabled(ctx)
				if enabled {
					t.Fatalf("in WithTx ctx should be without tx, because exec is provided")
				}

				return errStub
			},
			&sql.TxOptions{
				Isolation: sql.LevelReadCommitted,
			},
		)
		require.ErrorIs(t, err, errStub)

		txtest.AssertUserNotFound(t, db, userID)
	})

	t.Run("no external tx, execution success, commit expected", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		userID := uuid.New()
		userAge := 100

		err := beginner.WithTx(ctx,
			func(ctx context.Context, exec sqlxtx.Executor) error {
				_, err := exec.ExecContext(ctx, "INSERT INTO users (id, age) VALUES ($1, $2)", userID, userAge)
				require.NoError(t, err)

				enabled := beginner.TxEnabled(ctx)
				if enabled {
					t.Fatalf("in WithTx ctx should be without tx, because exec is provided")
				}

				return nil
			},
			&sql.TxOptions{
				Isolation: sql.LevelReadCommitted,
			},
		)
		require.NoError(t, err)

		txtest.AssertUserExists(t, db, userID, userAge)
	})
}

func Test_SqlxBeginner_Error(t *testing.T) {
	t.Parallel()

	reusable := postgrescontainer.NewReusable(pgrunner.RunContainer(nil))

	db := postgrescontainer.ReuseForTesting(t,
		reusable,
		migrations.Nil,
	)

	sqlxDB := sqlx.NewDb(db, "pgx")

	beginner := sqlxtx.NewBeginner(sqlxDB)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	txtest.AssertBeginError(t, ctx, beginner, nil, context.Canceled)
}
