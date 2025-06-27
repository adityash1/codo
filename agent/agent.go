package agent

import (
	"codo/tools"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

type Message struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type ContentBlock struct {
	Type  string          `json:"type"`
	Text  string          `json:"text,omitempty"`
	ID    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
}

type BedrockResponse struct {
	Content []ContentBlock `json:"content"`
}

type Agent struct {
	client         *bedrockruntime.Client
	getUserMessage func() (string, bool)
	tools          []tools.ToolDefinition
}

func NewAgent(client *bedrockruntime.Client, getUserMessage func() (string, bool), tools []tools.ToolDefinition) *Agent {
	return &Agent{
		client:         client,
		getUserMessage: getUserMessage,
		tools:          tools,
	}
}

func (a *Agent) Run(ctx context.Context) error {
	conversation := []Message{}
	fmt.Println("Chat with Claude (use 'ctrl-c' to quit)")

	readUserInput := true
	for {
		if readUserInput {
			fmt.Print("\u001b[94mYou\u001b[0m: ")
			userInput, ok := a.getUserMessage()
			if !ok {
				break
			}

			userMessage := NewUserMessage(userInput)
			conversation = append(conversation, userMessage)
		}

		response, err := a.runInference(ctx, conversation)
		if err != nil {
			return err
		}

		assistantMessage := NewAssistantMessage(response.Content)
		conversation = append(conversation, assistantMessage)

		toolResults := []map[string]any{}
		for _, content := range response.Content {
			switch content.Type {
			case "text":
				fmt.Printf("\u001b[93mClaude\u001b[0m: %s\n", content.Text)
			case "tool_use":
				result := a.executeTool(content.ID, content.Name, content.Input)
				toolResults = append(toolResults, result)
			}
		}
		if len(toolResults) == 0 {
			readUserInput = true
			continue
		}
		readUserInput = false
		conversation = append(conversation, Message{
			Role:    "user",
			Content: toolResults,
		})
	}

	return nil
}

func (a *Agent) runInference(ctx context.Context, conversation []Message) (*BedrockResponse, error) {
	bedrockTools := []map[string]any{}
	for _, tool := range a.tools {
		bedrockTools = append(bedrockTools, map[string]any{
			"name":         tool.Name,
			"description":  tool.Description,
			"input_schema": tool.InputSchema,
		})
	}

	reqBody := map[string]any{
		"anthropic_version": "bedrock-2023-05-31",
		"max_tokens":        1024,
		"messages":          conversation,
		"tools":             bedrockTools,
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(os.Getenv("BEDROCK_MODEL_ID")),
		ContentType: aws.String("application/json"),
		Body:        reqBodyBytes,
	})
	if err != nil {
		return nil, err
	}

	var respBody BedrockResponse
	if err := json.Unmarshal(resp.Body, &respBody); err != nil {
		return nil, err
	}

	return &respBody, nil
}

func (a *Agent) executeTool(id, name string, input json.RawMessage) map[string]any {
	var toolDef tools.ToolDefinition
	var found bool
	for _, tool := range a.tools {
		if tool.Name == name {
			toolDef = tool
			found = true
			break
		}
	}
	if !found {
		return map[string]any{
			"type":        "tool_result",
			"tool_use_id": id,
			"content":     "tool not found",
			"is_error":    true,
		}
	}

	fmt.Printf("\u001b[92mtool\u001b[0m: %s(%s)\n", name, input)
	response, err := toolDef.Function(input)
	if err != nil {
		return map[string]any{
			"type":        "tool_result",
			"tool_use_id": id,
			"content":     err.Error(),
			"is_error":    true,
		}
	}
	return map[string]any{
		"type":        "tool_result",
		"tool_use_id": id,
		"content":     response,
		"is_error":    false,
	}
}

func NewUserMessage(content string) Message {
	return Message{Role: "user", Content: content}
}

func NewAssistantMessage(content []ContentBlock) Message {
	return Message{Role: "assistant", Content: content}
}
