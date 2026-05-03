package middleware

import (
	"github.com/gin-gonic/gin"

	"easy-clock/internal/i18n"
)

const LangKey = "lang"

func DetectLanguage() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := i18n.DetectLang(c.GetHeader("Accept-Language"))
		c.Set(LangKey, lang)
		c.Next()
	}
}
