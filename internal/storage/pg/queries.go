package pg

import (
	"context"
	"database/sql"
	"log/slog"

	errs "github.com/sshirox/isaac/internal/errors"
	"github.com/sshirox/isaac/internal/retries"
)

type ExecContext interface {
	ExecContext(ctx context.Context, qry string, args ...interface{}) (sql.Result, error)
}

func ExecuteContextWithRetry(ctx context.Context, exec ExecContext, qry string, agrs ...interface{}) error {
	err := retries.Retry(func() error {
		_, execErr := exec.ExecContext(ctx, qry, agrs...)
		if execErr != nil {
			if errs.IsRetryPGErr(execErr) {
				slog.ErrorContext(ctx, "postgres retry error", "err", execErr)

				return errs.ErrRetryPG
			}

			slog.ErrorContext(ctx, "postgres non retry error", "err", execErr)

			return errs.ErrNonRetryPG
		}

		return nil
	})

	if err != nil {
		slog.ErrorContext(ctx, "execute query", "err", err)

		return err
	}

	return nil
}

func QueryContextWithRetry(ctx context.Context, db *sql.DB, qry string, args ...interface{}) (*sql.Rows, error) {
	var rows *sql.Rows
	var err error

	err = retries.Retry(func() error {
		rows, err = db.QueryContext(ctx, qry, args...)
		if err != nil {
			if errs.IsRetryPGErr(err) {
				slog.ErrorContext(ctx, "postgres retry error", "err", err)

				return errs.ErrRetryPG
			}
			slog.ErrorContext(ctx, "postgres non retry error", "err", err)

			return errs.ErrNonRetryPG
		}

		if rows.Err() != nil {
			slog.ErrorContext(ctx, "select rows", "err", rows.Err())

			return errs.ErrNonRetry
		}

		return nil
	})

	if err != nil {
		slog.ErrorContext(ctx, "execute query", "err", err)

		return nil, err
	}

	return rows, nil
}
