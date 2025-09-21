// Package model provides code generation coordination for API models.
package model

//go:generate go-jsonschema -p gemini -o ./gemini/models.gen.go ../../api/models/gemini/SetupRequest.json ../../api/models/gemini/ClientInputText.json ../../api/models/gemini/ClientInputAudio.json ../../api/models/gemini/ToolResult.json ../../api/models/gemini/SessionEnd.json ../../api/models/gemini/ServerOutputText.json ../../api/models/gemini/ServerOutputAudio.json ../../api/models/gemini/FunctionCall.json ../../api/models/gemini/SessionResumptionUpdate.json ../../api/models/gemini/Error.json
