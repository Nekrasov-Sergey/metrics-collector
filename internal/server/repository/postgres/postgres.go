package postgres

import (
	"context"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

	if err := migrateDB(databaseDSN, logger); err != nil {
		return nil, err
	}

	return &Postgres{
		db:     db,
		logger: logger,
	}, nil
}

func (p *Postgres) PingDB(ctx context.Context) error {
	if err := p.db.PingContext(ctx); err != nil {
		return errors.Wrap(err, "БД недоступна")
	}
	return nil
}

func (p *Postgres) CloseDB() {
	if err := p.db.Close(); err != nil {
		p.logger.Error().Err(err).Msg("Не удалось закрыть соединения с БД")
	}
	p.logger.Info().Msg("Закрыто соединение с БД")
}

func migrateDB(databaseDSN string, logger zerolog.Logger) error {
	logger.Info().Msg("Запуск миграций базы данных...")

	m, err := migrate.New("file://migrations", databaseDSN)
	if err != nil {
		return errors.Wrap(err, "не удалось инициализировать миграции")
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info().Msg("Миграции не требуются — база данных уже актуальна")
		} else {
			return errors.Wrap(err, "не удалось применить миграции")
		}
	} else {
		logger.Info().Msg("Миграции успешно применены")
	}

	return nil
}
