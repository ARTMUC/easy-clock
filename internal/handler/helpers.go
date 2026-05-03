package handler

import (
	"context"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"easy-clock/internal/middleware"
)

type templComponent interface {
	Render(ctx context.Context, w io.Writer) error
}

func renderTempl(c *gin.Context, component templComponent) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
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
