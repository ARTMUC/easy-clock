package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const SessionUserKey = "user_id"

// RequireAuth redirects unauthenticated requests to /login.
// Sets "user_id" in the Gin context for downstream handlers.
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
