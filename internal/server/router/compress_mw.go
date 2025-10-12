package router

import (
	"compress/gzip"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/Nekrasov-Sergey/metrics-collector/pkg/logger"
)

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (r *compressReader) Read(p []byte) (n int, err error) {
	return r.zr.Read(p)
}

func (r *compressReader) Close() error {
	if err := r.r.Close(); err != nil {
		return err
	}
	return r.zr.Close()
}

type compressWriter struct {
	gin.ResponseWriter
	zw           *gzip.Writer
	isCompressed bool
}

func newCompressWriter(rw gin.ResponseWriter) *compressWriter {
	return &compressWriter{ResponseWriter: rw}
}

func (w *compressWriter) Write(data []byte) (int, error) {
	ct := w.ResponseWriter.Header().Get("Content-Type")
	if strings.HasPrefix(ct, "application/json") || strings.HasPrefix(ct, "text/html") {
		if !w.isCompressed {
			w.ResponseWriter.Header().Set("Content-Encoding", "gzip")
			w.ResponseWriter.Header().Set("Vary", "Accept-Encoding")
			w.zw = gzip.NewWriter(w.ResponseWriter)
			w.isCompressed = true
		}
		return w.zw.Write(data)
	}
	return w.ResponseWriter.Write(data)
}

func (w *compressWriter) Close() error {
	if w.zw != nil {
		return w.zw.Close()
	}
	return nil
}

// CompressMiddleware добавляет сжатие и декомпрессию данных в gin
func CompressMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверяем, что клиент отправил серверу сжатые данные в формате gzip
		if strings.Contains(c.GetHeader("Content-Encoding"), "gzip") {
			cr, err := newCompressReader(c.Request.Body)
			if err != nil {
				logger.InternalServerError(c, err)
				return
			}
			defer func(cr *compressReader) {
				if err := cr.Close(); err != nil {
					log.Error().Err(err).Msg("Не удалось закрыть compressReader")
				}
			}(cr)
			c.Request.Body = cr
		}

		// Проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			cw := newCompressWriter(c.Writer)
			defer func(cw *compressWriter) {
				if err := cw.Close(); err != nil {
					log.Error().Err(err).Msg("Не удалось закрыть compressWriter")
				}
			}(cw)
			c.Writer = cw
		}

		c.Next()
	}
}
