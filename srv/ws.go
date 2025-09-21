package srv

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	g "jig.sx/twinspeak/pkg/model/gemini"
	"jig.sx/twinspeak/pkg/session"
)

// envelope represents the message envelope for type-based routing
type envelope struct {
	Type string `json:"type"`
}

// handleSpeakWS handles WebSocket upgrade and message processing
func (s *Server) handleSpeakWS(w http.ResponseWriter, r *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	var sess *session.Session
	var sessionID session.ID

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		msg, op, err := wsutil.ReadClientData(conn)
		if err != nil {
			log.Printf("Failed to read WebSocket message: %v", err)
			return
		}

		if op != ws.OpText {
			s.sendError(conn, "bad_json", "Only text messages are supported")
			continue
		}

		var env envelope
		if err := json.Unmarshal(msg, &env); err != nil {
			s.sendError(conn, "bad_json", "Invalid JSON format")
			continue
		}

		switch env.Type {
		case "setup":
			if sess != nil {
				s.sendError(conn, "already_setup", "Session already configured")
				continue
			}

			var setupReq g.SetupRequestJson
			if err := json.Unmarshal(msg, &setupReq); err != nil {
				s.sendError(conn, "bad_setup", "Invalid setup request format")
				continue
			}

			sess = session.NewSession(setupReq.Model)
			sess.State = session.StateConfigured
			sess.ResumptionHandle = fmt.Sprintf("session_%s", sess.ID)
			sessionID = sess.ID

			s.Store.Put(sess)
			sess.Append(setupReq)

			resumptionUpdate := g.SessionResumptionUpdateJson{
				Type:   "session_resumption_update",
				Handle: sess.ResumptionHandle,
			}
			s.writeJSON(conn, resumptionUpdate)

		case "input_text":
			if sess == nil {
				s.sendError(conn, "no_session", "No active session")
				continue
			}

			var textInput g.ClientInputTextJson
			if err := json.Unmarshal(msg, &textInput); err != nil {
				s.sendError(conn, "bad_json", "Invalid text input format")
				continue
			}

			sess.State = session.StateActive
			sess.Append(textInput)

			echoResponse := g.ServerOutputTextJson{
				Type:  "output_text",
				Text:  fmt.Sprintf("[echo] %s", textInput.Text),
				Final: true,
			}
			s.writeJSON(conn, echoResponse)

		case "input_audio":
			if sess == nil {
				s.sendError(conn, "no_session", "No active session")
				continue
			}

			var audioInput g.ClientInputAudioJson
			if err := json.Unmarshal(msg, &audioInput); err != nil {
				s.sendError(conn, "bad_json", "Invalid audio input format")
				continue
			}

			sess.State = session.StateActive
			sess.Append(audioInput)

			ackResponse := g.ServerOutputTextJson{
				Type:  "output_text",
				Text:  fmt.Sprintf("Received audio chunk in %s format (final: %t)", audioInput.Format, audioInput.Final),
				Final: true,
			}
			s.writeJSON(conn, ackResponse)

		case "tool_result":
			if sess == nil {
				s.sendError(conn, "no_session", "No active session")
				continue
			}

			var toolResult g.ToolResultJson
			if err := json.Unmarshal(msg, &toolResult); err != nil {
				s.sendError(conn, "bad_json", "Invalid tool result format")
				continue
			}

			sess.Append(toolResult)

		case "end_session":
			if sess == nil {
				s.sendError(conn, "no_session", "No active session")
				continue
			}

			var endSession g.SessionEndJson
			if err := json.Unmarshal(msg, &endSession); err != nil {
				s.sendError(conn, "bad_json", "Invalid session end format")
				continue
			}

			sess.State = session.StateClosing
			sess.Append(endSession)

			goodbyeResponse := g.ServerOutputTextJson{
				Type:  "output_text",
				Text:  "Goodbye! Session ended.",
				Final: true,
			}
			s.writeJSON(conn, goodbyeResponse)

			sess.State = session.StateClosed
			s.Store.Delete(sessionID)
			return

		default:
			s.sendError(conn, "unknown_type", fmt.Sprintf("Unknown message type: %s", env.Type))
		}
	}
}

// writeJSON writes a JSON message to the WebSocket connection
func (s *Server) writeJSON(conn net.Conn, v any) error {
	data := s.mustJSON(v)
	return wsutil.WriteServerMessage(conn, ws.OpText, data)
}

// sendError sends a structured error message to the client
func (s *Server) sendError(conn net.Conn, code, message string) {
	errorMsg := g.ErrorJson{
		Type:    "error",
		Code:    code,
		Message: message,
	}
	if err := s.writeJSON(conn, errorMsg); err != nil {
		log.Printf("Failed to send error message: %v", err)
	}
}

// ensure panics if the error is not nil
func (s *Server) ensure(err error) {
	if err != nil {
		panic(err)
	}
}

// mustJSON marshals v to JSON, panicking on error
func (s *Server) mustJSON(v any) []byte {
	data, err := json.Marshal(v)
	s.ensure(err)
	return data
}
