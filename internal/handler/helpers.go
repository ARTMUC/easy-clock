package handler

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"easy-clock/internal/domain"
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

func apiErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, domain.ErrInvalidHourRange),
		errors.Is(err, domain.ErrActivityOverlap),
		errors.Is(err, domain.ErrImageRequired),
		errors.Is(err, domain.ErrInvalidTimeRange),
		errors.Is(err, domain.ErrEventProfileXorActivities),
		errors.Is(err, domain.ErrEmptyName),
		errors.Is(err, domain.ErrInvalidTimezone):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
