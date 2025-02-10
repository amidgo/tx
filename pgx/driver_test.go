package pgxtx_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	postgrescontainer "github.com/amidgo/containers/postgres"
	"github.com/amidgo/tx"
	pgxtx "github.com/amidgo/tx/pgx"
)

func Test_Driver(t *testing.T) {
	t.Parallel()

	t.Run("serializable level", driverSerializationTest)
	t.Run("repeatable read", driverRepeatableReadTest)
}

func driverSerializationTest(t *testing.T) {
	t.Parallel()
	db := postgrescontainer.RunForTesting(t, postgrescontainer.EmptyMigrations{},
		"create table tx_serializable_sums(num integer)",
	)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tx1, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		panic("func, failed begin tx, " + err.Error())
	}

	defer tx1.Rollback()

	tx2, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		panic("func, failed begin tx, " + err.Error())
	}

	defer tx2.Rollback()

	_, err = tx1.ExecContext(ctx, `insert into tx_serializable_sums(num) select sum(num)::int from tx_serializable_sums`)
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}

	_, err = tx2.ExecContext(ctx, `insert into tx_serializable_sums(num) select sum(num)::int from tx_serializable_sums`)
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}

	err = tx1.Commit()
	if err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	err = tx2.Commit()

	driverErr := pgxtx.Driver().Error(err)

	if !errors.Is(driverErr, err) {
		t.Fatalf("invalid error wrapping, original error was erased, original: %+v, driverErr: %+v", err, driverErr)
	}

	if !errors.Is(driverErr, tx.ErrSerialization) {
		t.Fatalf("expected serialization error, actual %+v", err)
	}
}

func driverRepeatableReadTest(t *testing.T) {
	t.Parallel()

	db := postgrescontainer.RunForTesting(t, postgrescontainer.EmptyMigrations{},
		`
CREATE TABLE accounts (
    id INT PRIMARY KEY,
    balance DECIMAL
)
		`,
		"INSERT INTO accounts (id, balance) VALUES (1, 1000);",
	)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	tx1, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		t.Fatalf("failed begin tx, %+v", err)
	}

	defer tx1.Rollback()

	tx2, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		t.Fatalf("failed begin tx, %+v", err)
	}

	defer tx2.Rollback()

	_, err = tx1.Exec("SELECT balance FROM accounts WHERE id = 1")
	if err != nil {
		t.Fatalf("failed select balance from accounts, %+v", err)
	}

	_, err = tx2.Exec("SELECT balance FROM accounts WHERE id = 1")
	if err != nil {
		t.Fatalf("failed select balance from accounts, %+v", err)
	}

	_, err = tx1.Exec("UPDATE accounts SET balance = 1500 WHERE id = 1;")
	if err != nil {
		t.Fatalf("failed update balance from accounts, %+v", err)
	}

	err = tx1.Commit()
	if err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	_, err = tx2.Exec("UPDATE accounts SET balance = 2000 WHERE id = 1")

	driverErr := pgxtx.Driver().Error(err)

	if !errors.Is(driverErr, err) {
		t.Fatalf("invalid error wrapping, original error was erased, original: %+v, driverErr: %+v", err, driverErr)
	}

	if !errors.Is(driverErr, tx.ErrSerialization) {
		t.Fatalf("expected serialization error, actual %+v", err)
	}
}
