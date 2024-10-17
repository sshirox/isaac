package errors

import (
	"errors"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"slices"
)

var (
	ConnectionErr = errors.New("connection error")
	NonRetryErr   = errors.New("non retry error")
	NonRetryPGErr = errors.New("non retry postgres error")
	RetryPGErr    = errors.New("retry postgres error")
	ServerErr     = errors.New("server error")
)

var (
	pgErrors = []string{
		pgerrcode.AdminShutdown,
		pgerrcode.ConnectionFailure,
		pgerrcode.ConnectionException,
		pgerrcode.DeadlockDetected,
		pgerrcode.SerializationFailure,
	}
)

func IsRetryPGErr(pgErr error) bool {
	var err *pgconn.PgError
	ok := errors.As(pgErr, &err)
	if ok && slices.Contains(pgErrors, err.Code) {
		return true
	}

	return false
}
