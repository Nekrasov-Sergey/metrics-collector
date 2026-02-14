package config

import (
	"flag"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/Nekrasov-Sergey/metrics-collector/pkg/utils"
)

// ServerConfig содержит конфигурацию сервера метрик.
type ServerConfig struct {
	Addr            string         `env:"ADDRESS" json:"address"`
	StoreInterval   SecondDuration `env:"STORE_INTERVAL" json:"store_interval"`
	FileStoragePath string         `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	Restore         bool           `env:"RESTORE" json:"restore"`
	DatabaseDSN     string         `env:"DATABASE_DSN" json:"database_dsn"`
	SignKey         string         `env:"KEY" json:"key"`
	AuditFile       string         `env:"AUDIT_FILE" json:"audit_file"`
	AuditURL        string         `env:"AUDIT_URL" json:"audit_url"`
	CryptoKey       string         `env:"CRYPTO_KEY" json:"crypto_key"`
}

func NewServerConfig(logger zerolog.Logger) (*ServerConfig, error) {
	cfg := ServerConfig{
		Addr:            "localhost:8080",
		StoreInterval:   SecondDuration(300 * time.Second),
		FileStoragePath: "./internal/server/repository/saved_data/metrics.json",
		Restore:         false,
	}

	var configPath string
	flag.StringVar(&configPath, "c", "", "имя файла конфигурации")
	flag.StringVar(&configPath, "config", "", "имя файла конфигурации")

	var addr NetAddress
	flag.Var(&addr, "a", "адрес HTTP-сервера")

	var storeInterval SecondDuration
	flag.Var(&storeInterval, "i", "частота сохранения показаний сервера")

	fileStoragePath := flag.String("f", "", "путь до файла хранения")
	restore := flag.Bool("r", false, "загружать данные при старте")
	databaseDSN := flag.String("d", "", "адрес подключения к БД")
	signKey := flag.String("k", "", "ключ для вычисления хеша")
	auditFile := flag.String("audit-file", "", "путь к файлу аудита")
	auditURL := flag.String("audit-url", "", "URL для отправки логов аудита")
	cryptoKey := flag.String("crypto-key", "", "путь до файла с приватным ключом")

	flag.Parse()

	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, errors.Wrap(err, "не удалось прочитать файл с конфигом")
		}
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, errors.Wrap(err, "не удалось распарсить файл с конфигом")
		}
	}

	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "a":
			cfg.Addr = addr.String()
		case "i":
			cfg.StoreInterval = storeInterval
		case "f":
			cfg.FileStoragePath = utils.Deref(fileStoragePath)
		case "r":
			cfg.Restore = utils.Deref(restore)
		case "d":
			cfg.DatabaseDSN = utils.Deref(databaseDSN)
		case "k":
			cfg.SignKey = utils.Deref(signKey)
		case "audit-file":
			cfg.AuditFile = utils.Deref(auditFile)
		case "audit-url":
			cfg.AuditURL = utils.Deref(auditURL)
		case "crypto-key":
			cfg.CryptoKey = utils.Deref(cryptoKey)
		}
	})

	if err := env.Parse(&cfg); err != nil {
		return nil, errors.Wrap(err, "не удалось распарсить переменные окружения в конфиг")
	}

	logger.Info().
		Str("address", cfg.Addr).
		Str("store_interval", cfg.StoreInterval.String()).
		Str("file_storage_path", cfg.FileStoragePath).
		Bool("restore", cfg.Restore).
		Str("database_dsn", cfg.DatabaseDSN).
		Str("key", cfg.SignKey).
		Str("audit_file", cfg.AuditFile).
		Str("audit_url", cfg.AuditURL).
		Str("crypto_key", cfg.CryptoKey).
		Msg("Загружена конфигурация сервера")

	return &cfg, nil
}
