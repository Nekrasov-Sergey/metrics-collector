package dbutils

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

func NamedGet(ctx context.Context, db sqlx.ExtContext, dest any, q string, arg any) error {
	nq, args, err := db.BindNamed(q, arg)
	if err != nil {
		return errors.Wrap(err, "не удалось подготовить SQL-запрос (NamedGet)")
	}

	if err := sqlx.GetContext(ctx, db, dest, nq, args...); err != nil {
		return errors.Wrap(err, "не удалось выполнить SQL-запрос (NamedGet)")
	}

	return nil
}

func NamedSelect(ctx context.Context, db sqlx.ExtContext, dest any, q string, arg any) error {
	nq, args, err := db.BindNamed(q, arg)
	if err != nil {
		return errors.Wrap(err, "не удалось подготовить SQL-запрос (NamedSelect)")
	}

	if err := sqlx.SelectContext(ctx, db, dest, nq, args...); err != nil {
		return errors.Wrap(err, "не удалось выполнить SQL-запрос (NamedSelect)")
	}

	return nil
}

func NamedExec(ctx context.Context, db sqlx.ExtContext, q string, arg any) error {
	nq, args, err := db.BindNamed(q, arg)
	if err != nil {
		return errors.Wrap(err, "не удалось подготовить SQL-запрос (NamedExec)")
	}

	if _, err := db.ExecContext(ctx, nq, args...); err != nil {
		return errors.Wrap(err, "не удалось выполнить SQL-запрос (NamedExec)")
	}

	return nil
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
