package messagebus

import (
	"testing"
	"time"

	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
)

// MockLoggingClient is a mock implementation of logger.LoggingClient for testing
type MockLoggingClient struct{}

func (m *MockLoggingClient) Trace(msg string, args ...interface{}) {}
func (m *MockLoggingClient) Debug(msg string, args ...interface{}) {}
func (m *MockLoggingClient) Info(msg string, args ...interface{})  {}
func (m *MockLoggingClient) Warn(msg string, args ...interface{})  {}
func (m *MockLoggingClient) Error(msg string, args ...interface{}) {}

func (m *MockLoggingClient) Tracef(msg string, args ...interface{}) {}
func (m *MockLoggingClient) Debugf(msg string, args ...interface{}) {}
func (m *MockLoggingClient) Infof(msg string, args ...interface{})  {}
func (m *MockLoggingClient) Warnf(msg string, args ...interface{})  {}
func (m *MockLoggingClient) Errorf(msg string, args ...interface{}) {}

func (m *MockLoggingClient) SetLogLevel(logLevel string) error { return nil }
func (m *MockLoggingClient) LogLevel() string                  { return "DEBUG" }

func TestNewClient(t *testing.T) {
	config := Config{
		Host:     "localhost",
		Port:     1883,
		Protocol: "tcp",
		Type:     "mqtt",
		ClientID: "test-client",
	}

	lc := &MockLoggingClient{}

	client, err := NewClient(config, lc)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}

	if client.IsConnected() {
		t.Error("Expected client to not be connected initially")
	}
}

func TestConfig(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		valid  bool
	}{
		{
			name: "valid MQTT config",
			config: Config{
				Host:     "localhost",
				Port:     1883,
				Protocol: "tcp",
				Type:     "mqtt",
				ClientID: "test-client",
			},
			valid: true,
		},
		{
			name: "valid NATS config",
			config: Config{
				Host:     "localhost",
				Port:     4222,
				Protocol: "tcp",
				Type:     "nats",
				ClientID: "test-client",
			},
			valid: false, // NATS may not be supported in this EdgeX version
		},
		{
			name: "config with authentication",
			config: Config{
				Host:     "localhost",
				Port:     1883,
				Protocol: "tcp",
				Type:     "mqtt",
				ClientID: "test-client",
				Username: "user",
				Password: "pass",
				QoS:      1,
			},
			valid: true,
		},
	}

	lc := &MockLoggingClient{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config, lc)
			if tt.valid {
				if err != nil {
					t.Errorf("Expected no error for valid config, got %v", err)
				}
				if client == nil {
					t.Error("Expected client to be created for valid config")
				}
			} else {
				if err == nil {
					t.Error("Expected error for invalid config")
				}
			}
		})
	}
}

func TestCreateMessageEnvelope(t *testing.T) {
	config := Config{
		Host:     "localhost",
		Port:     1883,
		Protocol: "tcp",
		Type:     "mqtt",
		ClientID: "test-client",
	}

	lc := &MockLoggingClient{}
	client, err := NewClient(config, lc)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tests := []struct {
		name          string
		data          interface{}
		correlationID string
		expectError   bool
	}{
		{
			name:          "string data",
			data:          "test message",
			correlationID: "test-correlation-id",
			expectError:   false,
		},
		{
			name:          "byte data",
			data:          []byte("test message"),
			correlationID: "",
			expectError:   false,
		},
		{
			name: "struct data",
			data: map[string]interface{}{
				"temperature": 25.6,
				"timestamp":   time.Now(),
			},
			correlationID: "struct-test",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envelope, err := client.CreateMessageEnvelope(tt.data, tt.correlationID)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if envelope.Payload == nil {
				t.Error("Expected payload to be set")
			}

			if tt.correlationID != "" && envelope.CorrelationID != tt.correlationID {
				t.Errorf("Expected CorrelationID %s, got %s", tt.correlationID, envelope.CorrelationID)
			}

			if tt.correlationID == "" && envelope.CorrelationID == "" {
				t.Error("Expected auto-generated CorrelationID")
			}

			if envelope.ContentType != "application/json" {
				t.Errorf("Expected ContentType 'application/json', got %s", envelope.ContentType)
			}
		})
	}
}

func TestClientInfo(t *testing.T) {
	config := Config{
		Host:     "localhost",
		Port:     1883,
		Protocol: "tcp",
		Type:     "mqtt",
		ClientID: "test-client",
	}

	lc := &MockLoggingClient{}
	client, err := NewClient(config, lc)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	info := client.GetClientInfo()

	if info == nil {
		t.Fatal("Expected client info to be returned")
	}

	// Check that expected keys exist
	expectedKeys := []string{"connected", "subscribedTopics", "errorChannelBuffer"}
	for _, key := range expectedKeys {
		if _, exists := info[key]; !exists {
			t.Errorf("Expected key %s to exist in client info", key)
		}
	}

	// Check initial values
	if info["connected"] != false {
		t.Error("Expected connected to be false initially")
	}

	if info["subscribedTopics"] != 0 {
		t.Error("Expected subscribedTopics to be 0 initially")
	}
}

func TestGetSubscribedTopics(t *testing.T) {
	config := Config{
		Host:     "localhost",
		Port:     1883,
		Protocol: "tcp",
		Type:     "mqtt",
		ClientID: "test-client",
	}

	lc := &MockLoggingClient{}
	client, err := NewClient(config, lc)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	topics := client.GetSubscribedTopics()
	if len(topics) != 0 {
		t.Errorf("Expected 0 subscribed topics initially, got %d", len(topics))
	}
}

func TestHealthCheck(t *testing.T) {
	config := Config{
		Host:     "localhost",
		Port:     1883,
		Protocol: "tcp",
		Type:     "mqtt",
		ClientID: "test-client",
	}

	lc := &MockLoggingClient{}
	client, err := NewClient(config, lc)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Health check should fail when not connected
	err = client.HealthCheck()
	if err == nil {
		t.Error("Expected health check to fail when not connected")
	}
}

func TestMessageHandler(t *testing.T) {
	// Test that MessageHandler type is properly defined
	var handler MessageHandler = func(topic string, message types.MessageEnvelope) error {
		return nil
	}

	if handler == nil {
		t.Error("MessageHandler should be assignable")
	}

	// Test handler execution
	testTopic := "test/topic"
	testMessage := types.MessageEnvelope{
		CorrelationID: "test-id",
		Payload:       []byte("test payload"),
		ContentType:   "application/json",
	}

	err := handler(testTopic, testMessage)
	if err != nil {
		t.Errorf("Handler should not return error, got %v", err)
	}
}
