package tools

import (
	"encoding/json"

	"github.com/invopop/jsonschema"
)

type ToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
	Function    func(input json.RawMessage) (string, error)
}

func GenerateSchema[T any]() map[string]any {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T

	schema := reflector.Reflect(v)

	return map[string]any{
		"type":       "object",
		"properties": schema.Properties,
	}
}
