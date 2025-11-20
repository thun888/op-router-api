package handler

import (
	"log"
	"net/http"
	"router-api/service"

	"github.com/labstack/echo/v4"
)

// GetNetworkInterfaces 获取所有网络接口
func GetNetworkInterfaces(c echo.Context) error {
	// 优先尝试使用 ubus
	ubusSvc := service.NewUbusService()
	interfaces, err := ubusSvc.GetNetworkInterfaces()

	if err != nil {
		log.Printf("ubus 方式失败: %v, 尝试 LuCI RPC 方式", err)

		// 回退到 LuCI RPC
		svc := service.NewOpenWrtService()
		interfaces, err = svc.GetNetworkInterfaces()

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error":   "failed to get network interfaces",
				"message": err.Error(),
			})
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    interfaces,
	})
}

// GetNetworkInterface 获取指定网络接口
func GetNetworkInterface(c echo.Context) error {
	name := c.Param("name")
	if name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "interface name is required",
		})
	}

	// 优先尝试使用 ubus
	ubusSvc := service.NewUbusService()
	iface, err := ubusSvc.GetNetworkInterface(name)

	if err != nil {
		log.Printf("ubus 方式失败: %v, 尝试 LuCI RPC 方式", err)

		// 回退到 LuCI RPC
		svc := service.NewOpenWrtService()
		iface, err = svc.GetNetworkInterface(name)

		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error":   "interface not found",
				"message": err.Error(),
			})
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    iface,
	})
}

// GetNetworkStatus 获取网络状态
func GetNetworkStatus(c echo.Context) error {
	// 使用 ubus 获取网络设备信息
	ubusSvc := service.NewUbusService()
	status, err := ubusSvc.GetNetworkDevices()

	if err != nil {
		log.Printf("ubus 方式失败: %v, 尝试 LuCI RPC 方式", err)

		// 回退到 LuCI RPC
		svc := service.NewOpenWrtService()
		status, err = svc.GetNetworkStatus()

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error":   "failed to get network status",
				"message": err.Error(),
			})
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    status,
	})
}
