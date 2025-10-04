package db

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/avast/retry-go"
	zerologadapter "github.com/jackc/pgx-zerolog"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	OrderAsc  = "ASC"
	OrderDesc = "DESC NULLS LAST"

	UniqueViolation = "23505"
)

type txFn func(pgx.Tx) error

type DBService interface {
	Query(q string, args ...any) (pgx.Rows, error)
	QueryRow(q string, args ...any) pgx.Row
	Commit(tx pgx.Tx, fn txFn) error
	Begin(ctx context.Context) (pgx.Tx, error)
}

type impl struct {
	db *pgxpool.Pool
}

var Service DBService

func init() {
	Service = New()
}

func New() DBService {
	pgxCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	pgxCfg.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   zerologadapter.NewLogger(log.Level(zerolog.InfoLevel)),
		LogLevel: tracelog.LogLevelTrace,
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), pgxCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	go Ping(pool)

	return &impl{db: pool}
}

func (s *impl) Query(q string, args ...any) (pgx.Rows, error) {
	return s.db.Query(context.Background(), q, args...)
}

func (s *impl) QueryRow(q string, args ...any) pgx.Row {
	return s.db.QueryRow(context.Background(), q, args...)
}

func (s *impl) Begin(ctx context.Context) (pgx.Tx, error) {
	return s.db.Begin(context.Background())
}

func (s *impl) Commit(tx pgx.Tx, fn txFn) error {
	// Wrap for retry.
	return retry.Do(
		func() (err error) {
			// Check if tx is given.
			given := tx != nil

			if !given {
				tx, err = s.db.Begin(context.Background())
				if err != nil {
					return
				}
			}

			defer func() {
				if !given && err != nil {
					if err := tx.Rollback(context.Background()); err != nil {
						log.Error().Err(err).Msg("db tx rollback failed")
					}
				}
			}()

			if err = fn(tx); err != nil {
				return
			}

			if !given {
				if err = tx.Commit(context.Background()); err != nil {
					return
				}
			}

			return
		},
		retry.RetryIf(func(err error) bool {
			var pgErr *pgconn.PgError
			// 40P01: deadlock_detected
			if errors.As(err, &pgErr) && pgErr.Code == "40P01" {
				log.Warn().Err(err).Msg("having error, retrying transaction")
				return true
			}
			return false
		}),
		retry.LastErrorOnly(true),
	)
}

func QuoteString(s string) string {
	return "'" + strings.Replace(s, "'", "''", -1) + "'"
}

func Ping(db *pgxpool.Pool) {
	ticker := time.NewTicker(time.Duration(cfg.PingInterval) * time.Second)

	defer ticker.Stop()

	for range ticker.C {
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := db.Ping(ctx); err != nil {
				log.Fatal().Msgf("Could not contact DB, terminating: %s", err)
			}
		}()
	}
}

func Order(asc bool) string {
	if asc {
		return OrderAsc
	}

	return OrderDesc
}
