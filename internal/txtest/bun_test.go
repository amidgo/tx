package txtest_test

import (
	context "context"
	sql "database/sql"
	"errors"
	"testing"

	postgrescontainer "github.com/amidgo/containers/postgres"
	"github.com/amidgo/tx"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	buntx "github.com/amidgo/tx/bun"
)

func Test_BunBeginner_Begin_BeginTx(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db := postgrescontainer.RunForTesting(t, postgrescontainer.EmptyMigrations{})

	bunDB := bun.NewDB(db, pgdialect.New())

	beginner := buntx.NewBeginner(bunDB)

	exec := beginner.Executor(ctx)
	_, ok := exec.(*bun.DB)
	require.True(t, ok)

	tx, err := beginner.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable, ReadOnly: false})
	require.NoError(t, err)

	assertBunTransactionEnabled(t, beginner, tx, "serializable", false)

	tx, err = beginner.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	require.NoError(t, err)

	assertBunTransactionEnabled(t, beginner, tx, "repeatable read", true)

	tx, err = beginner.Begin(ctx)
	require.NoError(t, err)

	assertBunTransactionEnabled(t, beginner, tx, "read committed", false)
}

func Test_BunBeginner_Rollback_Commit(t *testing.T) {
	t.Parallel()

	const createUsersTableQuery = `
		CREATE TABLE users (
			id uuid primary key,
			age smallint not null
		)
	`

	ctx := context.Background()

	db := postgrescontainer.RunForTesting(t, postgrescontainer.EmptyMigrations{}, createUsersTableQuery)

	bunDB := bun.NewDB(db, pgdialect.New())

	beginner := buntx.NewBeginner(bunDB)

	tx, err := beginner.Begin(ctx)
	require.NoError(t, err)

	AssertBunTxCommit(t, beginner, beginner.Executor(tx.Context()), tx, db)

	tx, err = beginner.Begin(ctx)
	require.NoError(t, err)

	AssertBunTxRollback(t, beginner, beginner.Executor(tx.Context()), tx, db)

	opts := &sql.TxOptions{Isolation: sql.LevelReadCommitted}

	tx, err = beginner.BeginTx(ctx, opts)
	require.NoError(t, err)

	AssertBunTxCommit(t, beginner, beginner.Executor(tx.Context()), tx, db)

	tx, err = beginner.BeginTx(ctx, opts)
	require.NoError(t, err)

	AssertBunTxRollback(t, beginner, beginner.Executor(tx.Context()), tx, db)
}

func Test_BunBeginner(t *testing.T) {
	t.Parallel()

	const createUsersTableQuery = `
		CREATE TABLE users (
			id uuid primary key,
			age smallint not null
		)
	`

	errStub := errors.New("stub err")

	db := postgrescontainer.RunForTesting(t, postgrescontainer.EmptyMigrations{}, createUsersTableQuery)

	bunDB := bun.NewDB(db, pgdialect.New())

	beginner := buntx.NewBeginner(bunDB)

	t.Run("no external tx, execution failed, rollback expected", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		userID := uuid.New()

		err := beginner.WithTx(ctx,
			func(ctx context.Context, exec buntx.Executor) error {
				_, err := exec.ExecContext(ctx, "INSERT INTO users (id, age) VALUES (?, ?)", userID, 100)
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

		AssertUserNotFound(t, db, userID)
	})

	t.Run("no external tx, execution success, commit expected", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		userID := uuid.New()
		userAge := 100

		err := beginner.WithTx(ctx,
			func(ctx context.Context, exec buntx.Executor) error {
				_, err := exec.ExecContext(ctx, "INSERT INTO users (id, age) VALUES (?, ?)", userID, userAge)
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

		AssertUserExists(t, db, userID, userAge)
	})
}

func assertBunTransactionEnabled(t *testing.T, beginner *buntx.Beginner, tx tx.Tx, expectedIsolationLevel string, readOnly bool) {
	enabled := beginner.TxEnabled(tx.Context())
	require.True(t, enabled)

	exec := beginner.Executor(tx.Context())
	_, ok := exec.(bun.Tx)
	require.True(t, ok)

	AssertSQLTransactionLevel(t, exec, expectedIsolationLevel, readOnly)

	err := tx.Rollback()
	require.NoError(t, err)

	enabled = beginner.TxEnabled(tx.Context())
	require.False(t, enabled)
}
