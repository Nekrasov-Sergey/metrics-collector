package serverconfig

import (
	"flag"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/config"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/utils"
)

type Config struct {
	Addr            string                `env:"ADDRESS"`
	StoreInterval   config.SecondDuration `env:"STORE_INTERVAL"`
	FileStoragePath string                `env:"FILE_STORAGE_PATH"`
	Restore         bool                  `env:"RESTORE"`
}

func New() (*Config, error) {
	addr := config.NetAddress{
		Host: "localhost",
		Port: 8080,
	}
	flag.Var(&addr, "a", "адрес HTTP-сервера")

	storeInterval := config.SecondDuration(300 * time.Second)
	flag.Var(&storeInterval, "i", "частота сохранения показаний сервера на диск в секундах")

	fileStoragePath := flag.String("f", "./internal/server/repository/saved_data/metrics.json", "путь до файла, куда сохраняются текущие значения")

	restore := flag.Bool("r", false, "следует ли загружать значения из указанного файла при старте сервера")

	flag.Parse()

	cfg := Config{
		Addr:            addr.String(),
		StoreInterval:   storeInterval,
		FileStoragePath: utils.Deref(fileStoragePath),
		Restore:         utils.Deref(restore),
	}

	if err := env.Parse(&cfg); err != nil {
		return nil, errors.Wrap(err, "не удалось распарсить переменные окружения в конфиг")
	}

	return &cfg, nil
}
