package main

import (
	"net/http"
	"router-api/handler"
	"router-api/middleware"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

func main() {
	// 创建 Echo 实例
	e := echo.New()

	// 中间件
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORS())
	e.Use(middleware.ErrorHandler)

	// 路由
	api := e.Group("/api")
	{
		// 获取所有网络接口信息
		api.GET("/network/interfaces", handler.GetNetworkInterfaces)

		// 获取指定接口信息
		api.GET("/network/interfaces/:name", handler.GetNetworkInterface)

		// 获取接口状态
		api.GET("/network/status", handler.GetNetworkStatus)

		// 健康检查
		api.GET("/health", func(c echo.Context) error {
			return c.JSON(http.StatusOK, map[string]string{
				"status": "ok",
			})
		})
	}

	// 启动服务器
	e.Logger.Fatal(e.Start(":8080"))
}
