package pg

import (
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/sshirox/isaac/internal/metric"
	"github.com/sshirox/isaac/internal/storage"
	"log/slog"
	"time"
)

const (
	timeout = 10 * time.Second
)

func Open(driver, addr string) (*sql.DB, error) {
	db, err := sql.Open(driver, addr)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func Ping(db *sql.DB) error {
	err := db.Ping()
	if err != nil {
		return err
	}
	return nil
}

func Bootstrap(db *sql.DB, ctx context.Context) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	tx.ExecContext(ctx, `CREATE SCHEMA IF NOT EXISTS observability`)

	tx.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS observability.metrics (
            id SERIAL PRIMARY KEY,
            type character varying(255) NOT NULL,
            name character varying(255) NOT NULL UNIQUE,
            value double precision,
            delta int
        )
    `)

	return tx.Commit()
}

func RunSaver(db *sql.DB, ms *storage.MemStorage, interval int64, stopChan chan struct{}) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	for {
		select {
		case <-ticker.C:
			if err := upsertMetrics(db, ms); err != nil {
				slog.Error("save metrics to database", "err", err)
			}
		case <-stopChan:
			slog.Info("stop database saver")
			return
		}
	}
}

func upsertMetrics(db *sql.DB, s *storage.MemStorage) error {
	gauges := s.ReceiveAllGauges()
	counters := s.ReceiveAllCounters()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for name, value := range gauges {
		query := `
            INSERT INTO observability.metrics (name, type, value)
                VALUES ($1, $2, $3)
                ON CONFLICT (name)
                DO UPDATE SET name = $1, type = $2, value = $3`
		err := ExecuteContextWithRetry(ctx, db, query, name, metric.GaugeMetricType, value)
		if err != nil {
			slog.Error("upsert gauge", "err", err)
			return errors.Wrap(err, "upsert gauge")
		}
	}

	for name, delta := range counters {
		query := `
            INSERT INTO observability.metrics (name, type, delta)
                VALUES ($1, $2, $3)
                ON CONFLICT (name)
                DO UPDATE SET name = $1, type = $2, delta = $3`
		err := ExecuteContextWithRetry(ctx, db, query, name, metric.CounterMetricType, delta)
		if err != nil {
			slog.Error("upsert counter", "err", err)
			return errors.Wrap(err, "upsert counter")
		}
	}

	return nil
}

func ListMetrics(db *sql.DB, ms *storage.MemStorage) error {
	err := listGauges(db, ms)
	if err != nil {
		return errors.Wrap(err, "read metrics")
	}

	err = listCounters(db, ms)
	if err != nil {
		return errors.Wrap(err, "read metrics")
	}

	return nil
}

func listGauges(db *sql.DB, ms *storage.MemStorage) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var name string
	var value float64

	query := "SELECT name, value FROM observability.metrics WHERE type = 'gauge'"
	rows, err := QueryContextWithRetry(ctx, db, query)
	if err != nil {
		slog.Error("read gauge", "err", err)
		return errors.Wrap(err, "read gauge")
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&name, &value)
		if err != nil {
			slog.Error("scan gauge", "err", err)
			return errors.Wrap(err, "scan gauge")
		}
		ms.UpdateGauge(name, value)
	}

	err = rows.Err()
	if err != nil {
		slog.Error("read gauge", "err", err)
		return errors.Wrap(err, "read gauge")
	}

	return nil
}

func listCounters(db *sql.DB, ms *storage.MemStorage) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var name string
	var delta int64

	query := "SELECT name, delta FROM observability.metrics WHERE type = 'counter'"
	rows, err := QueryContextWithRetry(ctx, db, query)
	if err != nil {
		slog.Error("read counter", "err", err)
		return errors.Wrap(err, "read counter")
	}

	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&name, &delta)
		if err != nil {
			slog.Error("scan counter", "err", err)
			return errors.Wrap(err, "scan counter")
		}
		ms.UpdateCounter(name, delta)
	}

	err = rows.Err()
	if err != nil {
		slog.Error("select counter", "err", err)
		return errors.Wrap(err, "select counter")
	}

	return nil
}
