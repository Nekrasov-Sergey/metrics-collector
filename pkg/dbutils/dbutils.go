package dbutils

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"go.uber.org/multierr"
)

func NamedGet(ctx context.Context, db sqlx.ExtContext, dest any, q string, arg any) error {
	nq, args, err := db.BindNamed(q, arg)
	if err != nil {
		return errors.Wrap(err, "не удалось подготовить SQL-запрос (NamedGet)")
	}

	return runWithRetries(ctx, func() error {
		if err := sqlx.GetContext(ctx, db, dest, nq, args...); err != nil {
			return errors.Wrap(err, "не удалось выполнить SQL-запрос (NamedGet)")
		}
		return nil
	})
}

func NamedSelect(ctx context.Context, db sqlx.ExtContext, dest any, q string, arg any) error {
	nq, args, err := db.BindNamed(q, arg)
	if err != nil {
		return errors.Wrap(err, "не удалось подготовить SQL-запрос (NamedSelect)")
	}

	return runWithRetries(ctx, func() error {
		if err := sqlx.SelectContext(ctx, db, dest, nq, args...); err != nil {
			return errors.Wrap(err, "не удалось выполнить SQL-запрос (NamedSelect)")
		}
		return nil
	})
}

func NamedExec(ctx context.Context, db sqlx.ExtContext, q string, arg any) error {
	nq, args, err := db.BindNamed(q, arg)
	if err != nil {
		return errors.Wrap(err, "не удалось подготовить SQL-запрос (NamedExec)")
	}

	return runWithRetries(ctx, func() error {
		if _, err := db.ExecContext(ctx, nq, args...); err != nil {
			return errors.Wrap(err, "не удалось выполнить SQL-запрос (NamedExec)")
		}
		return nil
	})
}

func runWithRetries(ctx context.Context, fn func() error) (err error) {
	delays := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second, 0}
	for i, delay := range delays {
		if err = fn(); err == nil {
			return nil
		}

		if !isConnectionError(err) {
			return errors.Wrap(err, "не удалось выполнить SQL-запрос (NamedSelect)")
		}

		log.Error().Err(err).Msgf("Ошибка соединения c PostgreSQL, попытка №%d", i+1)
		if delay > 0 {
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				log.Error().Msg("Запрос отменён контекстом во время ожидания")
				return err
			case <-timer.C:
			}
		}
	}

	log.Error().Msg("Все попытки подключения исчерпаны")
	return err
}

func isConnectionError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.ConnectionException, pgerrcode.ConnectionDoesNotExist, pgerrcode.ConnectionFailure,
			pgerrcode.SQLClientUnableToEstablishSQLConnection, pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection,
			pgerrcode.TransactionResolutionUnknown, pgerrcode.ProtocolViolation, pgerrcode.SerializationFailure:
			return true
		}
	}
	return false
}

type DB interface {
	BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error)
}

type TxFunc func(tx *sqlx.Tx) error

func WrapTxx(ctx context.Context, db DB, opts *sql.TxOptions, f TxFunc) (err error) {
	tx, err := db.BeginTxx(ctx, opts)
	if err != nil {
		return errors.Wrap(err, "не удалось начать транзакцию")
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = multierr.Append(err, errors.Wrapf(rbErr, "не удалось выполнить Rollback"))
			}
			return
		}

		if commitErr := tx.Commit(); commitErr != nil {
			err = errors.Wrap(commitErr, "не удалось зафиксировать транзакцию")
		}
	}()

	return f(tx)
}
