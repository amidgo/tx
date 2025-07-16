package txtest

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

type placeholder string

const (
	quesionMarkPlaceholder placeholder = "?"
)

type txTestOptions struct {
	placeholder placeholder
}

func WithQuestionMarkPlaceholder(opts *txTestOptions) {
	opts.placeholder = quesionMarkPlaceholder
}

type Option func(*txTestOptions)

func makeTxTestOpts(opts ...Option) *txTestOptions {
	txTestOpts := new(txTestOptions)

	for _, op := range opts {
		op(txTestOpts)
	}

	return txTestOpts
}

func AssertTxCommit(
	t *testing.T,
	beginner tx.Beginner,
	exec Executor,
	tx tx.Tx,
	nonTxExec Executor,
	opts ...Option,
) {
	expectedUserID := uuid.New()
	expectedUserAge := 10

	insertUserQuery := "INSERT INTO users (id, age) VALUES ($1, $2)"

	if makeTxTestOpts(opts...).placeholder == quesionMarkPlaceholder {
		insertUserQuery = "INSERT INTO users (id, age) VALUES (?, ?)"
	}

	_, err := exec.ExecContext(tx.Context(), insertUserQuery, expectedUserID, expectedUserAge)
	require.NoError(t, err)

	AssertUserNotFound(t, nonTxExec, expectedUserID, opts...)

	err = tx.Commit()
	require.NoError(t, err)

	AssertUserExists(t, nonTxExec, expectedUserID, expectedUserAge, opts...)

	enabled := txEnabled(tx.Context(), beginner)

	require.False(t, enabled)
}

func AssertTxRollback(
	t *testing.T,
	beginner tx.Beginner,
	exec Executor,
	tx tx.Tx,
	nonTxExec Executor,
	opts ...Option,
) {
	expectedUserID := uuid.New()
	expectedUserAge := 10

	insertUserQuery := "INSERT INTO users (id, age) VALUES ($1, $2)"

	if makeTxTestOpts(opts...).placeholder == quesionMarkPlaceholder {
		insertUserQuery = "INSERT INTO users (id, age) VALUES (?, ?)"
	}

	_, err := exec.ExecContext(tx.Context(), insertUserQuery, expectedUserID, expectedUserAge)
	require.NoError(t, err)

	AssertUserNotFound(t, nonTxExec, expectedUserID, opts...)

	err = tx.Rollback()
	require.NoError(t, err)

	AssertUserNotFound(t, nonTxExec, expectedUserID, opts...)

	enabled := txEnabled(tx.Context(), beginner)
	require.False(t, enabled)
}

func AssertUserNotFound(
	t *testing.T,
	exec Executor,
	userID uuid.UUID,
	opts ...Option,
) {
	id := uuid.UUID{}

	query := "SELECT id FROM users WHERE id = $1"

	if makeTxTestOpts(opts...).placeholder == quesionMarkPlaceholder {
		query = "SELECT id FROM users WHERE id = ?"
	}

	err := exec.QueryRowContext(context.Background(), query, userID).Scan(&id)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

func AssertUserExists(
	t *testing.T,
	exec Executor,
	userID uuid.UUID,
	userAge int,
	opts ...Option,
) {
	id := uuid.UUID{}
	age := 0

	query := "SELECT id, age FROM users WHERE id = $1"

	if makeTxTestOpts(opts...).placeholder == quesionMarkPlaceholder {
		query = "SELECT id, age FROM users WHERE id = ?"
	}

	err := exec.QueryRowContext(context.Background(), query, userID).Scan(&id, &age)
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

func AssertBeginError(
	t *testing.T,
	ctx context.Context,
	beginner tx.Beginner,
	txOpts *sql.TxOptions,
	expectedBeginError error,
) {
	tx, err := beginner.Begin(ctx)
	if !errors.Is(err, expectedBeginError) {
		t.Fatalf("assert beginner.Begin error, not match\n\nexpected:\n%s\n\nactual:\n%s", expectedBeginError, err)
	}

	if tx != nil {
		t.Fatal("assert beginner.Begin tx is nil on error, unexpected non nil tx")
	}

	tx, err = beginner.BeginTx(ctx, txOpts)
	if !errors.Is(err, expectedBeginError) {
		t.Fatalf("assert beginner.BeginTx error, not match\n\nexpected:\n%s\n\nactual:\n%s", expectedBeginError, err)
	}

	if tx != nil {
		t.Fatal("assert beginner.BeginTx tx is nil on error, unexpected non nil tx")
	}
}
