package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"easy-clock/internal/app"
)

type EventHandler struct {
	svc *app.EventService
}

func NewEventHandler(svc *app.EventService) *EventHandler {
	return &EventHandler{svc: svc}
}

func (h *EventHandler) List(c *gin.Context) {
	from, err := time.Parse("2006-01-02", c.Query("from"))
	if err != nil {
		slog.Warn("bad request", "method", c.Request.Method, "path", c.Request.URL.Path, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "from must be YYYY-MM-DD"})
		return
	}
	to, err := time.Parse("2006-01-02", c.Query("to"))
	if err != nil {
		slog.Warn("bad request", "method", c.Request.Method, "path", c.Request.URL.Path, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "to must be YYYY-MM-DD"})
		return
	}
	events, err := h.svc.ListEvents(c.Request.Context(), c.Param("id"), sessionUserID(c), from, to)
	if err != nil {
		apiErr(c, err)
		return
	}
	c.JSON(http.StatusOK, events)
}

func (h *EventHandler) Create(c *gin.Context) {
	in, ok := bindEventInput(c)
	if !ok {
		return
	}
	event, err := h.svc.CreateEvent(c.Request.Context(), c.Param("id"), sessionUserID(c), in)
	if err != nil {
		apiErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, event)
}

func (h *EventHandler) Update(c *gin.Context) {
	in, ok := bindEventInput(c)
	if !ok {
		return
	}
	event, err := h.svc.UpdateEvent(c.Request.Context(), c.Param("id"), sessionUserID(c), in)
	if err != nil {
		apiErr(c, err)
		return
	}
	c.JSON(http.StatusOK, event)
}

func (h *EventHandler) Delete(c *gin.Context) {
	if err := h.svc.DeleteEvent(c.Request.Context(), c.Param("id"), sessionUserID(c)); err != nil {
		apiErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type eventActivityBody struct {
	Emoji     string `json:"emoji"`
	Label     string `json:"label"      binding:"required"`
	ImagePath string `json:"image_path" binding:"required"`
	FromHour  int    `json:"from_hour"`
	ToHour    int    `json:"to_hour"    binding:"required"`
}

func bindEventInput(c *gin.Context) (app.CreateEventInput, bool) {
	var body struct {
		Date       string              `json:"date"       binding:"required"`
		FromTime   string              `json:"from_time"  binding:"required"`
		ToTime     string              `json:"to_time"    binding:"required"`
		Label      string              `json:"label"      binding:"required"`
		Emoji      string              `json:"emoji"`
		ProfileID  string              `json:"profile_id"`
		Activities []eventActivityBody `json:"activities"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		bindErr(c, err)
		return app.CreateEventInput{}, false
	}
	date, err := time.Parse("2006-01-02", body.Date)
	if err != nil {
		slog.Warn("bad request", "method", c.Request.Method, "path", c.Request.URL.Path, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "date must be YYYY-MM-DD"})
		return app.CreateEventInput{}, false
	}
	in := app.CreateEventInput{
		Date:      date,
		FromTime:  body.FromTime,
		ToTime:    body.ToTime,
		Label:     body.Label,
		Emoji:     body.Emoji,
		ProfileID: body.ProfileID,
	}
	for _, a := range body.Activities {
		in.Activities = append(in.Activities, app.EventActivityInput{
			Emoji:     a.Emoji,
			Label:     a.Label,
			ImagePath: a.ImagePath,
			FromHour:  a.FromHour,
			ToHour:    a.ToHour,
		})
	}
	return in, true
}
