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

	// 订阅消息（先订阅，再发布）
	fmt.Println("\n=== 订阅消息 ===")
	messageHandler := func(topic string, message types.MessageEnvelope) error {
		fmt.Printf("📨 收到消息:\n")
		fmt.Printf("   实际主题: %s\n", topic)
		fmt.Printf("   CorrelationID: %s\n", message.CorrelationID)

		// 安全地处理 Payload
		var payloadStr string
		if payload, ok := message.Payload.([]byte); ok {
			payloadStr = string(payload)
		} else {
			payloadStr = fmt.Sprintf("%v", message.Payload)
		}
		fmt.Printf("   内容: %s\n", payloadStr)
		fmt.Println("   ---")
		return nil
	}

	// 只订阅 edgex/events/# 主题
	subscribeTopics := []string{"edgex/events/#"}
	if err := client.Subscribe(subscribeTopics, messageHandler); err != nil {
		log.Printf("订阅失败: %v", err)
	} else {
		fmt.Printf("✅ 成功订阅主题: %v\n", subscribeTopics)
	}

	// 等待一下确保订阅生效
	time.Sleep(1 * time.Second)

	// 发布多个消息到不同的子主题
	fmt.Println("\n=== 发布消息到不同子主题 ===")

	// 发布到 edgex/events/device/sensor01
	data1 := map[string]interface{}{
		"deviceName": "sensor01",
		"reading":    25.6,
		"timestamp":  time.Now().Unix(),
		"unit":       "°C",
		"type":       "temperature",
	}
	topic1 := "edgex/events/device/sensor01"
	if err := client.Publish(topic1, data1); err != nil {
		log.Printf("发布失败: %v", err)
	} else {
		fmt.Printf("✅ 成功发布消息到主题: %s\n", topic1)
	}

	// 发布到 edgex/events/device/sensor02/temperature
	data2 := map[string]interface{}{
		"deviceName": "sensor02",
		"reading":    30.2,
		"timestamp":  time.Now().Unix(),
		"unit":       "°C",
		"type":       "temperature",
	}
	topic2 := "edgex/events/device/sensor02/temperature"
	if err := client.Publish(topic2, data2); err != nil {
		log.Printf("发布失败: %v", err)
	} else {
		fmt.Printf("✅ 成功发布消息到主题: %s\n", topic2)
	}

	// 发布到 edgex/events/gateway/status
	data3 := map[string]interface{}{
		"gatewayId": "gateway01",
		"status":    "online",
		"timestamp": time.Now().Unix(),
		"uptime":    3600,
	}
	topic3 := "edgex/events/gateway/status"
	if err := client.Publish(topic3, data3); err != nil {
		log.Printf("发布失败: %v", err)
	} else {
		fmt.Printf("✅ 成功发布消息到主题: %s\n", topic3)
	}

	// 发布到 edgex/events/alert/critical/fire
	data4 := map[string]interface{}{
		"alertType": "fire",
		"severity":  "critical",
		"location":  "building-A",
		"timestamp": time.Now().Unix(),
		"message":   "Fire detected in building A",
	}
	topic4 := "edgex/events/alert/critical/fire"
	if err := client.Publish(topic4, data4); err != nil {
		log.Printf("发布失败: %v", err)
	} else {
		fmt.Printf("✅ 成功发布消息到主题: %s\n", topic4)
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
	fmt.Println("\n等待10秒钟以接收消息...")
	time.Sleep(10 * time.Second)

	fmt.Println("\n示例完成!")
}
