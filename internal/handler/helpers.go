package handler

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"easy-clock/internal/i18n"
	"easy-clock/internal/middleware"
)

type templComponent interface {
	Render(ctx context.Context, w io.Writer) error
}

func renderTempl(c *gin.Context, component templComponent) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		slog.Error("render template", "method", c.Request.Method, "path", c.Request.URL.Path, "error", err)
		c.Status(http.StatusInternalServerError)
	}
}

func isHTMX(c *gin.Context) bool {
	return c.GetHeader("HX-Request") == "true"
}

func sessionUserID(c *gin.Context) string {
	if v, exists := c.Get(middleware.SessionUserKey); exists {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

func lang(c *gin.Context) i18n.Lang {
	if v, ok := c.Get(middleware.LangKey); ok {
		if l, ok := v.(i18n.Lang); ok {
			return l
		}
	}
	return i18n.EN
}

func apiErr(c *gin.Context, err error) {
	slog.Error("api error", "method", c.Request.Method, "path", c.Request.URL.Path, "error", err)
	msg, status := i18n.DomainError(err, lang(c))
	c.JSON(status, gin.H{"error": msg})
}

func bindErr(c *gin.Context, err error) {
	slog.Warn("bad request", "method", c.Request.Method, "path", c.Request.URL.Path, "error", err)
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}
