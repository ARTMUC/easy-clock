package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"easy-clock/internal/app"
)

type ProfileHandler struct {
	svc *app.ProfileService
}

func NewProfileHandler(svc *app.ProfileService) *ProfileHandler {
	return &ProfileHandler{svc: svc}
}

func (h *ProfileHandler) List(c *gin.Context) {
	profiles, err := h.svc.ListProfiles(c.Request.Context(), c.Param("childID"), sessionUserID(c))
	if err != nil {
		apiErr(c, err)
		return
	}
	c.JSON(http.StatusOK, profiles)
}

func (h *ProfileHandler) Create(c *gin.Context) {
	var body struct {
		Name  string `json:"name"  binding:"required"`
		Color string `json:"color"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	profile, err := h.svc.CreateProfile(c.Request.Context(), c.Param("childID"), sessionUserID(c), body.Name, body.Color)
	if err != nil {
		apiErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, profile)
}

func (h *ProfileHandler) Get(c *gin.Context) {
	profile, err := h.svc.GetProfile(c.Request.Context(), c.Param("id"), sessionUserID(c))
	if err != nil {
		apiErr(c, err)
		return
	}
	c.JSON(http.StatusOK, profile)
}

func (h *ProfileHandler) Update(c *gin.Context) {
	var body struct {
		Name  string `json:"name"  binding:"required"`
		Color string `json:"color" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	profile, err := h.svc.UpdateProfile(c.Request.Context(), c.Param("id"), sessionUserID(c), body.Name, body.Color)
	if err != nil {
		apiErr(c, err)
		return
	}
	c.JSON(http.StatusOK, profile)
}

func (h *ProfileHandler) Delete(c *gin.Context) {
	if err := h.svc.DeleteProfile(c.Request.Context(), c.Param("id"), sessionUserID(c)); err != nil {
		apiErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ProfileHandler) AddActivity(c *gin.Context) {
	var body struct {
		PresetID  string `json:"preset_id"`
		Emoji     string `json:"emoji"`
		Label     string `json:"label"       binding:"required"`
		ImagePath string `json:"image_path"`
		FromHour  int    `json:"from_hour"`
		ToHour    int    `json:"to_hour"     binding:"required"`
		Ring      int    `json:"ring"        binding:"required"`
		SortOrder int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	in := app.AddActivityInput{
		PresetID:  body.PresetID,
		Emoji:     body.Emoji,
		Label:     body.Label,
		ImagePath: body.ImagePath,
		FromHour:  body.FromHour,
		ToHour:    body.ToHour,
		Ring:      body.Ring,
		SortOrder: body.SortOrder,
	}
	activity, err := h.svc.AddActivity(c.Request.Context(), c.Param("id"), sessionUserID(c), in)
	if err != nil {
		apiErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, activity)
}

func (h *ProfileHandler) UpdateActivity(c *gin.Context) {
	var body struct {
		Emoji     string `json:"emoji"`
		Label     string `json:"label"       binding:"required"`
		ImagePath string `json:"image_path"  binding:"required"`
		FromHour  int    `json:"from_hour"`
		ToHour    int    `json:"to_hour"     binding:"required"`
		Ring      int    `json:"ring"        binding:"required"`
		SortOrder int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	in := app.AddActivityInput{
		Emoji:     body.Emoji,
		Label:     body.Label,
		ImagePath: body.ImagePath,
		FromHour:  body.FromHour,
		ToHour:    body.ToHour,
		Ring:      body.Ring,
		SortOrder: body.SortOrder,
	}
	activity, err := h.svc.UpdateActivity(c.Request.Context(), c.Param("id"), sessionUserID(c), in)
	if err != nil {
		apiErr(c, err)
		return
	}
	c.JSON(http.StatusOK, activity)
}

func (h *ProfileHandler) DeleteActivity(c *gin.Context) {
	if err := h.svc.RemoveActivity(c.Request.Context(), c.Param("id"), sessionUserID(c)); err != nil {
		apiErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
