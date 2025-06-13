package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	messagebus "github.com/clint456/edgex-messagebus-client"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
)

func main() {
	fmt.Println("🚀 EdgeX MessageBus 通配符订阅示例")
	fmt.Println("===================================")

	// 创建日志客户端
	lc := logger.NewClient("WildcardExample", "INFO")

	// 配置MessageBus客户端
	config := messagebus.Config{
		Host:     "localhost",
		Port:     1883,
		Protocol: "tcp",
		Type:     "mqtt",
		ClientID: "wildcard-example-client",
		QoS:      1,
	}

	// 创建客户端
	client, err := messagebus.NewClient(config, lc)
	if err != nil {
		log.Fatalf("创建MessageBus客户端失败: %v", err)
	}

	// 连接到MessageBus
	fmt.Println("\n正在连接到MessageBus...")
	if err := client.Connect(); err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer client.Disconnect()

	fmt.Println("✅ 连接成功!")

	// 定义消息处理函数
	messageHandler := func(topic string, message types.MessageEnvelope) error {
		fmt.Printf("\n📨 收到消息:\n")
		fmt.Printf("   具体主题: %s\n", topic)
		fmt.Printf("   CorrelationID: %s\n", message.CorrelationID)

		// 根据主题路径分类处理
		var messageType string
		switch {
		case strings.Contains(topic, "/device/"):
			messageType = "🔧 设备事件"
		case strings.Contains(topic, "/alert/"):
			messageType = "🚨 告警事件"
		case strings.Contains(topic, "/gateway/"):
			messageType = "🌐 网关事件"
		case strings.Contains(topic, "/system/"):
			messageType = "⚙️ 系统事件"
		default:
			messageType = "📄 通用事件"
		}
		fmt.Printf("   事件类型: %s\n", messageType)

		// 安全地处理 Payload
		var payloadStr string
		if payload, ok := message.Payload.([]byte); ok {
			payloadStr = string(payload)
		} else {
			payloadStr = fmt.Sprintf("%v", message.Payload)
		}
		fmt.Printf("   消息内容: %s\n", payloadStr)
		fmt.Println("   " + strings.Repeat("-", 50))
		return nil
	}

	// 订阅通配符主题
	fmt.Println("\n=== 订阅通配符主题 ===")
	wildcardTopics := []string{
		"edgex/events/#",
		"edgex/commands/#",
		"edgex/alerts/#",
	}

	if err := client.Subscribe(wildcardTopics, messageHandler); err != nil {
		log.Printf("订阅失败: %v", err)
		return
	}

	fmt.Printf("✅ 成功订阅通配符主题: %v\n", wildcardTopics)

	// 等待订阅生效
	time.Sleep(1 * time.Second)

	// 发布测试消息到不同的具体主题
	fmt.Println("\n=== 发布测试消息 ===")

	testMessages := []struct {
		topic string
		data  map[string]interface{}
		desc  string
	}{
		{
			topic: "edgex/events/device/temperature-sensor-01",
			data: map[string]interface{}{
				"deviceName": "temperature-sensor-01",
				"reading":    23.5,
				"unit":       "°C",
				"timestamp":  time.Now().Unix(),
			},
			desc: "温度传感器数据",
		},
		{
			topic: "edgex/events/device/humidity-sensor-02/data",
			data: map[string]interface{}{
				"deviceName": "humidity-sensor-02",
				"reading":    65.2,
				"unit":       "%RH",
				"timestamp":  time.Now().Unix(),
			},
			desc: "湿度传感器数据",
		},
		{
			topic: "edgex/alerts/critical/fire-alarm",
			data: map[string]interface{}{
				"alertType": "fire",
				"severity":  "critical",
				"location":  "building-A-floor-3",
				"timestamp": time.Now().Unix(),
				"message":   "Fire detected in server room",
			},
			desc: "火灾告警",
		},
		{
			topic: "edgex/events/gateway/status/health",
			data: map[string]interface{}{
				"gatewayId": "gateway-001",
				"status":    "healthy",
				"uptime":    7200,
				"timestamp": time.Now().Unix(),
			},
			desc: "网关健康状态",
		},
		{
			topic: "edgex/commands/device/actuator-01/set",
			data: map[string]interface{}{
				"command":   "setPosition",
				"value":     45,
				"unit":      "degrees",
				"timestamp": time.Now().Unix(),
			},
			desc: "执行器控制命令",
		},
	}

	for i, msg := range testMessages {
		fmt.Printf("\n%d. 发布 %s 到主题: %s\n", i+1, msg.desc, msg.topic)
		if err := client.Publish(msg.topic, msg.data); err != nil {
			log.Printf("   ❌ 发布失败: %v", err)
		} else {
			fmt.Printf("   ✅ 发布成功\n")
		}
		time.Sleep(500 * time.Millisecond) // 稍微延迟以便观察消息接收
	}

	// 等待接收消息
	fmt.Println("\n=== 等待接收消息 ===")
	fmt.Println("等待15秒钟以接收所有消息...")
	time.Sleep(15 * time.Second)

	// 显示订阅统计
	fmt.Println("\n=== 订阅统计 ===")
	subscribedTopics := client.GetSubscribedTopics()
	fmt.Printf("当前订阅的主题数量: %d\n", len(subscribedTopics))
	for _, topic := range subscribedTopics {
		fmt.Printf("  - %s\n", topic)
	}

	// 健康检查
	fmt.Println("\n=== 健康检查 ===")
	if err := client.HealthCheck(); err != nil {
		fmt.Printf("❌ 健康检查失败: %v\n", err)
	} else {
		fmt.Println("✅ 健康检查通过")
	}

	fmt.Println("\n🎉 通配符订阅示例完成!")
	fmt.Println("\n💡 提示:")
	fmt.Println("   - 通配符 '#' 匹配多级主题")
	fmt.Println("   - 通配符 '+' 匹配单级主题")
	fmt.Println("   - 消息处理函数的 topic 参数包含实际接收到的具体主题路径")
	fmt.Println("   - 可以根据具体主题路径进行不同的消息处理逻辑")
}
