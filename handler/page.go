package handler

import (
	_ "embed"
	"net/http"

	"github.com/labstack/echo/v4"
)

//go:embed static/index.html
var indexHTML []byte

// NetworkPage 返回网卡信息展示页面
func NetworkPage(c echo.Context) error {
	return c.HTMLBlob(http.StatusOK, indexHTML)
}
