package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"easy-clock/internal/app"
	"easy-clock/internal/views/pages"
)

type DashboardHandler struct {
	childSvc *app.ChildService
}

func NewDashboardHandler(childSvc *app.ChildService) *DashboardHandler {
	return &DashboardHandler{childSvc: childSvc}
}

func (h *DashboardHandler) ShowDashboard(c *gin.Context) {
	children, err := h.childSvc.ListChildren(c.Request.Context(), sessionUserID(c))
	if err != nil {
		slog.Error("list children", "error", err)
		children = nil
	}
	renderTempl(c, pages.DashboardPage(children, lang(c)))
}

func (h *DashboardHandler) CreateChild(c *gin.Context) {
	name := c.PostForm("name")
	timezone := c.PostForm("timezone")
	child, err := h.childSvc.AddChild(c.Request.Context(), sessionUserID(c), name, timezone)
	if err != nil {
		slog.Error("create child", "error", err)
		children, _ := h.childSvc.ListChildren(c.Request.Context(), sessionUserID(c))
		renderTempl(c, pages.DashboardPage(children, lang(c)))
		return
	}
	c.Redirect(http.StatusSeeOther, "/children/"+child.ID)
}
