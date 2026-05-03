package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"easy-clock/internal/app"
)

type ChildHandler struct {
	svc *app.ChildService
}

func NewChildHandler(svc *app.ChildService) *ChildHandler {
	return &ChildHandler{svc: svc}
}

func (h *ChildHandler) List(c *gin.Context) {
	children, err := h.svc.ListChildren(c.Request.Context(), sessionUserID(c))
	if err != nil {
		apiErr(c, err)
		return
	}
	c.JSON(http.StatusOK, children)
}

func (h *ChildHandler) Create(c *gin.Context) {
	var body struct {
		Name     string `json:"name"     binding:"required"`
		Timezone string `json:"timezone" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	child, err := h.svc.AddChild(c.Request.Context(), sessionUserID(c), body.Name, body.Timezone)
	if err != nil {
		apiErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, child)
}

func (h *ChildHandler) Get(c *gin.Context) {
	child, err := h.svc.GetChild(c.Request.Context(), c.Param("id"), sessionUserID(c))
	if err != nil {
		apiErr(c, err)
		return
	}
	c.JSON(http.StatusOK, child)
}

func (h *ChildHandler) Update(c *gin.Context) {
	var body struct {
		Name       string `json:"name"        binding:"required"`
		Timezone   string `json:"timezone"    binding:"required"`
		AvatarPath string `json:"avatar_path"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	child, err := h.svc.UpdateChild(c.Request.Context(), c.Param("id"), sessionUserID(c), body.Name, body.Timezone, body.AvatarPath)
	if err != nil {
		apiErr(c, err)
		return
	}
	c.JSON(http.StatusOK, child)
}

func (h *ChildHandler) Delete(c *gin.Context) {
	if err := h.svc.RemoveChild(c.Request.Context(), c.Param("id"), sessionUserID(c)); err != nil {
		apiErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ChildHandler) SetDefaultProfile(c *gin.Context) {
	var body struct {
		ProfileID string `json:"profile_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.SetDefaultProfile(c.Request.Context(), c.Param("id"), sessionUserID(c), body.ProfileID); err != nil {
		apiErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
