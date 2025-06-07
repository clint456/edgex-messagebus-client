package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	messagebus "github.com/clint456/edgex-messagebus-client"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
)

// SensorData represents a sensor reading
type SensorData struct {
	DeviceID   string    `json:"deviceId"`
	SensorType string    `json:"sensorType"`
	Value      float64   `json:"value"`
	Unit       string    `json:"unit"`
	Timestamp  time.Time `json:"timestamp"`
	Location   string    `json:"location"`
	Quality    int       `json:"quality"`
}

// CommandRequest represents a device command
type CommandRequest struct {
	DeviceID   string                 `json:"deviceId"`
	Command    string                 `json:"command"`
	Parameters map[string]interface{} `json:"parameters"`
	RequestID  string                 `json:"requestId"`
	Timestamp  time.Time              `json:"timestamp"`
}

// CommandResponse represents a command response
type CommandResponse struct {
	RequestID string                 `json:"requestId"`
	DeviceID  string                 `json:"deviceId"`
	Success   bool                   `json:"success"`
	Result    map[string]interface{} `json:"result"`
	Error     string                 `json:"error,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

func main() {
	// Create logger with structured logging
	lc := logger.NewClient("AdvancedMessageBusExample", "DEBUG")

	// Configuration with environment variable support
	config := messagebus.Config{
		Host:     getEnvOrDefault("MESSAGEBUS_HOST", "localhost"),
		Port:     getEnvIntOrDefault("MESSAGEBUS_PORT", 1883),
		Protocol: getEnvOrDefault("MESSAGEBUS_PROTOCOL", "tcp"),
		Type:     getEnvOrDefault("MESSAGEBUS_TYPE", "mqtt"),
		ClientID: getEnvOrDefault("MESSAGEBUS_CLIENT_ID", "advanced-example-client"),
		Username: os.Getenv("MESSAGEBUS_USERNAME"),
		Password: os.Getenv("MESSAGEBUS_PASSWORD"),
		QoS:      getEnvIntOrDefault("MESSAGEBUS_QOS", 1),
	}

	// Create client
	client, err := messagebus.NewClient(config, lc)
	if err != nil {
		log.Fatalf("Failed to create MessageBus client: %v", err)
	}

	// Connect with retry logic
	if err := connectWithRetry(client, 5, 2*time.Second); err != nil {
		log.Fatalf("Failed to connect after retries: %v", err)
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start error monitoring
	go monitorErrors(client, lc)

	// Start health monitoring
	go monitorHealth(ctx, client, lc)

	// Start sensor data publisher
	go publishSensorData(ctx, client, lc)

	// Start command handler
	go handleCommands(ctx, client, lc)

	// Start event subscriber
	go subscribeToEvents(ctx, client, lc)

	// Start request-response example
	go requestResponseExample(ctx, client, lc)

	lc.Info("Advanced MessageBus example started. Press Ctrl+C to stop.")

	// Wait for shutdown signal
	<-sigChan
	lc.Info("Shutdown signal received, cleaning up...")

	// Cancel context to stop all goroutines
	cancel()

	// Give goroutines time to cleanup
	time.Sleep(2 * time.Second)

	// Disconnect from MessageBus
	if err := client.Disconnect(); err != nil {
		lc.Errorf("Error during disconnect: %v", err)
	}

	lc.Info("Advanced MessageBus example stopped.")
}

// connectWithRetry attempts to connect with exponential backoff
func connectWithRetry(client *messagebus.Client, maxRetries int, initialDelay time.Duration) error {
	for i := 0; i < maxRetries; i++ {
		if err := client.Connect(); err != nil {
			if i == maxRetries-1 {
				return fmt.Errorf("failed to connect after %d attempts: %v", maxRetries, err)
			}
			delay := time.Duration(1<<uint(i)) * initialDelay
			log.Printf("Connection attempt %d failed, retrying in %v: %v", i+1, delay, err)
			time.Sleep(delay)
			continue
		}
		return nil
	}
	return nil
}

// monitorErrors monitors the error channel
func monitorErrors(client *messagebus.Client, lc logger.LoggingClient) {
	for err := range client.GetErrorChannel() {
		lc.Errorf("MessageBus error: %v", err)
		// In a real application, you might want to:
		// - Send alerts
		// - Increment error metrics
		// - Trigger reconnection logic
	}
}

// monitorHealth performs periodic health checks
func monitorHealth(ctx context.Context, client *messagebus.Client, lc logger.LoggingClient) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := client.HealthCheck(); err != nil {
				lc.Errorf("Health check failed: %v", err)
				// Implement reconnection logic here if needed
			} else {
				lc.Debug("Health check passed")
			}

			// Log client statistics
			info := client.GetClientInfo()
			lc.Debugf("Client info: %+v", info)
		}
	}
}

// publishSensorData simulates publishing sensor data
func publishSensorData(ctx context.Context, client *messagebus.Client, lc logger.LoggingClient) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	sensors := []string{"temperature", "humidity", "pressure"}
	locations := []string{"room1", "room2", "outdoor"}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for i, sensorType := range sensors {
				data := SensorData{
					DeviceID:   fmt.Sprintf("sensor-%s-%d", sensorType, i+1),
					SensorType: sensorType,
					Value:      20.0 + float64(i*5) + (float64(time.Now().Unix()%10) - 5),
					Unit:       getUnitForSensor(sensorType),
					Timestamp:  time.Now(),
					Location:   locations[i%len(locations)],
					Quality:    95 + (int(time.Now().Unix()) % 5),
				}

				topic := fmt.Sprintf("edgex/events/device/%s", data.DeviceID)
				if err := client.Publish(topic, data); err != nil {
					lc.Errorf("Failed to publish sensor data: %v", err)
				} else {
					lc.Debugf("Published sensor data: %s = %.2f %s", data.DeviceID, data.Value, data.Unit)
				}
			}
		}
	}
}

// handleCommands handles incoming device commands
func handleCommands(ctx context.Context, client *messagebus.Client, lc logger.LoggingClient) {
	handler := func(topic string, message types.MessageEnvelope) error {
		lc.Debugf("Received command on topic: %s", topic)

		var cmd CommandRequest
		if err := json.Unmarshal(message.Payload.([]byte), &cmd); err != nil {
			lc.Errorf("Failed to parse command: %v", err)
			return err
		}

		// Simulate command processing
		response := CommandResponse{
			RequestID: cmd.RequestID,
			DeviceID:  cmd.DeviceID,
			Success:   true,
			Result:    map[string]interface{}{"status": "executed", "value": "success"},
			Timestamp: time.Now(),
		}

		// Simulate some commands failing
		if cmd.Command == "fail" {
			response.Success = false
			response.Error = "Command execution failed"
		}

		// Publish response
		responseTopic := fmt.Sprintf("edgex/command/response/%s", cmd.DeviceID)
		if err := client.Publish(responseTopic, response); err != nil {
			lc.Errorf("Failed to publish command response: %v", err)
			return err
		}

		lc.Infof("Processed command %s for device %s", cmd.Command, cmd.DeviceID)
		return nil
	}

	// Subscribe to command topics
	topics := []string{"edgex/command/request/+", "edgex/device/+/command"}
	if err := client.Subscribe(topics, handler); err != nil {
		lc.Errorf("Failed to subscribe to command topics: %v", err)
		return
	}

	lc.Info("Command handler started")
	<-ctx.Done()
	lc.Info("Command handler stopped")
}

// subscribeToEvents subscribes to various event topics
func subscribeToEvents(ctx context.Context, client *messagebus.Client, lc logger.LoggingClient) {
	handler := func(topic string, message types.MessageEnvelope) error {
		lc.Debugf("Received event on topic: %s, CorrelationID: %s", topic, message.CorrelationID)

		// Parse and process different types of events
		var eventData map[string]interface{}
		if err := json.Unmarshal(message.Payload.([]byte), &eventData); err != nil {
			lc.Errorf("Failed to parse event data: %v", err)
			return err
		}

		// Process based on topic pattern
		switch {
		case contains(topic, "events/device"):
			lc.Infof("Device event: %+v", eventData)
		case contains(topic, "events/system"):
			lc.Infof("System event: %+v", eventData)
		default:
			lc.Infof("Other event: %+v", eventData)
		}

		return nil
	}

	// Subscribe to multiple event topics
	topics := []string{
		"edgex/events/device/+",
		"edgex/events/system/+",
		"edgex/alerts/+",
	}

	if err := client.Subscribe(topics, handler); err != nil {
		lc.Errorf("Failed to subscribe to event topics: %v", err)
		return
	}

	lc.Info("Event subscriber started")
	<-ctx.Done()
	lc.Info("Event subscriber stopped")
}

// requestResponseExample demonstrates request-response pattern
func requestResponseExample(ctx context.Context, client *messagebus.Client, lc logger.LoggingClient) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Create a command request
			cmd := CommandRequest{
				DeviceID:   "sensor-temperature-1",
				Command:    "read",
				Parameters: map[string]interface{}{"format": "json"},
				RequestID:  fmt.Sprintf("req-%d", time.Now().Unix()),
				Timestamp:  time.Now(),
			}

			// Create message envelope
			envelope, err := client.CreateMessageEnvelope(cmd, cmd.RequestID)
			if err != nil {
				lc.Errorf("Failed to create message envelope: %v", err)
				continue
			}

			// Send request and wait for response
			response, err := client.Request(
				envelope,
				"edgex/command/request/sensor-temperature-1",
				"edgex/command/response",
				10*time.Second,
			)

			if err != nil {
				lc.Errorf("Request-response failed: %v", err)
				continue
			}

			lc.Infof("Request-response successful, CorrelationID: %s", response.CorrelationID)
		}
	}
}

// Helper functions
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := fmt.Sscanf(value, "%d", &defaultValue); err == nil && intValue == 1 {
			return defaultValue
		}
	}
	return defaultValue
}

func getUnitForSensor(sensorType string) string {
	switch sensorType {
	case "temperature":
		return "°C"
	case "humidity":
		return "%"
	case "pressure":
		return "hPa"
	default:
		return ""
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr)))
}
