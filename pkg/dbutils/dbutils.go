package dbutils

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func NamedGet(ctx context.Context, db *sqlx.DB, dest any, q string, arg any) error {
	nq, args, err := db.BindNamed(q, arg)
	if err != nil {
		return errors.Wrap(err, "не удалось подготовить SQL-запрос (NamedGet)")
	}

	if err = db.GetContext(ctx, dest, nq, args...); err != nil {
		return errors.Wrap(err, "не удалось выполнить SQL-запрос (NamedGet)")
	}

	return nil
}

func NamedSelect(ctx context.Context, db *sqlx.DB, dest any, q string, arg any) error {
	nq, args, err := db.BindNamed(q, arg)
	if err != nil {
		return errors.Wrap(err, "не удалось подготовить SQL-запрос (NamedSelect)")
	}

	if err := db.SelectContext(ctx, dest, nq, args...); err != nil {
		return errors.Wrap(err, "не удалось выполнить SQL-запрос (NamedSelect)")
	}

	return nil
}

func NamedExec(ctx context.Context, db *sqlx.DB, q string, arg any) error {
	nq, args, err := db.BindNamed(q, arg)
	if err != nil {
		return errors.Wrap(err, "не удалось подготовить SQL-запрос (NamedExec)")
	}

	if _, err := db.ExecContext(ctx, nq, args...); err != nil {
		return errors.Wrap(err, "не удалось выполнить SQL-запрос (NamedExec)")
	}

	return nil
}
