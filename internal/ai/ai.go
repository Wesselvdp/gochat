package ai

import (
	"context"
	"errors"
	"fmt"
	"gochat/internal/services"
	"io"
	"net/http"
	"time"

	"github.com/sashabaranov/go-openai"
)

func initClient() (*openai.Client, error) {
	config := openai.ClientConfig{
		BaseURL: "http://5.22.250.243:11434/v1",
		HTTPClient: &http.Client{
			Timeout: time.Second * 30,
		},
	}

	client := openai.NewClientWithConfig(config)
	return client, nil
}

// GetCompletionStream handles streaming completions with empty message handling
func GetCompletionStream(ctx context.Context, conversationID string, messages []openai.ChatCompletionMessage, manager *services.ClientManager) error {
	client, err := initClient()
	if err != nil {
		return fmt.Errorf("failed to initialize client: %w", err)
	}

	// Create stream request with the provided context
	stream, err := client.CreateChatCompletionStream(
		ctx,
		openai.ChatCompletionRequest{
			Model:    "gemma3:27b",
			Messages: messages,
			Stream:   true,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create chat completion stream: %w", err)
	}
	defer stream.Close()

	// Process streaming responses
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("stream canceled: %w", ctx.Err())
		default:
			response, err := stream.Recv()

			if errors.Is(err, io.EOF) {
				fmt.Println("stream closed")
				// Send a completion message with finished flag
				finishedMsg := `{"content":"","isDone":true}`
				manager.SendRawEventToConversation(conversationID, "message", finishedMsg)
				// Stream finished naturally
				return nil
			}

			if err != nil && !errors.Is(err, openai.ErrTooManyEmptyStreamMessages) {
				return fmt.Errorf("error receiving from stream: %w", err)
			}

			// Process content if available
			if len(response.Choices) > 0 && response.Choices[0].Delta.Content != "" {
				content := response.Choices[0].Delta.Content
				//if content == "..." {
				//	return nil
				//}
				// Format as JSON with content and finished flag
				jsonMsg := fmt.Sprintf(`{"content":%q,"isDone":false}`, content)
				manager.SendRawEventToConversation(conversationID, "message", jsonMsg)
			}
		}
	}
}

func GetCompletion(messages []openai.ChatCompletionMessage) (string, error) {
	client, err := initClient()
	if err != nil {
		return "", err
	}
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    "gemma3:27b",
			Messages: messages,
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		fmt.Println("resp", resp)
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
func SingleQueryStream(ctx context.Context, convsersationId string, query string, manager *services.ClientManager) error {
	err := GetCompletionStream(ctx, convsersationId, []openai.ChatCompletionMessage{
		{
			Role:    "user",
			Content: query,
		},
	}, manager)
	if err != nil {
		fmt.Printf("Failed to get completion stream for singleQuery: %v\n", err)
		return err
	}
	return nil
}
func SingleQuery(query string) (string, error) {
	completion, err := GetCompletion([]openai.ChatCompletionMessage{
		{
			Role:    "user",
			Content: query,
		},
	})
	if err != nil {
		return "", err
	}
	return completion, nil
}

//type OllamaEmbeddingRequest struct {
//	Model  string `json:"model"`
//	Prompt string `json:"prompt"`
//}
//
//type OllamaEmbeddingResponse struct {
//	Embedding []float32 `json:"embedding"`
//}

//func GetEmbeddingOld(ctx context.Context, texts []string) ([]float32, error) {
//
//	reqBody := OllamaEmbeddingRequest{
//		Model:  "mxbai-embed-large",
//		Prompt: text,
//	}
//
//	jsonBody, err := json.Marshal(reqBody)
//	if err != nil {
//		return nil, err
//	}
//
//	resp, err := http.Post("https://"+podId+"-11434.proxy.runpod.net/api/embeddings",
//		"application/json",
//		bytes.NewBuffer(jsonBody))
//
//	if err != nil {
//		return nil, err
//	}
//	defer resp.Body.Close()
//
//	var result OllamaEmbeddingResponse
//	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
//		return nil, err
//	}
//
//	return result.Embedding, nil
//}

func GetEmbeddings(ctx context.Context, texts []string) ([]openai.Embedding, error) {
	client, err := initClient()
	if err != nil {
		return nil, err
	}

	request := openai.EmbeddingRequest{
		Input: texts,
		Model: "mxbai-embed-large",
	}

	response, err := client.CreateEmbeddings(ctx, request)
	if err != nil {
		return nil, err
	}
	return response.Data, nil
}
