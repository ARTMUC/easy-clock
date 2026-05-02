package handler

import (
	"github.com/gin-gonic/gin"

	"starter/internal/views/pages"
)

type DashboardHandler struct{}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{}
}

func (h *DashboardHandler) ShowDashboard(c *gin.Context) {
	userID := sessionUserID(c)
	renderTempl(c, pages.DashboardPage(userID))
}
