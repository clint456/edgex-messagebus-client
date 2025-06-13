# EdgeX MessageBus Client 示例

本目录包含了 EdgeX MessageBus Client 的各种使用示例。

## 📁 示例目录

### 1. 基础示例 (`main.go`)
基础的 MessageBus 客户端使用示例，展示：
- 连接到 MessageBus
- 发布消息到多个主题
- 订阅通配符主题并接收具体主题路径
- 健康检查和客户端信息获取

```bash
cd example
go run main.go
```

### 2. 通配符订阅示例 (`wildcard/main.go`)
专门展示通配符订阅功能的示例，包括：
- 订阅多个通配符主题 (`edgex/events/#`, `edgex/commands/#`, `edgex/alerts/#`)
- 根据具体主题路径进行消息分类处理
- 发布测试消息到不同的子主题
- 实时接收 EdgeX 系统中的设备事件

```bash
cd example/wildcard
go run main.go
```

### 3. 高级示例 (`advanced/main.go`)
展示高级功能的综合示例，包括：
- 请求-响应模式
- 二进制数据处理
- 错误处理和监控
- 多种消息类型处理

```bash
cd example/advanced
go run main.go
```

## 🚀 快速开始

### 前置条件
1. 确保 MQTT Broker 正在运行（默认 localhost:1883）
2. 可选：运行 EdgeX 虚拟设备服务以查看真实的设备事件

### 启动 MQTT Broker
使用 Docker 快速启动：
```bash
docker run -it -p 1883:1883 eclipse-mosquitto:2.0
```

### 运行示例
```bash
# 克隆项目
git clone https://github.com/clint456/edgex-messagebus-client.git
cd edgex-messagebus-client

# 运行基础示例
cd example
go run main.go

# 运行通配符示例
cd wildcard
go run main.go
```

## 🎯 通配符订阅重点功能

### 关键特性
- **自动主题解析**：订阅 `edgex/events/#` 时，处理函数接收到的是具体主题如 `edgex/events/device/sensor01`
- **智能消息分类**：根据具体主题路径自动分类处理不同类型的消息
- **实时事件接收**：能够接收来自 EdgeX 系统的真实设备事件

### 示例输出
```
📨 收到消息:
   具体主题: edgex/events/device/temperature-sensor-01
   CorrelationID: MessageBus-a506e631-faca-4ad8-8c2a-9811080b827b
   事件类型: 🔧 设备事件
   消息内容: eyJkZXZpY2VOYW1lIjoidGVtcGVyYXR1cmUtc2Vuc29yLTAxIiwicmVhZGluZyI6MjMuNSwidGltZXN0YW1wIjoxNzQ5ODA4MTgzLCJ1bml0IjoiwrBDIn0=
```

## 📚 更多信息

- [主项目 README](../README.md)
- [使用指南](../USAGE.md)
- [API 文档](https://pkg.go.dev/github.com/clint456/edgex-messagebus-client)

## 🔧 故障排除

### 连接问题
- 确保 MQTT Broker 正在运行
- 检查主机地址和端口配置
- 验证网络连接

### 无法接收消息
- 确认订阅主题正确
- 检查 QoS 设置
- 验证消息处理函数

### 性能优化
- 调整消息缓冲区大小
- 使用适当的 QoS 级别
- 实现错误处理机制
