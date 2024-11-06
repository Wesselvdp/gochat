package ai

import (
	"context"
	"errors"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"net/http"
	"os"
	"time"
)

func GetCompletion(messages []openai.ChatCompletionMessage) (string, error) {

	podId := os.Getenv("RUNPOD_POD_ID")

	if len(podId) == 0 {
		return "", errors.New("RUNPOD_POD_ID not set")
	}

	config := openai.ClientConfig{
		BaseURL: "https://" + podId + "-11434.proxy.runpod.net/v1",
		HTTPClient: &http.Client{
			Timeout: time.Second * 30,
		},
	}

	client := openai.NewClientWithConfig(config)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    "gemma2:27b",
			Messages: messages,
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return "", err
	}

	// Check if there are any choices returned
	if len(resp.Choices) == 0 {
		return "", errors.New("MISSING CHOICES")
	}

	// Get the content from the first choice
	content := resp.Choices[0].Message.Content
	return content, nil
}
