package errcodes

import (
	"github.com/pkg/errors"
)

var (
	ErrMetricNotFound = errors.New("метрика не найдена")
)
