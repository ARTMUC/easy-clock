package handler

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	uploadsDir  = "./static/uploads"
	maxFileSize = 5 << 20 // 5 MB
)

func UploadHandler(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxFileSize)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		slog.Warn("bad request", "method", c.Request.Method, "path", c.Request.URL.Path, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "file field is required (max 5 MB)"})
		return
	}
	defer file.Close()

	// Read first 512 bytes for MIME detection without consuming the stream
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	mt := mimetype.Detect(buf[:n])
	if !strings.HasPrefix(mt.String(), "image/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only image files are accepted"})
		return
	}

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = mt.Extension()
	}

	id, err := uuid.NewV7()
	if err != nil {
		slog.Error("api error", "method", c.Request.Method, "path", c.Request.URL.Path, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate filename"})
		return
	}
	filename := id.String() + ext
	dst := filepath.Join(uploadsDir, filename)

	if err := os.MkdirAll(uploadsDir, 0o755); err != nil {
		slog.Error("api error", "method", c.Request.Method, "path", c.Request.URL.Path, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create uploads directory"})
		return
	}

	out, err := os.Create(dst)
	if err != nil {
		slog.Error("api error", "method", c.Request.Method, "path", c.Request.URL.Path, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save file"})
		return
	}
	defer out.Close()

	// Write the already-read bytes first, then stream the rest
	if _, err := out.Write(buf[:n]); err != nil {
		slog.Error("api error", "method", c.Request.Method, "path", c.Request.URL.Path, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save file"})
		return
	}
	if _, err := io.Copy(out, file); err != nil {
		slog.Error("api error", "method", c.Request.Method, "path", c.Request.URL.Path, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save file"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"image_path": fmt.Sprintf("/static/uploads/%s", filename)})
}
