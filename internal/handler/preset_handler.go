package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"easy-clock/internal/app"
)

type PresetHandler struct {
	profileSvc *app.ProfileService
}

func NewPresetHandler(profileSvc *app.ProfileService) *PresetHandler {
	return &PresetHandler{profileSvc: profileSvc}
}

func (h *PresetHandler) List(c *gin.Context) {
	presets, err := h.profileSvc.ListPresets(c.Request.Context())
	if err != nil {
		slog.Error("api error", "method", c.Request.Method, "path", c.Request.URL.Path, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch presets"})
		return
	}
	c.JSON(http.StatusOK, presets)
}
