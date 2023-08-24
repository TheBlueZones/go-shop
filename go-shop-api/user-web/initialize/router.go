package initialize

import (
	"github.com/gin-gonic/gin"
	"mxshop-api/user-web/middlewares"
	"mxshop-api/user-web/router"
	"net/http"
)

func CustomLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if path != "/health" {
			gin.Logger()(c)
		}
		c.Next()
	}
}

func Routers() *gin.Engine {
	Router := gin.New()        // 使用 gin.New() 创建没有默认中间件的 Gin 引擎
	Router.Use(CustomLogger()) // 使用自定义日志中间件
	Router.Use(gin.Recovery()) // 添加恢复中间件

	Router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"success": true,
		})
	})

	//配置跨域
	Router.Use(middlewares.Cors())

	ApiGroup := Router.Group("/u/v1")
	router.InitUserRouter(ApiGroup)
	router.InitBaseRouter(ApiGroup)

	return Router
}
