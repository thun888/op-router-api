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

type OpenWrtService struct {
	config *config.OpenWrtConfig
	client *http.Client
	token  string
}

// NewOpenWrtService 创建 OpenWrt 服务实例
func NewOpenWrtService() *OpenWrtService {
	cfg := config.GetOpenWrtConfig()

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // 跳过证书验证（生产环境需要配置正确的证书）
			},
		},
	}

	return &OpenWrtService{
		config: cfg,
		client: client,
	}
}

// Login 登录 OpenWrt 获取 token
func (s *OpenWrtService) Login() error {
	url := s.config.GetBaseURL() + "/cgi-bin/luci/rpc/auth"

	log.Printf("尝试登录 OpenWrt: %s", url)

	payload := map[string]interface{}{
		"id":     1,
		"method": "login",
		"params": []string{s.config.Username, s.config.Password},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("构建登录请求失败: %v", err)
	}

	log.Printf("登录请求: %s", string(jsonData))

	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("连接 OpenWrt 失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	log.Printf("登录响应状态: %d", resp.StatusCode)
	log.Printf("登录响应内容: %s", string(body))

	// 检查响应是否是 HTML（可能是错误页面）
	if strings.Contains(string(body), "<html") || strings.Contains(string(body), "<!DOCTYPE") {
		return fmt.Errorf("收到 HTML 响应而非 JSON，可能是 URL 错误或 LuCI RPC 未启用")
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析登录响应失败: %v, 响应内容: %s", err, string(body))
	}

	if token, ok := result["result"].(string); ok && token != "" {
		s.token = token
		log.Printf("登录成功，获取到 token")
		return nil
	}

	return fmt.Errorf("登录失败，响应: %v", result)
}

// CallUCI 调用 UCI 接口
func (s *OpenWrtService) CallUCI(method string, params []string) (map[string]interface{}, error) {
	if s.token == "" {
		if err := s.Login(); err != nil {
			return nil, fmt.Errorf("登录失败: %v", err)
		}
	}

	url := s.config.GetBaseURL() + "/cgi-bin/luci/rpc/uci?auth=" + s.token

	payload := map[string]interface{}{
		"id":     1,
		"method": method,
		"params": params,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("构建请求失败: %v", err)
	}

	log.Printf("UCI 请求: %s", string(jsonData))

	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	log.Printf("UCI 响应状态: %d", resp.StatusCode)
	log.Printf("UCI 响应内容: %s", string(body))

	// 检查响应是否是 HTML
	if strings.Contains(string(body), "<html") || strings.Contains(string(body), "<!DOCTYPE") {
		return nil, fmt.Errorf("收到 HTML 响应，token 可能已过期，响应: %s", string(body[:200]))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v, 响应内容: %s", err, string(body))
	}

	// 检查是否有错误
	if errMsg, ok := result["error"].(map[string]interface{}); ok {
		return nil, fmt.Errorf("OpenWrt 返回错误: %v", errMsg)
	}

	return result, nil
}

// CallSystem 调用系统接口
func (s *OpenWrtService) CallSystem(method string, params []string) (map[string]interface{}, error) {
	if s.token == "" {
		if err := s.Login(); err != nil {
			return nil, err
		}
	}

	url := s.config.GetBaseURL() + "/cgi-bin/luci/rpc/sys?auth=" + s.token

	payload := map[string]interface{}{
		"id":     1,
		"method": method,
		"params": params,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetNetworkInterfaces 获取网络接口信息
func (s *OpenWrtService) GetNetworkInterfaces() ([]map[string]interface{}, error) {
	// 方式1: 通过 UCI 获取配置的接口
	result, err := s.CallUCI("get_all", []string{"network"})
	if err != nil {
		return nil, err
	}

	interfaces := make([]map[string]interface{}, 0)

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		for name, data := range resultData {
			if ifaceData, ok := data.(map[string]interface{}); ok {
				if ifaceType, exists := ifaceData[".type"]; exists && ifaceType == "interface" {
					ifaceData["name"] = name
					interfaces = append(interfaces, ifaceData)
				}
			}
		}
	}

	return interfaces, nil
}

// GetNetworkInterface 获取指定接口信息
func (s *OpenWrtService) GetNetworkInterface(name string) (map[string]interface{}, error) {
	result, err := s.CallUCI("get_all", []string{"network", name})
	if err != nil {
		return nil, err
	}

	if resultData, ok := result["result"].(map[string]interface{}); ok {
		resultData["name"] = name
		return resultData, nil
	}

	return nil, fmt.Errorf("接口不存在")
}

// GetNetworkStatus 获取网络状态信息（通过执行系统命令）
func (s *OpenWrtService) GetNetworkStatus() (map[string]interface{}, error) {
	result, err := s.CallSystem("net.deviceinfo", []string{})
	if err != nil {
		return nil, err
	}

	return result, nil
}
