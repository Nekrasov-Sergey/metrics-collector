package audit

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
)

type FileObserver struct {
	mu   *sync.Mutex
	file *os.File
}

func NewFileObserver(path string) (*FileObserver, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, errors.Wrap(err, "не удалось создать директорию для аудита")
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, errors.Wrap(err, "не удалось создать файл для аудита")
	}

	return &FileObserver{
		mu:   &sync.Mutex{},
		file: f,
	}, nil
}

func (o *FileObserver) Notify(_ context.Context, event *types.AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return errors.Wrap(err, "не удалось спарсить событие для аудита")
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	if _, err = o.file.WriteString(string(data) + "\n"); err != nil {
		return errors.Wrap(err, "не удалось записать событие в файл аудита")
	}

	return nil
}
