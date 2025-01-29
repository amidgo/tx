package transaction_test

import (
	context "context"
	sql "database/sql"
	"errors"
	"testing"

	postgrescontainer "github.com/amidgo/containers/postgres"
	"github.com/amidgo/transaction"
	buntransaction "github.com/amidgo/transaction/bun"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func Test_BunProvider_Begin_BeginTx(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db := postgrescontainer.RunForTesting(t, postgrescontainer.EmptyMigrations{})

	bunDB := bun.NewDB(db, pgdialect.New())

	provider := buntransaction.NewProvider(bunDB)

	exec := provider.Executor(ctx)
	_, ok := exec.(*bun.DB)
	require.True(t, ok)

	tx, err := provider.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable, ReadOnly: false})
	require.NoError(t, err)

	assertBunTransactionEnabled(t, provider, tx, "serializable", false)

	tx, err = provider.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	require.NoError(t, err)

	assertBunTransactionEnabled(t, provider, tx, "repeatable read", true)

	tx, err = provider.Begin(ctx)
	require.NoError(t, err)

	assertBunTransactionEnabled(t, provider, tx, "read committed", false)
}

func Test_BunProvider_Rollback_Commit(t *testing.T) {
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

	provider := buntransaction.NewProvider(bunDB)

	tx, err := provider.Begin(ctx)
	require.NoError(t, err)

	assertBunTxCommit(t, provider, provider.Executor(tx.Context()), tx, db)

	tx, err = provider.Begin(ctx)
	require.NoError(t, err)

	assertBunTxRollback(t, provider, provider.Executor(tx.Context()), tx, db)

	opts := &sql.TxOptions{Isolation: sql.LevelReadCommitted}

	tx, err = provider.BeginTx(ctx, opts)
	require.NoError(t, err)

	assertBunTxCommit(t, provider, provider.Executor(tx.Context()), tx, db)

	tx, err = provider.BeginTx(ctx, opts)
	require.NoError(t, err)

	assertBunTxRollback(t, provider, provider.Executor(tx.Context()), tx, db)
}

func Test_BunProvider_WithTx(t *testing.T) {
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

	provider := buntransaction.NewProvider(bunDB)

	t.Run("no external tx, execution failed, rollback expected", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		userID := uuid.New()

		err := provider.WithTx(ctx,
			func(ctx context.Context, exec buntransaction.Executor) error {
				_, err := exec.ExecContext(ctx, "INSERT INTO users (id, age) VALUES (?, ?)", userID, 100)
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
			func(ctx context.Context, exec buntransaction.Executor) error {
				_, err := exec.ExecContext(ctx, "INSERT INTO users (id, age) VALUES (?, ?)", userID, userAge)
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

	t.Run("external tx, success execution", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		userID := uuid.New()
		userAge := 100

		tx, err := provider.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelDefault})
		require.NoError(t, err)

		err = provider.WithTx(tx.Context(),
			func(ctx context.Context, exec buntransaction.Executor) error {
				_, err := exec.ExecContext(ctx, "INSERT INTO users (id, age) VALUES (?, ?)", userID, userAge)
				require.NoError(t, err)

				return nil
			},
			&sql.TxOptions{
				Isolation: sql.LevelReadCommitted,
			},
		)
		require.NoError(t, err)

		assertUserNotFound(t, db, userID)

		err = tx.Commit()
		require.NoError(t, err)

		assertUserExists(t, db, userID, userAge)
	})

	t.Run("external tx, execution failed", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)

		userID := uuid.New()
		userAge := 100

		tx, err := provider.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelDefault})
		require.NoError(t, err)

		err = provider.WithTx(tx.Context(),
			func(ctx context.Context, exec buntransaction.Executor) error {
				_, err := exec.ExecContext(ctx, "INSERT INTO users (id, age) VALUES (?, ?)", userID, userAge)
				require.NoError(t, err)

				return errStub
			},
			&sql.TxOptions{
				Isolation: sql.LevelReadCommitted,
			},
		)
		require.ErrorIs(t, err, errStub)

		assertUserNotFound(t, db, userID)

		err = tx.Commit()
		require.NoError(t, err)

		assertUserExists(t, db, userID, userAge)
	})
}

func assertBunTransactionEnabled(t *testing.T, provider *buntransaction.Provider, tx transaction.Transaction, expectedIsolationLevel string, readOnly bool) {
	enabled := provider.TxEnabled(tx.Context())
	require.True(t, enabled)

	exec := provider.Executor(tx.Context())
	_, ok := exec.(*bun.Tx)
	require.True(t, ok)

	assertSQLTransactionLevel(t, exec, expectedIsolationLevel, readOnly)

	err := tx.Rollback()
	require.NoError(t, err)

	enabled = provider.TxEnabled(tx.Context())
	require.False(t, enabled)
}
