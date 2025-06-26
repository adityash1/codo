package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

func NewAgent(client *bedrockruntime.Client, getUserMessage func() (string, bool)) *Agent {
	return &Agent{
		client:         client,
		getUserMessage: getUserMessage,
	}
}

type Agent struct {
	client         *bedrockruntime.Client
	getUserMessage func() (string, bool)
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func NewUserMessage(content string) Message {
	return Message{Role: "user", Content: content}
}

func NewAssistantMessage(content string) Message {
	return Message{Role: "assistant", Content: content}
}

func (a *Agent) Run(ctx context.Context) error {
	conversation := []Message{}
	fmt.Println("Chat with Claude (use 'ctrl-c' to quit)")

	for {
		fmt.Print("\u001b[94mYou\u001b[0m: ")
		userInput, ok := a.getUserMessage()
		if !ok {
			break
		}

		userMessage := NewUserMessage(userInput)
		conversation = append(conversation, userMessage)

		response, err := a.runInference(ctx, conversation)
		if err != nil {
			return err
		}

		assistantMessage := NewAssistantMessage(response)
		conversation = append(conversation, assistantMessage)

		fmt.Printf("\u001b[93mClaude\u001b[0m: %s\n", response)
	}

	return nil
}

func (a *Agent) runInference(ctx context.Context, conversation []Message) (string, error) {
	reqBody := map[string]any{
		"anthropic_version": "bedrock-2023-05-31",
		"max_tokens":        1024,
		"messages":          conversation,
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := a.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(os.Getenv("BEDROCK_MODEL_ID")),
		ContentType: aws.String("application/json"),
		Body:        reqBodyBytes,
	})
	if err != nil {
		return "", err
	}

	var respBody map[string]any
	if err := json.Unmarshal(resp.Body, &respBody); err != nil {
		return "", err
	}

	// Extract the text content from the response
	if content, ok := respBody["content"].([]any); ok && len(content) > 0 {
		if textBlock, ok := content[0].(map[string]any); ok {
			if text, ok := textBlock["text"].(string); ok {
				return strings.TrimSpace(text), nil
			}
		}
	}

	return "", fmt.Errorf("failed to extract text from response")
}
