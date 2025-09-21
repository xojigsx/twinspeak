package srv

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"

	g "jig.sx/twinspeak/pkg/model/gemini"
)

// TestWebSocketConnection tests basic WebSocket connection establishment
func TestWebSocketConnection(t *testing.T) {
	server := New()
	httpServer := httptest.NewServer(server.Handler())
	defer httpServer.Close()

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/v1/speak"

	conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), wsURL)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Connection should be established successfully
	if conn == nil {
		t.Fatal("WebSocket connection is nil")
	}
}

// TestSetupMessageRequirement tests that setup message is required first
func TestSetupMessageRequirement(t *testing.T) {
	server := New()
	httpServer := httptest.NewServer(server.Handler())
	defer httpServer.Close()

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/v1/speak"

	conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), wsURL)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Try to send text input without setup first
	textInput := g.ClientInputTextJson{
		Type: "input_text",
		Text: "Hello without setup",
	}

	data, err := json.Marshal(textInput)
	if err != nil {
		t.Fatalf("Failed to marshal text input: %v", err)
	}

	err = wsutil.WriteClientMessage(conn, ws.OpText, data)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// Should receive error response
	msg, _, err := wsutil.ReadServerData(conn)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	var errorResp g.ErrorJson
	err = json.Unmarshal(msg, &errorResp)
	if err != nil {
		t.Fatalf("Failed to unmarshal error response: %v", err)
	}

	if errorResp.Type != "error" {
		t.Errorf("Expected error type, got %s", errorResp.Type)
	}
	if errorResp.Code != "no_session" {
		t.Errorf("Expected no_session error code, got %s", errorResp.Code)
	}
}

// TestSessionCreationAndSetup tests successful session setup
func TestSessionCreationAndSetup(t *testing.T) {
	server := New()
	httpServer := httptest.NewServer(server.Handler())
	defer httpServer.Close()

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/v1/speak"

	conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), wsURL)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Send setup request
	setupReq := g.SetupRequestJson{
		Type:  "setup",
		Model: "gemini-1.5-flash",
	}

	data, err := json.Marshal(setupReq)
	if err != nil {
		t.Fatalf("Failed to marshal setup request: %v", err)
	}

	err = wsutil.WriteClientMessage(conn, ws.OpText, data)
	if err != nil {
		t.Fatalf("Failed to send setup message: %v", err)
	}

	// Should receive session resumption update
	msg, _, err := wsutil.ReadServerData(conn)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	var resumptionUpdate g.SessionResumptionUpdateJson
	err = json.Unmarshal(msg, &resumptionUpdate)
	if err != nil {
		t.Fatalf("Failed to unmarshal resumption update: %v", err)
	}

	if resumptionUpdate.Type != "session_resumption_update" {
		t.Errorf("Expected session_resumption_update type, got %s", resumptionUpdate.Type)
	}
	if resumptionUpdate.Handle == "" {
		t.Error("Expected non-empty resumption handle")
	}
	if !strings.HasPrefix(resumptionUpdate.Handle, "session_") {
		t.Errorf("Expected resumption handle to start with 'session_', got %s", resumptionUpdate.Handle)
	}
}

// TestTextInputOutputExchange tests text message exchange
func TestTextInputOutputExchange(t *testing.T) {
	server := New()
	httpServer := httptest.NewServer(server.Handler())
	defer httpServer.Close()

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/v1/speak"

	conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), wsURL)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Setup session first
	setupReq := g.SetupRequestJson{
		Type:  "setup",
		Model: "gemini-1.5-flash",
	}

	data, err := json.Marshal(setupReq)
	if err != nil {
		t.Fatalf("Failed to marshal setup request: %v", err)
	}

	err = wsutil.WriteClientMessage(conn, ws.OpText, data)
	if err != nil {
		t.Fatalf("Failed to send setup message: %v", err)
	}

	// Read setup response
	_, _, err = wsutil.ReadServerData(conn)
	if err != nil {
		t.Fatalf("Failed to read setup response: %v", err)
	}

	// Send text input
	testText := "Hello, world!"
	textInput := g.ClientInputTextJson{
		Type: "input_text",
		Text: testText,
	}

	data, err = json.Marshal(textInput)
	if err != nil {
		t.Fatalf("Failed to marshal text input: %v", err)
	}

	err = wsutil.WriteClientMessage(conn, ws.OpText, data)
	if err != nil {
		t.Fatalf("Failed to send text input: %v", err)
	}

	// Should receive echo response
	msg, _, err := wsutil.ReadServerData(conn)
	if err != nil {
		t.Fatalf("Failed to read text response: %v", err)
	}

	var textOutput g.ServerOutputTextJson
	err = json.Unmarshal(msg, &textOutput)
	if err != nil {
		t.Fatalf("Failed to unmarshal text output: %v", err)
	}

	if textOutput.Type != "output_text" {
		t.Errorf("Expected output_text type, got %s", textOutput.Type)
	}
	expectedText := "[echo] " + testText
	if textOutput.Text != expectedText {
		t.Errorf("Expected text '%s', got '%s'", expectedText, textOutput.Text)
	}
	if !textOutput.Final {
		t.Error("Expected final to be true")
	}
}

// TestAudioInputAcknowledgment tests audio input handling
func TestAudioInputAcknowledgment(t *testing.T) {
	server := New()
	httpServer := httptest.NewServer(server.Handler())
	defer httpServer.Close()

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/v1/speak"

	conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), wsURL)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Setup session first
	setupReq := g.SetupRequestJson{
		Type:  "setup",
		Model: "gemini-1.5-flash",
	}

	data, err := json.Marshal(setupReq)
	if err != nil {
		t.Fatalf("Failed to marshal setup request: %v", err)
	}

	err = wsutil.WriteClientMessage(conn, ws.OpText, data)
	if err != nil {
		t.Fatalf("Failed to send setup message: %v", err)
	}

	// Read setup response
	_, _, err = wsutil.ReadServerData(conn)
	if err != nil {
		t.Fatalf("Failed to read setup response: %v", err)
	}

	// Send audio input
	audioInput := g.ClientInputAudioJson{
		Type:   "input_audio",
		Format: g.ClientInputAudioJsonFormatWav,
		Chunk:  "dGVzdCBhdWRpbyBkYXRh", // base64 encoded "test audio data"
		Final:  true,
	}

	data, err = json.Marshal(audioInput)
	if err != nil {
		t.Fatalf("Failed to marshal audio input: %v", err)
	}

	err = wsutil.WriteClientMessage(conn, ws.OpText, data)
	if err != nil {
		t.Fatalf("Failed to send audio input: %v", err)
	}

	// Should receive acknowledgment response
	msg, _, err := wsutil.ReadServerData(conn)
	if err != nil {
		t.Fatalf("Failed to read audio response: %v", err)
	}

	var textOutput g.ServerOutputTextJson
	err = json.Unmarshal(msg, &textOutput)
	if err != nil {
		t.Fatalf("Failed to unmarshal audio acknowledgment: %v", err)
	}

	if textOutput.Type != "output_text" {
		t.Errorf("Expected output_text type, got %s", textOutput.Type)
	}
	expectedText := "Received audio chunk in wav format (final: true)"
	if textOutput.Text != expectedText {
		t.Errorf("Expected text '%s', got '%s'", expectedText, textOutput.Text)
	}
	if !textOutput.Final {
		t.Error("Expected final to be true")
	}
}

// TestSessionEndFlowAndCleanup tests session termination
func TestSessionEndFlowAndCleanup(t *testing.T) {
	server := New()
	httpServer := httptest.NewServer(server.Handler())
	defer httpServer.Close()

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/v1/speak"

	conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), wsURL)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Setup session first
	setupReq := g.SetupRequestJson{
		Type:  "setup",
		Model: "gemini-1.5-flash",
	}

	data, err := json.Marshal(setupReq)
	if err != nil {
		t.Fatalf("Failed to marshal setup request: %v", err)
	}

	err = wsutil.WriteClientMessage(conn, ws.OpText, data)
	if err != nil {
		t.Fatalf("Failed to send setup message: %v", err)
	}

	// Read setup response
	_, _, err = wsutil.ReadServerData(conn)
	if err != nil {
		t.Fatalf("Failed to read setup response: %v", err)
	}

	// Verify session exists in store (we need to access the session ID somehow)
	// For now, we'll just verify the session end flow works

	// Send session end
	sessionEnd := g.SessionEndJson{
		Type:   "end_session",
		Reason: "user_requested",
	}

	data, err = json.Marshal(sessionEnd)
	if err != nil {
		t.Fatalf("Failed to marshal session end: %v", err)
	}

	err = wsutil.WriteClientMessage(conn, ws.OpText, data)
	if err != nil {
		t.Fatalf("Failed to send session end: %v", err)
	}

	// Should receive goodbye message
	msg, _, err := wsutil.ReadServerData(conn)
	if err != nil {
		t.Fatalf("Failed to read goodbye response: %v", err)
	}

	var textOutput g.ServerOutputTextJson
	err = json.Unmarshal(msg, &textOutput)
	if err != nil {
		t.Fatalf("Failed to unmarshal goodbye message: %v", err)
	}

	if textOutput.Type != "output_text" {
		t.Errorf("Expected output_text type, got %s", textOutput.Type)
	}
	expectedText := "Goodbye! Session ended."
	if textOutput.Text != expectedText {
		t.Errorf("Expected text '%s', got '%s'", expectedText, textOutput.Text)
	}
	if !textOutput.Final {
		t.Error("Expected final to be true")
	}

	// Connection should be closed by server after session end
	// Try to read again - should get EOF or connection closed error
	_, _, err = wsutil.ReadServerData(conn)
	if err == nil {
		t.Error("Expected connection to be closed after session end")
	}
}

// TestErrorConditions tests various error scenarios
func TestErrorConditions(t *testing.T) {
	tests := []struct {
		name         string
		setupFirst   bool
		message      string
		expectedCode string
		expectedType string
	}{
		{
			name:         "Invalid JSON",
			setupFirst:   false,
			message:      `{"invalid": json}`,
			expectedCode: "bad_json",
			expectedType: "error",
		},
		{
			name:         "Unknown message type",
			setupFirst:   true,
			message:      `{"type": "unknown_message"}`,
			expectedCode: "unknown_type",
			expectedType: "error",
		},
		{
			name:         "Duplicate setup",
			setupFirst:   true,
			message:      `{"type": "setup", "model": "gemini-1.5-flash"}`,
			expectedCode: "already_setup",
			expectedType: "error",
		},
		{
			name:         "Invalid setup format",
			setupFirst:   false,
			message:      `{"type": "setup"}`,
			expectedCode: "bad_setup",
			expectedType: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := New()
			httpServer := httptest.NewServer(server.Handler())
			defer httpServer.Close()

			wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/v1/speak"

			conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), wsURL)
			if err != nil {
				t.Fatalf("Failed to connect to WebSocket: %v", err)
			}
			defer conn.Close()

			// Setup session if required
			if tt.setupFirst {
				setupReq := g.SetupRequestJson{
					Type:  "setup",
					Model: "gemini-1.5-flash",
				}

				data, err := json.Marshal(setupReq)
				if err != nil {
					t.Fatalf("Failed to marshal setup request: %v", err)
				}

				err = wsutil.WriteClientMessage(conn, ws.OpText, data)
				if err != nil {
					t.Fatalf("Failed to send setup message: %v", err)
				}

				// Read setup response
				_, _, err = wsutil.ReadServerData(conn)
				if err != nil {
					t.Fatalf("Failed to read setup response: %v", err)
				}
			}

			// Send the test message
			err = wsutil.WriteClientMessage(conn, ws.OpText, []byte(tt.message))
			if err != nil {
				t.Fatalf("Failed to send test message: %v", err)
			}

			// Should receive error response
			msg, _, err := wsutil.ReadServerData(conn)
			if err != nil {
				t.Fatalf("Failed to read error response: %v", err)
			}

			var errorResp g.ErrorJson
			err = json.Unmarshal(msg, &errorResp)
			if err != nil {
				t.Fatalf("Failed to unmarshal error response: %v", err)
			}

			if errorResp.Type != tt.expectedType {
				t.Errorf("Expected type %s, got %s", tt.expectedType, errorResp.Type)
			}
			if errorResp.Code != tt.expectedCode {
				t.Errorf("Expected code %s, got %s", tt.expectedCode, errorResp.Code)
			}
			if errorResp.Message == "" {
				t.Error("Expected non-empty error message")
			}
		})
	}
}

// TestToolResultHandling tests tool result message processing
func TestToolResultHandling(t *testing.T) {
	server := New()
	httpServer := httptest.NewServer(server.Handler())
	defer httpServer.Close()

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/v1/speak"

	conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), wsURL)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Setup session first
	setupReq := g.SetupRequestJson{
		Type:  "setup",
		Model: "gemini-1.5-flash",
	}

	data, err := json.Marshal(setupReq)
	if err != nil {
		t.Fatalf("Failed to marshal setup request: %v", err)
	}

	err = wsutil.WriteClientMessage(conn, ws.OpText, data)
	if err != nil {
		t.Fatalf("Failed to send setup message: %v", err)
	}

	// Read setup response
	_, _, err = wsutil.ReadServerData(conn)
	if err != nil {
		t.Fatalf("Failed to read setup response: %v", err)
	}

	// Send tool result
	toolResult := g.ToolResultJson{
		Type:   "tool_result",
		Name:   "test_tool",
		CallId: "call_123",
		Result: map[string]interface{}{"status": "success", "data": "test result"},
	}

	data, err = json.Marshal(toolResult)
	if err != nil {
		t.Fatalf("Failed to marshal tool result: %v", err)
	}

	err = wsutil.WriteClientMessage(conn, ws.OpText, data)
	if err != nil {
		t.Fatalf("Failed to send tool result: %v", err)
	}

	// Tool result should be processed without response (just logged)
	// We can verify this by sending another message and ensuring the connection is still active
	textInput := g.ClientInputTextJson{
		Type: "input_text",
		Text: "Test after tool result",
	}

	data, err = json.Marshal(textInput)
	if err != nil {
		t.Fatalf("Failed to marshal text input: %v", err)
	}

	err = wsutil.WriteClientMessage(conn, ws.OpText, data)
	if err != nil {
		t.Fatalf("Failed to send text input after tool result: %v", err)
	}

	// Should receive echo response, confirming connection is still active
	msg, _, err := wsutil.ReadServerData(conn)
	if err != nil {
		t.Fatalf("Failed to read text response after tool result: %v", err)
	}

	var textOutput g.ServerOutputTextJson
	err = json.Unmarshal(msg, &textOutput)
	if err != nil {
		t.Fatalf("Failed to unmarshal text output: %v", err)
	}

	if textOutput.Type != "output_text" {
		t.Errorf("Expected output_text type, got %s", textOutput.Type)
	}
}

// TestConcurrentSessions tests multiple concurrent WebSocket sessions
func TestConcurrentSessions(t *testing.T) {
	server := New()
	httpServer := httptest.NewServer(server.Handler())
	defer httpServer.Close()

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/v1/speak"

	// Create multiple concurrent connections
	numConnections := 3
	connections := make([]net.Conn, numConnections)
	defer func() {
		for _, conn := range connections {
			if conn != nil {
				conn.Close()
			}
		}
	}()

	// Establish connections and setup sessions
	for i := 0; i < numConnections; i++ {
		conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), wsURL)
		if err != nil {
			t.Fatalf("Failed to connect to WebSocket %d: %v", i, err)
		}
		connections[i] = conn

		// Setup session
		setupReq := g.SetupRequestJson{
			Type:  "setup",
			Model: "gemini-1.5-flash",
		}

		data, err := json.Marshal(setupReq)
		if err != nil {
			t.Fatalf("Failed to marshal setup request %d: %v", i, err)
		}

		err = wsutil.WriteClientMessage(conn, ws.OpText, data)
		if err != nil {
			t.Fatalf("Failed to send setup message %d: %v", i, err)
		}

		// Read setup response
		_, _, err = wsutil.ReadServerData(conn)
		if err != nil {
			t.Fatalf("Failed to read setup response %d: %v", i, err)
		}
	}

	// Send messages concurrently and verify responses
	done := make(chan bool, numConnections)

	for i := 0; i < numConnections; i++ {
		go func(connIndex int) {
			defer func() { done <- true }()

			conn := connections[connIndex]
			testText := fmt.Sprintf("Hello from connection %d", connIndex)

			textInput := g.ClientInputTextJson{
				Type: "input_text",
				Text: testText,
			}

			data, err := json.Marshal(textInput)
			if err != nil {
				t.Errorf("Failed to marshal text input %d: %v", connIndex, err)
				return
			}

			err = wsutil.WriteClientMessage(conn, ws.OpText, data)
			if err != nil {
				t.Errorf("Failed to send text input %d: %v", connIndex, err)
				return
			}

			// Read response
			msg, _, err := wsutil.ReadServerData(conn)
			if err != nil {
				t.Errorf("Failed to read response %d: %v", connIndex, err)
				return
			}

			var textOutput g.ServerOutputTextJson
			err = json.Unmarshal(msg, &textOutput)
			if err != nil {
				t.Errorf("Failed to unmarshal response %d: %v", connIndex, err)
				return
			}

			expectedText := "[echo] " + testText
			if textOutput.Text != expectedText {
				t.Errorf("Connection %d: expected '%s', got '%s'", connIndex, expectedText, textOutput.Text)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	timeout := time.After(5 * time.Second)
	for i := 0; i < numConnections; i++ {
		select {
		case <-done:
			// Success
		case <-timeout:
			t.Fatal("Timeout waiting for concurrent sessions to complete")
		}
	}
}
