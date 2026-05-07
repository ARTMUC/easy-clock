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

type ProfileConfigHandler struct {
	profileSvc *app.ProfileService
}

func NewProfileConfigHandler(profileSvc *app.ProfileService) *ProfileConfigHandler {
	return &ProfileConfigHandler{profileSvc: profileSvc}
}

func (h *ProfileConfigHandler) Show(c *gin.Context) {
	profile, err := h.profileSvc.GetProfile(c.Request.Context(), c.Param("id"), sessionUserID(c))
	if err != nil {
		apiErr(c, err)
		return
	}
	renderTempl(c, pages.ProfileConfigPage(profile, lang(c)))
}

func (h *ProfileConfigHandler) Delete(c *gin.Context) {
	childID := c.PostForm("child_id")
	if err := h.profileSvc.DeleteProfile(c.Request.Context(), c.Param("id"), sessionUserID(c)); err != nil {
		slog.Error("delete profile", "error", err)
	}
	if childID != "" {
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/children/%s", childID))
		return
	}
	c.Redirect(http.StatusSeeOther, "/dashboard")
}

func (h *ProfileConfigHandler) AddActivity(c *gin.Context) {
	profileID := c.Param("id")
	fromHour, _ := strconv.Atoi(c.PostForm("from_hour"))
	toHour, _ := strconv.Atoi(c.PostForm("to_hour"))
	sortOrder, _ := strconv.Atoi(c.PostForm("sort_order"))

	in := app.AddActivityInput{
		Emoji:     c.PostForm("emoji"),
		Label:     c.PostForm("label"),
		ImagePath: c.PostForm("image_path"),
		FromHour:  fromHour,
		ToHour:    toHour,
		SortOrder: sortOrder,
	}
	if _, err := h.profileSvc.AddActivity(c.Request.Context(), profileID, sessionUserID(c), in); err != nil {
		slog.Error("add activity", "error", err)
	}
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/profiles/%s", profileID))
}

func (h *ProfileConfigHandler) DeleteActivity(c *gin.Context) {
	profileID := c.PostForm("profile_id")
	if err := h.profileSvc.RemoveActivity(c.Request.Context(), c.Param("id"), sessionUserID(c)); err != nil {
		slog.Error("delete activity", "error", err)
	}
	if profileID != "" {
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/profiles/%s", profileID))
		return
	}
	c.Redirect(http.StatusSeeOther, "/dashboard")
}
