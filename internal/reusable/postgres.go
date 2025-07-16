package reusable

import (
	postgrescontainer "github.com/amidgo/containers/postgres"
	pgrunner "github.com/amidgo/containers/postgres/runner"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var pgreusable = postgrescontainer.NewReusable(
	pgrunner.RunContainer(nil),
)

func Postgres() *postgrescontainer.Reusable {
	return pgreusable
}
