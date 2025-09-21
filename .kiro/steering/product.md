# Product Overview

Twinspeak is a drop-in replacement for Google's Gemini Live API that provides real-time conversational AI capabilities over WebSocket connections.

## Core Features

- WebSocket-based real-time communication at `/v1/speak` endpoint
- Gemini Live API compatibility for seamless integration
- Stateful sessions with setup, active conversation, and resumption support
- Multi-modal input/output: text and audio (base64-encoded)
- Function/tool calling capabilities
- Session resumption with time-limited handles

## Target Use Cases

- Real-time voice assistants and chatbots
- Interactive AI applications requiring low-latency responses
- Applications migrating from or testing against Gemini Live API
- Multi-modal AI experiences combining text and audio

## Architecture Philosophy

- Simple, focused implementation prioritizing compatibility
- Modular design with clear separation between API definitions, session management, and server logic
- Code generation from JSON Schema for type safety and consistency

## Testing Guidelines

- Never generate unit tests that rely on behavior testing with mocks or stubs
- Avoid change detector tests that break when implementation details change
- Focus on integration tests and functional tests that verify actual behavior
- Test real functionality rather than implementation patterns
- Keep tests simple and focused on verifying actual outcomes

## Code Style Guidelines

- Do not add comments that explain how code works
- Code should be self-documenting through clear naming and structure
- Only add TODO comments for future improvements
- Only add comments to document hacks or workarounds explaining WHY they were necessary
- Avoid redundant comments that restate what the code obviously does