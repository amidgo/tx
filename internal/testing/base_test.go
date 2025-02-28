package tx_test

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

type executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

var (
	_ executor = bun.IDB(nil)
	_ executor = sqltx.Executor(nil)
	_ executor = sqlxtx.Executor(nil)
)

func assertTxCommit(
	t *testing.T,
	provider tx.Beginner,
	exec executor,
	tx tx.Tx,
	db *sql.DB,
) {
	expectedUserID := uuid.New()
	expectedUserAge := 10

	const insertUserQuery = "INSERT INTO users (id, age) VALUES ($1, $2)"

	_, err := exec.ExecContext(tx.Context(), insertUserQuery, expectedUserID, expectedUserAge)
	require.NoError(t, err)

	assertUserNotFound(t, db, expectedUserID)

	err = tx.Commit()
	require.NoError(t, err)

	assertUserExists(t, db, expectedUserID, expectedUserAge)

	enabled := provider.TxEnabled(tx.Context())
	require.False(t, enabled)
}

func assertBunTxCommit(
	t *testing.T,
	provider tx.Beginner,
	exec executor,
	tx tx.Tx,
	db *sql.DB,
) {
	expectedUserID := uuid.New()
	expectedUserAge := 10

	const insertUserQuery = "INSERT INTO users (id, age) VALUES (?, ?)"

	_, err := exec.ExecContext(tx.Context(), insertUserQuery, expectedUserID, expectedUserAge)
	require.NoError(t, err)

	assertUserNotFound(t, db, expectedUserID)

	err = tx.Commit()
	require.NoError(t, err)

	assertUserExists(t, db, expectedUserID, expectedUserAge)

	enabled := provider.TxEnabled(tx.Context())
	require.False(t, enabled)
}

func assertTxRollback(
	t *testing.T,
	provider tx.Beginner,
	exec executor,
	tx tx.Tx,
	db *sql.DB,
) {
	expectedUserID := uuid.New()
	expectedUserAge := 10

	const insertUserQuery = "INSERT INTO users (id, age) VALUES ($1, $2)"

	_, err := exec.ExecContext(tx.Context(), insertUserQuery, expectedUserID, expectedUserAge)
	require.NoError(t, err)

	assertUserNotFound(t, db, expectedUserID)

	err = tx.Rollback()
	require.NoError(t, err)

	assertUserNotFound(t, db, expectedUserID)

	enabled := provider.TxEnabled(tx.Context())
	require.False(t, enabled)
}

func assertBunTxRollback(
	t *testing.T,
	provider tx.Beginner,
	exec executor,
	tx tx.Tx,
	db *sql.DB,
) {
	expectedUserID := uuid.New()
	expectedUserAge := 10

	const insertUserQuery = "INSERT INTO users (id, age) VALUES (?, ?)"

	_, err := exec.ExecContext(tx.Context(), insertUserQuery, expectedUserID, expectedUserAge)
	require.NoError(t, err)

	assertUserNotFound(t, db, expectedUserID)

	err = tx.Rollback()
	require.NoError(t, err)

	assertUserNotFound(t, db, expectedUserID)

	enabled := provider.TxEnabled(tx.Context())
	require.False(t, enabled)
}

func assertUserNotFound(t *testing.T, db *sql.DB, userID uuid.UUID) {
	id := uuid.UUID{}

	err := db.QueryRowContext(context.Background(), "SELECT id FROM users WHERE id = $1", userID).Scan(&id)
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
