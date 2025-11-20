# Router API

使用 Go Echo 框架开发的 OpenWrt 路由器网络接口查询 API。

## 功能

- 获取所有网络接口信息
- 获取指定网络接口信息
- 获取网络状态信息

## 环境要求

- Go 1.21+
- OpenWrt 路由器（支持 LuCI RPC API）

## 安装

1. 克隆项目
```bash
git clone <repository-url>
cd router-api
```

2. 安装依赖
```bash
go mod download
```

3. 配置环境变量

复制 `.env.example` 为 `.env` 并修改配置：

```bash
cp .env.example .env
```

编辑 `.env` 文件，设置你的 OpenWrt 路由器信息：

```env
OPENWRT_PROTOCOL=http
OPENWRT_HOST=192.168.1.1
OPENWRT_PORT=80
OPENWRT_USERNAME=root
OPENWRT_PASSWORD=your_password
```

## 运行

```bash
go run main.go
```

服务将在 `http://localhost:8080` 启动。

## API 接口

### 1. 健康检查

```
GET /api/health
```

响应：
```json
{
  "status": "ok"
}
```

### 2. 获取所有网络接口

```
GET /api/network/interfaces
```

响应：
```json
{
  "success": true,
  "data": [
    {
      "name": "lan",
      "proto": "static",
      "ipaddr": "192.168.1.1",
      "netmask": "255.255.255.0",
      "type": "bridge",
      ...
    },
    {
      "name": "wan",
      "proto": "dhcp",
      ...
    }
  ]
}
```

### 3. 获取指定网络接口

```
GET /api/network/interfaces/:name
```

示例：
```
GET /api/network/interfaces/lan
```

响应：
```json
{
  "success": true,
  "data": {
    "name": "lan",
    "proto": "static",
    "ipaddr": "192.168.1.1",
    "netmask": "255.255.255.0",
    "type": "bridge",
    ...
  }
}
```

### 4. 获取网络状态

```
GET /api/network/status
```

响应：
```json
{
  "success": true,
  "data": {
    "result": {
      // 网络设备信息
    }
  }
}
```

## 项目结构

```
router-api/
├── main.go              # 主程序入口
├── config/              # 配置模块
│   └── config.go
├── handler/             # 请求处理器
│   └── network.go
├── service/             # 业务逻辑
│   └── openwrt.go
├── middleware/          # 中间件
│   └── error.go
├── go.mod               # Go 模块定义
├── .env.example         # 环境变量示例
└── README.md            # 项目说明
```

## 开发

### 构建

```bash
go build -o router-api.exe
```

### 测试

```bash
# 测试健康检查
curl http://localhost:8080/api/health

# 测试获取所有接口
curl http://localhost:8080/api/network/interfaces

# 测试获取指定接口
curl http://localhost:8080/api/network/interfaces/lan

# 测试获取网络状态
curl http://localhost:8080/api/network/status
```

## 注意事项

1. 确保你的 OpenWrt 路由器已启用 LuCI RPC API
2. 默认跳过了 HTTPS 证书验证，生产环境建议配置正确的证书
3. 根据你的 OpenWrt 版本，API 接口可能略有不同
4. 建议使用环境变量管理敏感信息

## 许可证

MIT
