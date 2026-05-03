package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"easy-clock/internal/token"
)

const SessionUserKey = "user_id"

// RequireAuth redirects unauthenticated browser requests to /login (session-based).
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		uid := session.Get(SessionUserKey)
		if uid == nil {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}
		id, ok := uid.(string)
		if !ok || id == "" {
			session.Delete(SessionUserKey)
			_ = session.Save()
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}
		c.Set(SessionUserKey, id)
		c.Next()
	}
}

// RequireJWT validates a Bearer JWT and sets "user_id" in the Gin context.
// Returns 401 JSON on failure — intended for API routes.
func RequireJWT(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid Authorization header"})
			c.Abort()
			return
		}
		userID, err := token.Validate(strings.TrimPrefix(header, "Bearer "), secret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}
		c.Set(SessionUserKey, userID)
		c.Next()
	}
}
