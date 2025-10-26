package postgres

import (
	"context"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type Postgres struct {
	db     *sqlx.DB
	logger zerolog.Logger
}

func New(databaseDSN string, logger zerolog.Logger) (*Postgres, error) {
	db, err := sqlx.Connect("pgx", databaseDSN)
	if err != nil {
		return nil, errors.Wrap(err, "не удалось подключиться к БД")
	}

	logger.Info().Msg("Установлено подключение к БД")

	return &Postgres{
		db:     db,
		logger: logger,
	}, nil
}

func (p *Postgres) Ping(ctx context.Context) error {
	if err := p.db.PingContext(ctx); err != nil {
		return errors.Wrap(err, "БД недоступна")
	}
	return nil
}

func (p *Postgres) Close() error {
	if err := p.db.Close(); err != nil {
		return errors.Wrap(err, "не удалось закрыть соединения с БД")
	}
	p.logger.Info().Msg("Закрыто соединение с БД")
	return nil
}
