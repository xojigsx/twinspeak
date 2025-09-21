# Implementation Plan

- [x] 1. Initialize Go module and dependencies
  - Create go.mod file with module name "jig.sx/twinspeak"
  - Add required dependencies: go-chi/chi/v5, gobwas/ws, spf13/cobra, google/uuid
  - Create tools.go file with build constraints for atombender/go-jsonschema
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7_

- [x] 2. Create API specifications and JSON schemas
  - Create api/gemini.json AsyncAPI 3.0 specification with WebSocket endpoint definition
  - Create api/models/gemini/ directory structure
  - Implement SetupRequest.json schema with type, model, and sessionConfig fields
  - Implement ClientInputText.json schema with type, text, and turnId fields
  - Implement ClientInputAudio.json schema with type, format, chunk, and final fields
  - Implement ToolResult.json schema with type, name, callId, and result fields
  - Implement SessionEnd.json schema with type and reason fields
  - Implement ServerOutputText.json schema with type, text, and final fields
  - Implement ServerOutputAudio.json schema with type, format, chunk, and final fields
  - Implement FunctionCall.json schema with type, name, callId, and arguments fields
  - Implement SessionResumptionUpdate.json schema with type and handle fields
  - Implement Error.json schema with type, code, and message fields
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7, 2.8, 2.9_

- [x] 3. Set up code generation infrastructure
  - Create pkg/model/model.go with go:generate directives for JSON schema processing
  - Create pkg/model/gemini/.keep placeholder file
  - Configure go-jsonschema to generate structs in gemini package
  - Test code generation by running go generate ./pkg/model
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

- [x] 4. Implement session management and state machine
  - Create pkg/session/state.go with State enum and String() method
  - Implement session states: Connecting, Configured, Active, Closing, Closed
  - Create pkg/session/session.go with Session struct and ID type
  - Implement Session constructor with UUID generation and timestamp initialization
  - Implement thread-safe Append method for message logging
  - Create Store struct with thread-safe Put, Get, and Delete operations
  - Implement NewStore constructor for session store initialization
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6_

- [x] 5. Create HTTP server with Chi router
  - Create srv/srv.go with Server struct containing Store and Chi mux
  - Implement New() constructor for server initialization
  - Set up routes() method with health check endpoint at "/healthz"
  - Add WebSocket upgrade route at "/v1/speak"
  - Implement Handler() method returning Chi router
  - _Requirements: 5.1, 5.2, 5.3, 5.4_

- [x] 6. Implement WebSocket message handling
  - Create srv/ws.go with envelope struct for message type routing
  - Implement handleSpeakWS method with WebSocket upgrade using gobwas/ws
  - Create message processing loop with context cancellation support
  - Implement SetupRequest handling with session creation and state transition
  - Add SessionResumptionUpdate emission after successful setup
  - Implement ClientInputText handling with echo response functionality
  - Implement ClientInputAudio handling with acknowledgment responses
  - Implement ToolResult handling with message logging
  - Implement SessionEnd handling with state transition and goodbye message
  - Add error handling for unknown message types and invalid JSON
  - Create helper methods: writeJSON, sendError, ensure, mustJSON
  - _Requirements: 5.5, 6.1, 6.2, 6.3, 6.4, 6.5, 6.6, 6.7, 6.8, 6.9_

- [x] 7. Create Cobra CLI interface
  - Create cmd/twinspeak/main.go with Cobra root command
  - Implement command-line flag for --addr with default ":8080"
  - Add server startup logic with address logging
  - Integrate HTTP server with CLI command execution
  - Add error handling for command execution
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5_

- [x] 8. Add build automation and documentation
  - Create Makefile with gen target for code generation
  - Add run target for server startup on port 8080
  - Create README.md with quick start instructions
  - Document WebSocket protocol flow and example messages
  - Add AsyncAPI specification reference in documentation
  - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_

- [x] 9. Implement integration tests for WebSocket protocol
  - Create test files for WebSocket connection and message flow
  - Test setup message requirement and session creation
  - Test text input/output message exchange
  - Test audio input acknowledgment
  - Test session end flow and cleanup
  - Test error conditions and structured error responses
  - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5, 9.6_

- [-] 10. Validate code quality and Go best practices
  - Ensure UTF-8 LF encoding for all source files
  - Verify proper Go project structure and package organization
  - Test concurrent session handling with proper mutex usage
  - Validate resource cleanup and connection management
  - Review interface separation and package boundaries
  - Confirm generated code separation and clear marking
  - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5, 10.6_