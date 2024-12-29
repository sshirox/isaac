package errors

import (
	"errors"
	"slices"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)

var (
	ErrConnection = errors.New("connection error")
	ErrNonRetry   = errors.New("non retry error")
	ErrNonRetryPG = errors.New("non retry postgres error")
	ErrRetryPG    = errors.New("retry postgres error")
	ErrServer     = errors.New("server error")
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
