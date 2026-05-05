package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"easy-clock/internal/app"
	"easy-clock/internal/views/pages"
)

type ChildConfigHandler struct {
	childSvc    *app.ChildService
	profileSvc  *app.ProfileService
	scheduleSvc *app.ScheduleService
	eventSvc    *app.EventService
}

func NewChildConfigHandler(
	childSvc *app.ChildService,
	profileSvc *app.ProfileService,
	scheduleSvc *app.ScheduleService,
	eventSvc *app.EventService,
) *ChildConfigHandler {
	return &ChildConfigHandler{
		childSvc:    childSvc,
		profileSvc:  profileSvc,
		scheduleSvc: scheduleSvc,
		eventSvc:    eventSvc,
	}
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
	now := time.Now()
	events, err := h.eventSvc.ListEvents(ctx, childID, userID, now, now.AddDate(0, 0, 30))
	if err != nil {
		slog.Error("list events", "error", err)
	}
	renderTempl(c, pages.ChildConfigPage(pages.ChildConfigData{
		Child:       child,
		Profiles:    profiles,
		Assignments: assignments,
		Events:      events,
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
	profile, err := h.profileSvc.CreateProfile(
		c.Request.Context(), childID, sessionUserID(c),
		c.PostForm("name"), c.PostForm("color"),
	)
	if err != nil {
		slog.Error("create profile", "error", err)
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/children/%s", childID))
		return
	}
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/profiles/%s", profile.ID))
}

func (h *ChildConfigHandler) SetDefaultProfile(c *gin.Context) {
	childID := c.Param("id")
	if err := h.childSvc.SetDefaultProfile(c.Request.Context(), childID, sessionUserID(c), c.PostForm("profile_id")); err != nil {
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

func (h *ChildConfigHandler) CreateEvent(c *gin.Context) {
	ctx := c.Request.Context()
	childID := c.Param("id")
	userID := sessionUserID(c)

	dateStr := c.PostForm("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		slog.Warn("bad event date", "date", dateStr, "error", err)
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/children/%s", childID))
		return
	}
	fromTime := c.PostForm("from_time") + ":00"
	toTime := c.PostForm("to_time") + ":00"

	in := app.CreateEventInput{
		Date:      date,
		FromTime:  fromTime,
		ToTime:    toTime,
		Label:     c.PostForm("label"),
		Emoji:     c.PostForm("emoji"),
		ProfileID: c.PostForm("profile_id"),
	}
	if _, err := h.eventSvc.CreateEvent(ctx, childID, userID, in); err != nil {
		slog.Error("create event", "error", err)
	}
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/children/%s", childID))
}

func (h *ChildConfigHandler) DeleteEvent(c *gin.Context) {
	childID := c.PostForm("child_id")
	if err := h.eventSvc.DeleteEvent(c.Request.Context(), c.Param("id"), sessionUserID(c)); err != nil {
		slog.Error("delete event", "error", err)
	}
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/children/%s", childID))
}

func (h *ChildConfigHandler) UploadAvatar(c *gin.Context) {
	childID := c.Param("id")
	avatarPath := c.PostForm("avatar_path")
	if avatarPath == "" {
		c.Status(http.StatusBadRequest)
		return
	}
	if err := h.childSvc.SetAvatarPath(c.Request.Context(), childID, sessionUserID(c), avatarPath); err != nil {
		slog.Error("set avatar path", "error", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusNoContent)
}
