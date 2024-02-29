package middleware

import (
	"errors"
	"gin_scaffold/public"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

func SessionAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		if admininfo, ok := session.Get(public.AdminSessionInfoKey).(string); !ok || admininfo == "" {
			ResponseError(c, InternalErrorCode, errors.New("admin not login"))
			c.Abort()
			return
		}
		c.Next()
	}
}
