package service

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"router-api/config"
	"strings"
	"time"
)

// UbusService 使用 ubus 的服务
type UbusService struct {
	config  *config.OpenWrtConfig
	client  *http.Client
	sid     string
	ubusSid string
}

// NewUbusService 创建 ubus 服务实例
func NewUbusService() *UbusService {
	cfg := config.GetOpenWrtConfig()

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	return &UbusService{
		config: cfg,
		client: client,
	}
}

// Login 通过 ubus 登录
func (s *UbusService) Login() error {
	url := s.config.GetBaseURL() + "/ubus"

	log.Printf("尝试通过 ubus 登录: %s", url)

	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "call",
		"params": []interface{}{
			"00000000000000000000000000000000",
			"session",
			"login",
			map[string]string{
				"username": s.config.Username,
				"password": s.config.Password,
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("构建登录请求失败: %v", err)
	}

	log.Printf("ubus 登录请求: %s", string(jsonData))

	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("连接失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	log.Printf("ubus 登录响应: %s", string(body))

	if strings.Contains(string(body), "<html") || strings.Contains(string(body), "<!DOCTYPE") {
		return fmt.Errorf("收到 HTML 响应，可能 ubus 未启用或 URL 错误")
	}

	var result struct {
		Jsonrpc string        `json:"jsonrpc"`
		ID      int           `json:"id"`
		Result  []interface{} `json:"result"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析响应失败: %v, 内容: %s", err, string(body))
	}

	if len(result.Result) >= 2 {
		if resultData, ok := result.Result[1].(map[string]interface{}); ok {
			if ubusSid, ok := resultData["ubus_rpc_session"].(string); ok && ubusSid != "" {
				s.ubusSid = ubusSid
				log.Printf("登录成功，ubus_rpc_session: %s", s.ubusSid)
				return nil
			}
		}
	}

	return fmt.Errorf("登录失败，未获取到会话 ID")
}

// CallUbus 调用 ubus 方法
func (s *UbusService) CallUbus(object, method string, params map[string]interface{}) (map[string]interface{}, error) {
	if s.ubusSid == "" {
		if err := s.Login(); err != nil {
			return nil, err
		}
	}

	url := s.config.GetBaseURL() + "/ubus"

	if params == nil {
		params = make(map[string]interface{})
	}

	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "call",
		"params": []interface{}{
			s.ubusSid,
			object,
			method,
			params,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("构建请求失败: %v", err)
	}

	log.Printf("ubus 调用: %s", string(jsonData))

	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	log.Printf("ubus 响应: %s", string(body))

	var result struct {
		Jsonrpc string        `json:"jsonrpc"`
		ID      int           `json:"id"`
		Result  []interface{} `json:"result"`
		Error   interface{}   `json:"error"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("ubus 返回错误: %v", result.Error)
	}

	if len(result.Result) >= 2 {
		if resultData, ok := result.Result[1].(map[string]interface{}); ok {
			return resultData, nil
		}
	}

	return make(map[string]interface{}), nil
}

// GetNetworkInterfaces 获取网络接口
func (s *UbusService) GetNetworkInterfaces() ([]map[string]interface{}, error) {
	result, err := s.CallUbus("network.interface", "dump", nil)
	if err != nil {
		return nil, err
	}

	interfaces := make([]map[string]interface{}, 0)

	if ifaceList, ok := result["interface"].([]interface{}); ok {
		for _, iface := range ifaceList {
			if ifaceData, ok := iface.(map[string]interface{}); ok {
				interfaces = append(interfaces, ifaceData)
			}
		}
	}

	return interfaces, nil
}

// GetNetworkInterface 获取指定网络接口
func (s *UbusService) GetNetworkInterface(name string) (map[string]interface{}, error) {
	result, err := s.CallUbus("network.interface."+name, "status", nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetNetworkDevices 获取网络设备信息
func (s *UbusService) GetNetworkDevices() (map[string]interface{}, error) {
	result, err := s.CallUbus("network.device", "status", nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}
