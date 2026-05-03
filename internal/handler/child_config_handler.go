package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"easy-clock/internal/app"
	"easy-clock/internal/views/pages"
)

type ChildConfigHandler struct {
	childSvc    *app.ChildService
	profileSvc  *app.ProfileService
	scheduleSvc *app.ScheduleService
}

func NewChildConfigHandler(childSvc *app.ChildService, profileSvc *app.ProfileService, scheduleSvc *app.ScheduleService) *ChildConfigHandler {
	return &ChildConfigHandler{childSvc: childSvc, profileSvc: profileSvc, scheduleSvc: scheduleSvc}
}

func (h *ChildConfigHandler) Show(c *gin.Context) {
	ctx := c.Request.Context()
	userID := sessionUserID(c)
	childID := c.Param("id")

	child, err := h.childSvc.GetChild(ctx, childID, userID)
	if err != nil {
		apiErr(c, err)
		return
	}
	profiles, err := h.profileSvc.ListProfiles(ctx, childID, userID)
	if err != nil {
		slog.Error("list profiles", "error", err)
	}
	assignments, err := h.scheduleSvc.GetSchedule(ctx, childID, userID)
	if err != nil {
		slog.Error("get schedule", "error", err)
	}
	renderTempl(c, pages.ChildConfigPage(pages.ChildConfigData{
		Child:       child,
		Profiles:    profiles,
		Assignments: assignments,
	}, lang(c)))
}

func (h *ChildConfigHandler) DeleteChild(c *gin.Context) {
	if err := h.childSvc.RemoveChild(c.Request.Context(), c.Param("id"), sessionUserID(c)); err != nil {
		slog.Error("delete child", "error", err)
	}
	c.Redirect(http.StatusSeeOther, "/dashboard")
}

func (h *ChildConfigHandler) CreateProfile(c *gin.Context) {
	childID := c.Param("id")
	name := c.PostForm("name")
	color := c.PostForm("color")
	profile, err := h.profileSvc.CreateProfile(c.Request.Context(), childID, sessionUserID(c), name, color)
	if err != nil {
		slog.Error("create profile", "error", err)
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/children/%s", childID))
		return
	}
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/profiles/%s", profile.ID))
}

func (h *ChildConfigHandler) SetDefaultProfile(c *gin.Context) {
	childID := c.Param("id")
	profileID := c.PostForm("profile_id")
	if err := h.childSvc.SetDefaultProfile(c.Request.Context(), childID, sessionUserID(c), profileID); err != nil {
		slog.Error("set default profile", "error", err)
	}
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/children/%s", childID))
}

func (h *ChildConfigHandler) AssignScheduleDay(c *gin.Context) {
	ctx := c.Request.Context()
	childID := c.Param("id")
	userID := sessionUserID(c)
	day, err := strconv.Atoi(c.Param("day"))
	if err != nil || day < 0 || day > 6 {
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/children/%s", childID))
		return
	}
	profileID := c.PostForm("profile_id")
	if profileID == "" {
		_ = h.scheduleSvc.ClearDay(ctx, childID, userID, day)
	} else {
		_ = h.scheduleSvc.AssignProfileToDays(ctx, childID, userID, profileID, []int{day})
	}
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/children/%s", childID))
}
