# Requirements Document

## Introduction

Twinspeak is a drop-in replacement for Google's Gemini Live API that provides real-time conversational AI capabilities over WebSocket connections. The system enables stateful, multi-modal communication sessions supporting text and audio input/output, function calling, and session resumption. The implementation follows a modular architecture with clear separation between API definitions, session management, and server logic, using code generation from JSON Schema for type safety and consistency.

## Requirements

### Requirement 1

**User Story:** As a developer, I want to initialize a Go module with proper dependencies, so that I can build a WebSocket-based conversational AI server.

#### Acceptance Criteria

1. WHEN initializing the project THEN the system SHALL create a go.mod file with module name "jig.sx/twinspeak"
2. WHEN setting up dependencies THEN the system SHALL include go-chi/chi/v5 for HTTP routing
3. WHEN setting up dependencies THEN the system SHALL include gobwas/ws for WebSocket implementation
4. WHEN setting up dependencies THEN the system SHALL include spf13/cobra for CLI framework
5. WHEN setting up dependencies THEN the system SHALL include google/uuid for session management
6. WHEN setting up build tools THEN the system SHALL include atombender/go-jsonschema for code generation
7. WHEN creating the tools.go file THEN the system SHALL properly configure build constraints for development tools

### Requirement 2

**User Story:** As an API designer, I want to define WebSocket message schemas using AsyncAPI 3.0 and JSON Schema, so that I can ensure type safety and API compatibility.

#### Acceptance Criteria

1. WHEN creating the AsyncAPI specification THEN the system SHALL define a WebSocket endpoint at "/v1/speak"
2. WHEN defining message types THEN the system SHALL support SetupRequest, ClientInputText, ClientInputAudio, ToolResult, and SessionEnd for client messages
3. WHEN defining message types THEN the system SHALL support ServerOutputText, ServerOutputAudio, FunctionCall, SessionResumptionUpdate, and Error for server messages
4. WHEN creating JSON schemas THEN the system SHALL use JSON Schema Draft 2020-12 format
5. WHEN defining SetupRequest THEN the system SHALL require type and model fields with optional sessionConfig
6. WHEN defining input messages THEN the system SHALL support both text and base64-encoded audio formats
7. WHEN defining audio messages THEN the system SHALL support wav, pcm16, and opus formats
8. WHEN defining tool messages THEN the system SHALL include name, callId, and result/arguments fields
9. WHEN defining error messages THEN the system SHALL include type, code, and message fields

### Requirement 3

**User Story:** As a developer, I want automated Go struct generation from JSON schemas, so that I can maintain type safety without manual code maintenance.

#### Acceptance Criteria

1. WHEN setting up code generation THEN the system SHALL create a pkg/model/model.go file with go:generate directives
2. WHEN running code generation THEN the system SHALL generate Go structs in pkg/model/gemini/models.gen.go
3. WHEN generating structs THEN the system SHALL use the gemini package name for generated types
4. WHEN generating structs THEN the system SHALL process all JSON schema files in api/models/gemini/
5. WHEN using generated types THEN the system SHALL support proper JSON marshaling and unmarshaling

### Requirement 4

**User Story:** As a system architect, I want in-memory session management with state tracking, so that I can maintain stateful WebSocket connections.

#### Acceptance Criteria

1. WHEN defining session states THEN the system SHALL support Connecting, Configured, Active, Closing, and Closed states
2. WHEN creating a session THEN the system SHALL generate a unique UUID-based session ID
3. WHEN managing sessions THEN the system SHALL track creation time, update time, and model information
4. WHEN storing session data THEN the system SHALL maintain a message log for each session
5. WHEN implementing session store THEN the system SHALL provide thread-safe Put, Get, and Delete operations
6. WHEN updating sessions THEN the system SHALL automatically update the timestamp on message append

### Requirement 5

**User Story:** As a server developer, I want a Chi-based HTTP server with WebSocket upgrade capability, so that I can handle real-time client connections.

#### Acceptance Criteria

1. WHEN setting up the server THEN the system SHALL use Chi router for HTTP handling
2. WHEN defining routes THEN the system SHALL provide a health check endpoint at "/healthz"
3. WHEN defining routes THEN the system SHALL provide WebSocket upgrade at "/v1/speak"
4. WHEN handling WebSocket connections THEN the system SHALL use gobwas/ws for high-performance WebSocket handling
5. WHEN processing messages THEN the system SHALL support JSON message envelope pattern with type field routing

### Requirement 6

**User Story:** As a WebSocket handler, I want proper message processing with session state management, so that I can maintain protocol compliance with Gemini Live API.

#### Acceptance Criteria

1. WHEN receiving the first message THEN the system SHALL require it to be a SetupRequest
2. WHEN processing SetupRequest THEN the system SHALL create a new session and transition to Configured state
3. WHEN session is configured THEN the system SHALL emit a SessionResumptionUpdate message
4. WHEN processing input_text messages THEN the system SHALL echo the text with "[echo]" prefix
5. WHEN processing input_audio messages THEN the system SHALL acknowledge with format information
6. WHEN processing tool_result messages THEN the system SHALL log the result for future correlation
7. WHEN processing end_session messages THEN the system SHALL transition to Closing state and send goodbye message
8. WHEN encountering errors THEN the system SHALL send properly formatted error messages
9. WHEN receiving unknown message types THEN the system SHALL respond with "unknown_type" error

### Requirement 7

**User Story:** As a system administrator, I want a Cobra-based CLI interface, so that I can easily configure and run the server.

#### Acceptance Criteria

1. WHEN creating the CLI THEN the system SHALL use Cobra framework for command structure
2. WHEN running the server THEN the system SHALL accept an --addr flag for listen address configuration
3. WHEN no address is specified THEN the system SHALL default to ":8080"
4. WHEN starting the server THEN the system SHALL log the listening address
5. WHEN the server starts THEN the system SHALL serve the Chi router handler

### Requirement 8

**User Story:** As a developer, I want build automation and documentation, so that I can efficiently develop and deploy the system.

#### Acceptance Criteria

1. WHEN setting up build automation THEN the system SHALL provide a Makefile with gen and run targets
2. WHEN running make gen THEN the system SHALL execute go generate for model generation
3. WHEN running make run THEN the system SHALL start the server on port 8080
4. WHEN creating documentation THEN the system SHALL provide a README.md with quick start instructions
5. WHEN documenting the API THEN the system SHALL reference the AsyncAPI specification location

### Requirement 9

**User Story:** As a client developer, I want a predictable WebSocket protocol flow, so that I can integrate with the Twinspeak API.

#### Acceptance Criteria

1. WHEN connecting to the WebSocket THEN the client SHALL send a setup message first
2. WHEN setup is complete THEN the client SHALL receive a session resumption token
3. WHEN sending text input THEN the client SHALL receive echoed text responses
4. WHEN sending audio input THEN the client SHALL receive acknowledgment messages
5. WHEN ending a session THEN the client SHALL send end_session and receive a goodbye message
6. WHEN errors occur THEN the client SHALL receive structured error messages with code and message fields

### Requirement 10

**User Story:** As a quality assurance engineer, I want the system to follow Go best practices and maintain code quality, so that the codebase is maintainable and reliable.

#### Acceptance Criteria

1. WHEN writing Go code THEN the system SHALL use UTF-8 LF encoding for all source files
2. WHEN structuring the project THEN the system SHALL follow standard Go project layout conventions
3. WHEN handling concurrency THEN the system SHALL use proper mutex locking for shared data structures
4. WHEN managing resources THEN the system SHALL properly close connections and clean up resources
5. WHEN implementing interfaces THEN the system SHALL use clear separation of concerns between packages
6. WHEN generating code THEN the system SHALL keep generated files separate and clearly marked