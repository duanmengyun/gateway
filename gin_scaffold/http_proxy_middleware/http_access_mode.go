package http_proxy_middleware

import (
	"gin_scaffold/dao"
	"gin_scaffold/middleware"

	"github.com/gin-gonic/gin"
)

// 使用请求信息和服务列表匹配
func HTTPAccessModeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		service, err := dao.ServiceManagerHandler.HTTPAccessMode(c)
		if err != nil {
			middleware.ResponseError(c, 1001, err)
			c.Abort()
			return
		}
		//fmt.Println("matched service",public.Obj2Json(service))
		c.Set("service", service)
		c.Next()
	}
}
