package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// ErrorHandler 统一错误处理中间件
func ErrorHandler(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := next(c)
		if err != nil {
			// 如果错误已经被处理过，直接返回
			if c.Response().Committed {
				return err
			}

			// 处理 HTTP 错误
			if he, ok := err.(*echo.HTTPError); ok {
				return c.JSON(he.Code, map[string]interface{}{
					"error":   http.StatusText(he.Code),
					"message": he.Message,
				})
			}

			// 处理其他错误
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error":   "Internal Server Error",
				"message": err.Error(),
			})
		}
		return nil
	}
}
