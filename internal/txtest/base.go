package txtest_test

import (
	context "context"
	sql "database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"

	"github.com/amidgo/tx"

	sqltx "github.com/amidgo/tx/sql"
	sqlxtx "github.com/amidgo/tx/sqlx"
)

type Executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

var (
	_ Executor = bun.IDB(nil)
	_ Executor = sqltx.Executor(nil)
	_ Executor = sqlxtx.Executor(nil)
)

func AssertTxCommit(
	t *testing.T,
	beginner tx.Beginner,
	exec Executor,
	tx tx.Tx,
	db *sql.DB,
) {
	expectedUserID := uuid.New()
	expectedUserAge := 10

	const insertUserQuery = "INSERT INTO users (id, age) VALUES ($1, $2)"

	_, err := exec.ExecContext(tx.Context(), insertUserQuery, expectedUserID, expectedUserAge)
	require.NoError(t, err)

	AssertUserNotFound(t, db, expectedUserID)

	err = tx.Commit()
	require.NoError(t, err)

	AssertUserExists(t, db, expectedUserID, expectedUserAge)

	enabled := txEnabled(tx.Context(), beginner)

	require.False(t, enabled)
}

func AssertBunTxCommit(
	t *testing.T,
	beginner tx.Beginner,
	exec Executor,
	tx tx.Tx,
	db *sql.DB,
) {
	expectedUserID := uuid.New()
	expectedUserAge := 10

	const insertUserQuery = "INSERT INTO users (id, age) VALUES (?, ?)"

	_, err := exec.ExecContext(tx.Context(), insertUserQuery, expectedUserID, expectedUserAge)
	require.NoError(t, err)

	AssertUserNotFound(t, db, expectedUserID)

	err = tx.Commit()
	require.NoError(t, err)

	AssertUserExists(t, db, expectedUserID, expectedUserAge)

	enabled := txEnabled(tx.Context(), beginner)
	require.False(t, enabled)
}

func AssertTxRollback(
	t *testing.T,
	beginner tx.Beginner,
	exec Executor,
	tx tx.Tx,
	db *sql.DB,
) {
	expectedUserID := uuid.New()
	expectedUserAge := 10

	const insertUserQuery = "INSERT INTO users (id, age) VALUES ($1, $2)"

	_, err := exec.ExecContext(tx.Context(), insertUserQuery, expectedUserID, expectedUserAge)
	require.NoError(t, err)

	AssertUserNotFound(t, db, expectedUserID)

	err = tx.Rollback()
	require.NoError(t, err)

	AssertUserNotFound(t, db, expectedUserID)

	enabled := txEnabled(tx.Context(), beginner)
	require.False(t, enabled)
}

func AssertBunTxRollback(
	t *testing.T,
	beginner tx.Beginner,
	exec Executor,
	tx tx.Tx,
	db *sql.DB,
) {
	expectedUserID := uuid.New()
	expectedUserAge := 10

	const insertUserQuery = "INSERT INTO users (id, age) VALUES (?, ?)"

	_, err := exec.ExecContext(tx.Context(), insertUserQuery, expectedUserID, expectedUserAge)
	require.NoError(t, err)

	AssertUserNotFound(t, db, expectedUserID)

	err = tx.Rollback()
	require.NoError(t, err)

	AssertUserNotFound(t, db, expectedUserID)

	enabled := txEnabled(tx.Context(), beginner)
	require.False(t, enabled)
}

func AssertUserNotFound(t *testing.T, db *sql.DB, userID uuid.UUID) {
	id := uuid.UUID{}

	err := db.QueryRowContext(context.Background(), "SELECT id FROM users WHERE id = $1", userID).Scan(&id)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

func AssertUserExists(t *testing.T, db *sql.DB, userID uuid.UUID, userAge int) {
	id := uuid.UUID{}
	age := 0

	err := db.QueryRow("SELECT id,age FROM users WHERE id = $1", userID).Scan(&id, &age)
	require.NoError(t, err)

	require.Equal(t, userID, id)
	require.Equal(t, userAge, age)
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

func txEnabled(ctx context.Context, beginner tx.Beginner) bool {
	enabled := beginner.(interface {
		TxEnabled(ctx context.Context) bool
	})

	return enabled.TxEnabled(ctx)
}

func AssertSQLTransactionLevel(t *testing.T, exec Executor, expectedIsolationLevel string, readOnly bool) {
	var isolationLevel string

	err := exec.QueryRowContext(context.Background(), "SHOW transaction isolation level").Scan(&isolationLevel)
	require.NoError(t, err)

	require.Equal(t, expectedIsolationLevel, isolationLevel)

	var txReadOnly transactionReadOnly

	err = exec.QueryRowContext(context.Background(), "SHOW transaction_read_only").Scan(&txReadOnly)
	require.NoError(t, err)

	require.Equal(t, readOnly, txReadOnly.readOnly)
}
