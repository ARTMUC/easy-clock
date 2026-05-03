package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"easy-clock/internal/app"
	"easy-clock/internal/views/pages"
)

type ClockHandler struct {
	svc *app.ClockService
}

func NewClockHandler(svc *app.ClockService) *ClockHandler {
	return &ClockHandler{svc: svc}
}

// Show serves the clock HTML page. The JS on the page polls /api/clock/:token.
func (h *ClockHandler) Show(c *gin.Context) {
	renderTempl(c, pages.ClockPage(c.Param("token"), lang(c)))
}

// State returns the current ClockState as JSON.
func (h *ClockHandler) State(c *gin.Context) {
	state, err := h.svc.Resolve(c.Request.Context(), c.Param("token"), time.Now().UTC())
	if err != nil {
		apiErr(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}
