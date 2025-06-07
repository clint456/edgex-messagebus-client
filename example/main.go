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
	lc := logger.NewClient("MessageBusExample", "DEBUG")

	// 配置MessageBus客户端
	config := messagebus.Config{
		Host:     "localhost",
		Port:     1883,
		Protocol: "tcp",
		Type:     "mqtt",
		ClientID: "example-client",
		QoS:      1,
	}

	// 创建客户端
	client, err := messagebus.NewClient(config, lc)
	if err != nil {
		log.Fatalf("创建MessageBus客户端失败: %v", err)
	}

	// 连接到MessageBus
	fmt.Println("正在连接到MessageBus...")
	if err := client.Connect(); err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer client.Disconnect()

	fmt.Println("✅ 连接成功!")

	// 发布消息
	fmt.Println("\n=== 发布消息 ===")
	data := map[string]interface{}{
		"deviceName": "sensor01",
		"reading":    25.6,
		"timestamp":  time.Now().Unix(),
		"unit":       "°C",
	}

	topic := "edgex/events/device/sensor01"
	if err := client.Publish(topic, data); err != nil {
		log.Printf("发布失败: %v", err)
	} else {
		fmt.Printf("✅ 成功发布消息到主题: %s\n", topic)
	}

	// 订阅消息
	fmt.Println("\n=== 订阅消息 ===")
	messageHandler := func(topic string, message types.MessageEnvelope) error {
		fmt.Printf("📨 收到消息:\n")
		fmt.Printf("   主题: %s\n", topic)
		fmt.Printf("   CorrelationID: %s\n", message.CorrelationID)

		// 安全地处理 Payload
		var payloadStr string
		if payload, ok := message.Payload.([]byte); ok {
			payloadStr = string(payload)
		} else {
			payloadStr = fmt.Sprintf("%v", message.Payload)
		}
		fmt.Printf("   内容: %s\n", payloadStr)
		return nil
	}

	// 订阅主题
	subscribeTopics := []string{"edgex/events/#", "edgex/test/#"}
	if err := client.Subscribe(subscribeTopics, messageHandler); err != nil {
		log.Printf("订阅失败: %v", err)
	} else {
		fmt.Printf("✅ 成功订阅主题: %v\n", subscribeTopics)
	}

	// 获取客户端信息
	fmt.Println("\n=== 客户端信息 ===")
	info := client.GetClientInfo()
	fmt.Printf("客户端信息: %+v\n", info)

	// 健康检查
	fmt.Println("\n=== 健康检查 ===")
	if err := client.HealthCheck(); err != nil {
		fmt.Printf("❌ 健康检查失败: %v\n", err)
	} else {
		fmt.Println("✅ 健康检查通过")
	}

	// 等待一段时间以接收消息
	fmt.Println("\n等待5秒钟以接收消息...")
	time.Sleep(5 * time.Second)

	fmt.Println("\n示例完成!")
}
