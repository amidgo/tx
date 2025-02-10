package pgxtx

import (
	"errors"

	"github.com/amidgo/tx"
	"github.com/jackc/pgx/v5/pgconn"
)

type driver struct{}

func Driver() tx.Driver {
	return driver{}
}

func (driver) Error(err error) error {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "40P01", "40001":
			err = errors.Join(tx.ErrSerialization, err)
		}
	}

	return err
}
