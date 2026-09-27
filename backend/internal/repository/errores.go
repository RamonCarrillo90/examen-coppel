package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// EsViolacionFK indica si el error es una violacion a una llave foranea codigo (23503)
func EsViolacionFK(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}

func EsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
