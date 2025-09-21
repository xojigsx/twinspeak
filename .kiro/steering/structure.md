# Project Structure

## Directory Organization

```
twinspeak/
├── api/gemini/                    # API specifications
│   ├── live.json                  # AsyncAPI 3.0 WebSocket spec
│   └── models/                    # JSON Schema definitions
│       ├── SetupRequest.json      # Initial session configuration
│       ├── ClientInputText.json   # Text input messages
│       ├── ClientInputAudio.json  # Audio input messages
│       ├── ToolResult.json        # Tool execution results
│       ├── SessionEnd.json        # Session termination
│       ├── ServerOutputText.json  # Text responses
│       ├── ServerOutputAudio.json # Audio responses
│       ├── FunctionCall.json      # Function/tool invocations
│       ├── SessionResumptionUpdate.json # Resumption tokens
│       └── Error.json             # Error responses
├── cmd/twinspeak/                 # CLI entry point
│   └── main.go                    # Cobra-based CLI with server command
├── pkg/                           # Internal packages
│   ├── model/                     # Generated Go types
│   │   ├── model.go               # Code generation coordination
│   │   └── gemini/                # Generated Gemini API types
│   │       └── models.gen.go      # Auto-generated from JSON schemas
│   └── session/                   # Session management
│       ├── session.go             # Session store and data structures
│       └── state.go               # Session state machine
├── srv/                           # HTTP server implementation
│   ├── srv.go                     # Chi router setup and health checks
│   └── ws.go                      # WebSocket handler and message processing
├── docs/                          # Documentation
│   └── spec.md                    # Detailed implementation specification
├── .kiro/steering/                # AI assistant guidance
├── Makefile                       # Build automation
├── go.mod                         # Go module definition
├── tools.go                       # Build-time tool dependencies
└── README.md                      # Quick start guide
```

## Package Conventions

### Import Aliases
- `g "jig.sx/twinspeak/pkg/model/gemini"` - Generated Gemini types

### Naming Patterns
- **Sessions**: Use UUID-based IDs with type `session.ID`
- **Message Types**: Use descriptive naming (e.g., `setup`, `input_text`, `output_audio`)
- **State Management**: Explicit state machine with `session.State` enum
- **WebSocket Messages**: JSON envelope pattern with `type` field for routing

### Code Organization
- **API-first**: JSON schemas drive Go type generation
- **Separation of concerns**: Clear boundaries between transport (srv), business logic (session), and data models (model)
- **Minimal dependencies**: Focus on essential libraries for WebSocket, routing, and CLI
- **Generated code**: Keep generated files separate and clearly marked