package config

import (
	"flag"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/Nekrasov-Sergey/metrics-collector/pkg/utils"
)

// ServerConfig содержит конфигурацию сервера метрик.
type ServerConfig struct {
	Addr            string         `env:"ADDRESS"`
	StoreInterval   SecondDuration `env:"STORE_INTERVAL"`
	FileStoragePath string         `env:"FILE_STORAGE_PATH"`
	Restore         bool           `env:"RESTORE"`
	DatabaseDSN     string         `env:"DATABASE_DSN"`
	Key             string         `env:"KEY"`
	AuditFile       string         `env:"AUDIT_FILE"`
	AuditURL        string         `env:"AUDIT_URL"`
}

func NewServerConfig(logger zerolog.Logger) (*ServerConfig, error) {
	addr := NetAddress{
		Host: "localhost",
		Port: 8080,
	}
	flag.Var(&addr, "a", "адрес HTTP-сервера")

	storeInterval := SecondDuration(300 * time.Second)
	flag.Var(&storeInterval, "i", "частота сохранения показаний сервера на диск в секундах")

	fileStoragePath := flag.String("f", "./internal/server/repository/saved_data/metrics.json", "путь до файла, куда сохраняются текущие значения")

	restore := flag.Bool("r", false, "следует ли загружать значения из указанного файла при старте сервера")

	databaseDSN := flag.String("d", "", "адрес подключения к БД")

	key := flag.String("k", "", "ключ для вычисления хеша")

	auditFile := flag.String("audit-file", "", "путь к файлу, в который сохраняются логи аудита")

	auditURL := flag.String("audit-url", "", "полный URL, по которому отправляются логи аудита")

	flag.Parse()

	cfg := ServerConfig{
		Addr:            addr.String(),
		StoreInterval:   storeInterval,
		FileStoragePath: utils.Deref(fileStoragePath),
		Restore:         utils.Deref(restore),
		DatabaseDSN:     utils.Deref(databaseDSN),
		Key:             utils.Deref(key),
		AuditFile:       utils.Deref(auditFile),
		AuditURL:        utils.Deref(auditURL),
	}

	if err := env.Parse(&cfg); err != nil {
		return nil, errors.Wrap(err, "не удалось распарсить переменные окружения в конфиг")
	}

	logger.Info().
		Str("address", cfg.Addr).
		Str("store_interval", cfg.StoreInterval.String()).
		Str("file_storage_path", cfg.FileStoragePath).
		Bool("restore", cfg.Restore).
		Str("database_dsn", cfg.DatabaseDSN).
		Str("key", cfg.Key).
		Str("auditFile", cfg.AuditFile).
		Str("auditURL", cfg.AuditURL).
		Msg("Загружена конфигурация сервера")

	return &cfg, nil
}
