package transaction_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	postgrescontainer "github.com/amidgo/containers/postgres"
	"github.com/amidgo/transaction"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test_SQLProvider_Begin_BeginTx(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	db := postgrescontainer.RunForTesting(t, postgrescontainer.EmptyMigrations{})

	provider := transaction.NewSQLProvider(db)

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

func assertSQLTransactionEnabled(t *testing.T, provider *transaction.SQLProvider, tx transaction.Transaction, expectedIsolationLevel string, readOnly bool) {
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

type transactionReadOnly struct {
	readOnly bool
}

var errInvalidTransactionReadOnlyValue = errors.New("invalid transaction read only value")

func (t *transactionReadOnly) Scan(src any) error {
	s := sql.NullString{}

	err := s.Scan(src)
	if err != nil {
		return err
	}

	switch s.String {
	case "on":
		t.readOnly = true

		return nil
	case "off":
		t.readOnly = false

		return nil
	default:
		return errInvalidTransactionReadOnlyValue
	}
}

func assertSQLTransactionLevel(t *testing.T, exec transaction.SQLExecutor, expectedIsolationLevel string, readOnly bool) {
	var isolationLevel string

	err := exec.QueryRow("SHOW transaction isolation level").Scan(&isolationLevel)
	require.NoError(t, err)

	require.Equal(t, expectedIsolationLevel, isolationLevel)

	var txReadOnly transactionReadOnly

	err = exec.QueryRow("SHOW transaction_read_only").Scan(&txReadOnly)
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

	provider := transaction.NewSQLProvider(db)

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

func assertTxCommit(t *testing.T, exec transaction.SQLExecutor, tx transaction.Transaction, db *sql.DB) {
	expectedUserID := uuid.New()
	expectedUserAge := 10

	const insertUserQuery = "INSERT INTO users (id, age) VALUES ($1, $2)"

	_, err := exec.ExecContext(tx.Context(), insertUserQuery, expectedUserID, expectedUserAge)
	require.NoError(t, err)

	assertUserNotFound(t, db, expectedUserID)

	err = tx.Commit(tx.Context())
	require.NoError(t, err)

	assertUserExists(t, db, expectedUserID, expectedUserAge)

	enabled := transaction.TxEnabled(tx.Context())
	require.False(t, enabled)
}

func assertTxRollback(t *testing.T, exec transaction.SQLExecutor, tx transaction.Transaction, db *sql.DB) {
	expectedUserID := uuid.New()
	expectedUserAge := 10

	const insertUserQuery = "INSERT INTO users (id, age) VALUES ($1, $2)"

	_, err := exec.ExecContext(tx.Context(), insertUserQuery, expectedUserID, expectedUserAge)
	require.NoError(t, err)

	assertUserNotFound(t, db, expectedUserID)

	err = tx.Rollback(tx.Context())
	require.NoError(t, err)

	assertUserNotFound(t, db, expectedUserID)

	enabled := transaction.TxEnabled(tx.Context())
	require.False(t, enabled)
}

func assertUserNotFound(t *testing.T, db *sql.DB, userID uuid.UUID) {
	id := uuid.UUID{}

	err := db.QueryRow("SELECT id FROM users WHERE id = $1", userID).Scan(&id)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

func assertUserExists(t *testing.T, db *sql.DB, userID uuid.UUID, userAge int) {
	id := uuid.UUID{}
	age := 0

	err := db.QueryRow("SELECT id,age FROM users WHERE id = $1", userID).Scan(&id, &age)
	require.NoError(t, err)

	require.Equal(t, userID, id)
	require.Equal(t, userAge, age)
}
