package database

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

func IsNotFoundError(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func IsConstraintViolationError(err error) bool {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return false
	}
	return pqErr.Code.Class() == "23"
}
