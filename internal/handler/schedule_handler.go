package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"easy-clock/internal/app"
)

type ScheduleHandler struct {
	svc *app.ScheduleService
}

func NewScheduleHandler(svc *app.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{svc: svc}
}

func (h *ScheduleHandler) Get(c *gin.Context) {
	assignments, err := h.svc.GetSchedule(c.Request.Context(), c.Param("childID"), sessionUserID(c))
	if err != nil {
		apiErr(c, err)
		return
	}
	c.JSON(http.StatusOK, assignments)
}

func (h *ScheduleHandler) Assign(c *gin.Context) {
	var body struct {
		ProfileID string `json:"profile_id" binding:"required"`
		Days      []int  `json:"days"       binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		bindErr(c, err)
		return
	}
	if err := h.svc.AssignProfileToDays(c.Request.Context(), c.Param("childID"), sessionUserID(c), body.ProfileID, body.Days); err != nil {
		apiErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ScheduleHandler) ClearDay(c *gin.Context) {
	day, err := strconv.Atoi(c.Param("day"))
	if err != nil {
		slog.Warn("bad request", "method", c.Request.Method, "path", c.Request.URL.Path, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "day must be an integer 0–6"})
		return
	}
	if err := h.svc.ClearDay(c.Request.Context(), c.Param("childID"), sessionUserID(c), day); err != nil {
		apiErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
