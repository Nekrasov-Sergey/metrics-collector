package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/pkg/errors"
)

func (h *Handler) ping(c *gin.Context) {
	db, err := sqlx.Connect("pgx", h.config.DatabaseDSN)
	if err != nil {
		_ = c.Error(errors.Wrap(err, "не удалось подключиться к БД"))
		c.Status(http.StatusInternalServerError)
		return
	}
	
	if err := db.Close(); err != nil {
		_ = c.Error(errors.Wrap(err, "не удалось закрыть соединение с БД"))
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}
