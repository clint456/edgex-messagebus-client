# EdgeX MessageBus Client

[![Go Report Card](https://goreportcard.com/badge/github.com/clint456/edgex-messagebus-client)](https://goreportcard.com/report/github.com/clint456/edgex-messagebus-client) [![GitHub License](https://img.shields.io/github/license/clint456/edgex-messagebus-client)](https://choosealicense.com/licenses/apache-2.0/)

一个高级的 EdgeX Foundry MessageBus 客户端库，提供简单易用的 API 来进行 MQTT 消息的发布和订阅操作。

## 🚀 特性

- ✅ **完整的 EdgeX MessageBus 支持** - 基于官方 `go-mod-messaging` 库
- ✅ **连接管理** - 自动连接、断开连接和重连机制
- ✅ **消息发布** - 支持多种数据类型的消息发布
- ✅ **消息订阅** - 支持主题订阅和自定义消息处理
- ✅ **请求-响应模式** - 支持同步请求-响应操作
- ✅ **二进制数据支持** - 支持发布和订阅二进制数据
- ✅ **线程安全** - 所有操作都是线程安全的
- ✅ **错误处理** - 完善的错误处理和日志记录
- ✅ **健康检查** - 提供客户端健康状态检查

## 📦 安装

```bash
go get github.com/clint456/edgex-messagebus-client
```

## 🛠️ 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "log"
    "time"

    messagebus "github.com/clint456/edgex-messagebus-client"
    "github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
    "github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
)

func main() {
    // 创建日志客户端
    lc := logger.NewClient("MyApp", "DEBUG")

    // 配置MessageBus客户端
    config := messagebus.Config{
        Host:     "localhost",
        Port:     1883,
        Protocol: "tcp",
        Type:     "mqtt",
        ClientID: "my-client",
        QoS:      1,
    }

    // 创建并连接客户端
    client, err := messagebus.NewClient(config, lc)
    if err != nil {
        log.Fatal(err)
    }

    if err := client.Connect(); err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect()

    // 发布消息
    data := map[string]interface{}{
        "deviceName": "sensor01",
        "reading":    25.6,
        "timestamp":  time.Now(),
    }
    client.Publish("edgex/events/device/sensor01", data)

    // 订阅消息
    handler := func(topic string, message types.MessageEnvelope) error {
        fmt.Printf("收到消息: %s\n", string(message.Payload.([]byte)))
        return nil
    }
    client.SubscribeSingle("edgex/events/#", handler)

    // 等待消息
    time.Sleep(10 * time.Second)
}
```

## 📚 API 参考

### 配置结构

```go
type Config struct {
    Host     string  // MQTT Broker 主机地址
    Port     int     // MQTT Broker 端口
    Protocol string  // 协议 (tcp, ssl, ws, wss)
    Type     string  // 消息总线类型 (mqtt, nats)
    ClientID string  // 客户端 ID
    Username string  // 用户名 (可选)
    Password string  // 密码 (可选)
    QoS      int     // QoS 级别 (0, 1, 2)
}
```

### 主要方法

| 方法 | 描述 |
|------|------|
| `NewClient(config, logger)` | 创建新的客户端 |
| `Connect()` | 连接到 MessageBus |
| `Disconnect()` | 断开连接 |
| `IsConnected()` | 检查连接状态 |
| `Publish(topic, data)` | 发布消息 |
| `Subscribe(topics, handler)` | 订阅主题 |
| `Unsubscribe(topics...)` | 取消订阅 |
| `HealthCheck()` | 健康检查 |

### 高级方法

| 方法 | 描述 |
|------|------|
| `PublishWithCorrelationID()` | 使用指定 CorrelationID 发布 |
| `PublishBinaryData()` | 发布二进制数据 |
| `Request()` | 请求-响应操作 |
| `CreateMessageEnvelope()` | 创建消息信封 |
| `GetClientInfo()` | 获取客户端信息 |

## 🔧 高级用法

### 请求-响应模式

```go
// 创建请求消息
envelope, _ := client.CreateMessageEnvelope(requestData, "")

// 发送请求并等待响应
response, err := client.Request(
    envelope,
    "edgex/command/request",
    "edgex/command/response",
    5*time.Second,
)
```

### 二进制数据发布

```go
binaryData := []byte{0x01, 0x02, 0x03, 0x04}
client.PublishBinaryData("edgex/binary/data", binaryData)
```

### 错误监听

```go
go func() {
    for err := range client.GetErrorChannel() {
        log.Printf("MessageBus 错误: %v", err)
    }
}()
```

## 📄 许可证

本项目采用 [Apache-2.0](LICENSE) 许可证。

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📞 联系方式

- 项目地址: [https://github.com/clint456/edgex-messagebus-client](https://github.com/clint456/edgex-messagebus-client)
- 问题反馈: [Issues](https://github.com/clint456/edgex-messagebus-client/issues)
