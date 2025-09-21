package srv

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"

	g "jig.sx/twinspeak/pkg/model/gemini"
)

// TestCompleteSessionFlow tests the complete session lifecycle
func TestCompleteSessionFlow(t *testing.T) {
	server := New()
	httpServer := httptest.NewServer(server.Handler())
	defer httpServer.Close()

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/v1/speak"

	conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), wsURL)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Step 1: Setup session
	setupReq := g.SetupRequestJson{
		Type:  "setup",
		Model: "gemini-1.5-flash",
		SessionConfig: map[string]interface{}{
			"temperature": 0.7,
			"maxTokens":   1000,
		},
	}

	data, err := json.Marshal(setupReq)
	if err != nil {
		t.Fatalf("Failed to marshal setup request: %v", err)
	}

	err = wsutil.WriteClientMessage(conn, ws.OpText, data)
	if err != nil {
		t.Fatalf("Failed to send setup message: %v", err)
	}

	// Verify session resumption update
	msg, _, err := wsutil.ReadServerData(conn)
	if err != nil {
		t.Fatalf("Failed to read setup response: %v", err)
	}

	var resumptionUpdate g.SessionResumptionUpdateJson
	err = json.Unmarshal(msg, &resumptionUpdate)
	if err != nil {
		t.Fatalf("Failed to unmarshal resumption update: %v", err)
	}

	if resumptionUpdate.Type != "session_resumption_update" {
		t.Errorf("Expected session_resumption_update, got %s", resumptionUpdate.Type)
	}

	// Step 2: Send text input
	textInput := g.ClientInputTextJson{
		Type:   "input_text",
		Text:   "What is the weather like?",
		TurnId: stringPtr("turn_001"),
	}

	data, err = json.Marshal(textInput)
	if err != nil {
		t.Fatalf("Failed to marshal text input: %v", err)
	}

	err = wsutil.WriteClientMessage(conn, ws.OpText, data)
	if err != nil {
		t.Fatalf("Failed to send text input: %v", err)
	}

	// Verify text response
	msg, _, err = wsutil.ReadServerData(conn)
	if err != nil {
		t.Fatalf("Failed to read text response: %v", err)
	}

	var textOutput g.ServerOutputTextJson
	err = json.Unmarshal(msg, &textOutput)
	if err != nil {
		t.Fatalf("Failed to unmarshal text output: %v", err)
	}

	if textOutput.Type != "output_text" {
		t.Errorf("Expected output_text, got %s", textOutput.Type)
	}

	// Step 3: Send audio input with multiple chunks
	audioFormats := []g.ClientInputAudioJsonFormat{
		g.ClientInputAudioJsonFormatWav,
		g.ClientInputAudioJsonFormatPcm16,
		g.ClientInputAudioJsonFormatOpus,
	}

	for i, format := range audioFormats {
		audioInput := g.ClientInputAudioJson{
			Type:   "input_audio",
			Format: format,
			Chunk:  "dGVzdCBhdWRpbyBkYXRh", // base64 encoded "test audio data"
			Final:  i == len(audioFormats)-1,
		}

		data, err = json.Marshal(audioInput)
		if err != nil {
			t.Fatalf("Failed to marshal audio input: %v", err)
		}

		err = wsutil.WriteClientMessage(conn, ws.OpText, data)
		if err != nil {
			t.Fatalf("Failed to send audio input: %v", err)
		}

		// Verify audio acknowledgment
		msg, _, err = wsutil.ReadServerData(conn)
		if err != nil {
			t.Fatalf("Failed to read audio response: %v", err)
		}

		var audioAck g.ServerOutputTextJson
		err = json.Unmarshal(msg, &audioAck)
		if err != nil {
			t.Fatalf("Failed to unmarshal audio acknowledgment: %v", err)
		}

		if audioAck.Type != "output_text" {
			t.Errorf("Expected output_text for audio ack, got %s", audioAck.Type)
		}
	}

	// Step 4: Send tool result
	toolResult := g.ToolResultJson{
		Type:   "tool_result",
		Name:   "weather_api",
		CallId: "call_weather_001",
		Result: map[string]interface{}{
			"temperature": 22,
			"condition":   "sunny",
			"humidity":    65,
		},
	}

	data, err = json.Marshal(toolResult)
	if err != nil {
		t.Fatalf("Failed to marshal tool result: %v", err)
	}

	err = wsutil.WriteClientMessage(conn, ws.OpText, data)
	if err != nil {
		t.Fatalf("Failed to send tool result: %v", err)
	}

	// Step 5: End session
	sessionEnd := g.SessionEndJson{
		Type:   "end_session",
		Reason: "conversation_complete",
	}

	data, err = json.Marshal(sessionEnd)
	if err != nil {
		t.Fatalf("Failed to marshal session end: %v", err)
	}

	err = wsutil.WriteClientMessage(conn, ws.OpText, data)
	if err != nil {
		t.Fatalf("Failed to send session end: %v", err)
	}

	// Verify goodbye message
	msg, _, err = wsutil.ReadServerData(conn)
	if err != nil {
		t.Fatalf("Failed to read goodbye response: %v", err)
	}

	var goodbye g.ServerOutputTextJson
	err = json.Unmarshal(msg, &goodbye)
	if err != nil {
		t.Fatalf("Failed to unmarshal goodbye message: %v", err)
	}

	if goodbye.Type != "output_text" {
		t.Errorf("Expected output_text for goodbye, got %s", goodbye.Type)
	}
	if goodbye.Text != "Goodbye! Session ended." {
		t.Errorf("Expected goodbye message, got %s", goodbye.Text)
	}
}

// TestSessionStateTransitions tests that session states change correctly
func TestSessionStateTransitions(t *testing.T) {
	server := New()
	httpServer := httptest.NewServer(server.Handler())
	defer httpServer.Close()

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/v1/speak"

	conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), wsURL)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Setup session - should transition to Configured state
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

	// Send text input - should transition to Active state
	textInput := g.ClientInputTextJson{
		Type: "input_text",
		Text: "Hello",
	}

	data, err = json.Marshal(textInput)
	if err != nil {
		t.Fatalf("Failed to marshal text input: %v", err)
	}

	err = wsutil.WriteClientMessage(conn, ws.OpText, data)
	if err != nil {
		t.Fatalf("Failed to send text input: %v", err)
	}

	// Read text response
	_, _, err = wsutil.ReadServerData(conn)
	if err != nil {
		t.Fatalf("Failed to read text response: %v", err)
	}

	// Send session end - should transition to Closing then Closed
	sessionEnd := g.SessionEndJson{
		Type:   "end_session",
		Reason: "test_complete",
	}

	data, err = json.Marshal(sessionEnd)
	if err != nil {
		t.Fatalf("Failed to marshal session end: %v", err)
	}

	err = wsutil.WriteClientMessage(conn, ws.OpText, data)
	if err != nil {
		t.Fatalf("Failed to send session end: %v", err)
	}

	// Read goodbye response
	_, _, err = wsutil.ReadServerData(conn)
	if err != nil {
		t.Fatalf("Failed to read goodbye response: %v", err)
	}

	// Connection should be closed after session end
	_, _, err = wsutil.ReadServerData(conn)
	if err == nil {
		t.Error("Expected connection to be closed after session end")
	}
}

// TestInvalidMessageFormats tests various invalid message formats
func TestInvalidMessageFormats(t *testing.T) {
	tests := []struct {
		name         string
		setupFirst   bool
		message      string
		expectedCode string
	}{
		{
			name:         "Missing required field in setup",
			setupFirst:   false,
			message:      `{"type": "setup"}`,
			expectedCode: "bad_setup",
		},
		{
			name:         "Missing required field in text input",
			setupFirst:   true,
			message:      `{"type": "input_text"}`,
			expectedCode: "bad_json",
		},
		{
			name:         "Missing required field in audio input",
			setupFirst:   true,
			message:      `{"type": "input_audio", "format": "wav"}`,
			expectedCode: "bad_json",
		},
		{
			name:         "Missing required field in tool result",
			setupFirst:   true,
			message:      `{"type": "tool_result", "name": "test"}`,
			expectedCode: "bad_json",
		},
		{
			name:         "Missing required field in session end",
			setupFirst:   true,
			message:      `{"type": "end_session"}`,
			expectedCode: "bad_json",
		},
		{
			name:         "Invalid audio format",
			setupFirst:   true,
			message:      `{"type": "input_audio", "format": "invalid", "chunk": "data", "final": true}`,
			expectedCode: "bad_json",
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

			// Send the invalid message
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

			if errorResp.Type != "error" {
				t.Errorf("Expected error type, got %s", errorResp.Type)
			}
			if errorResp.Code != tt.expectedCode {
				t.Errorf("Expected code %s, got %s", tt.expectedCode, errorResp.Code)
			}
		})
	}
}

// TestNonTextWebSocketMessages tests that non-text WebSocket messages are rejected
func TestNonTextWebSocketMessages(t *testing.T) {
	server := New()
	httpServer := httptest.NewServer(server.Handler())
	defer httpServer.Close()

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/v1/speak"

	conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), wsURL)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Send binary message (should be rejected)
	err = wsutil.WriteClientMessage(conn, ws.OpBinary, []byte("binary data"))
	if err != nil {
		t.Fatalf("Failed to send binary message: %v", err)
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

	if errorResp.Type != "error" {
		t.Errorf("Expected error type, got %s", errorResp.Type)
	}
	if errorResp.Code != "bad_json" {
		t.Errorf("Expected bad_json error code, got %s", errorResp.Code)
	}
	if errorResp.Message != "Only text messages are supported" {
		t.Errorf("Expected specific error message, got %s", errorResp.Message)
	}
}

// TestHealthEndpoint tests the health check endpoint
func TestHealthEndpoint(t *testing.T) {
	server := New()
	httpServer := httptest.NewServer(server.Handler())
	defer httpServer.Close()

	resp, err := httpServer.Client().Get(httpServer.URL + "/healthz")
	if err != nil {
		t.Fatalf("Failed to call health endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var body []byte
	body = make([]byte, 1024)
	n, _ := resp.Body.Read(body)
	bodyStr := string(body[:n])

	expected := `{"status":"ok"}`
	if bodyStr != expected {
		t.Errorf("Expected body %s, got %s", expected, bodyStr)
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
