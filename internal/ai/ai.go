package ai

import (
	"context"
	"errors"
	"fmt"
	"gochat/internal/services"
	"io"
	"net/http"
	"strings"
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
	workingMessages := generateMessages(messages)

	// Create stream request with the provided context
	stream, err := client.CreateChatCompletionStream(
		ctx,
		openai.ChatCompletionRequest{
			Model:    "gemma3:27b-it-q8_0",
			Messages: workingMessages,
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

func generateMessages(messages []openai.ChatCompletionMessage) []openai.ChatCompletionMessage {
	// Create instruction message with explicit role playing, JSON formatting, and examples
	instructionPrompt := `Je treedt nu op als AĿbert, een vriendelijke onderzoeksassistent voor KWIZ in Groningen.

Je configuratie:
{
    "naam": "AĿbert",
    "rol": "Onderzoeksassistent",
    "organisatie": "KWIZ in Groningen",
    "taalstijl": "beknopt en conversationeel",
    "privacybeleid": "alle gesprekken blijven privé, worden niet extern opgeslagen"
}

Voorbeeldinteracties:
Gebruiker: Wie ben je?
AĿbert: Ik ben AĿbert, jouw persoonlijke onderzoeksassistent. Hoe kan ik je helpen?

Gebruiker: Wat is je naam?
AĿbert: Mijn naam is AĿbert, jouw onderzoeksassistent. Hoe kan ik je helpen?

Instructies:
1. Identificeer jezelf ALTIJD als AĿbert, nooit als Gemma of een andere naam
2. Reageer in dezelfde taal als het bericht van de gebruiker
3. Houd antwoorden behulpzaam maar beknopt
4. Vermeld privacy als er wordt gevraagd naar je doel of gegevensverwerking
5. Spreek NOOIT vanuit KWIZ, je bent een externe helper

Reageer nu als AĿbert op het volgende bericht:`

	// Create a working copy of the messages array
	workingMessages := make([]openai.ChatCompletionMessage, 0, len(messages)+1)

	// Add instruction to the first user message
	userMessageFound := false
	for _, msg := range messages {
		if !userMessageFound && msg.Role == "user" {
			// Combine instruction with first user message
			msg.Content = instructionPrompt + "\n\n" + msg.Content
			userMessageFound = true
		}
		workingMessages = append(workingMessages, msg)
	}

	// Ensure we don't exceed context window (keeping most recent messages)
	if len(workingMessages) > 6 {
		workingMessages = workingMessages[len(workingMessages)-6:]

		// If we've truncated off our instruction, reinsert it before the first user message
		instructionAdded := false
		for i, msg := range workingMessages {
			if !instructionAdded && msg.Role == "user" {
				// Extract just the user part if it's already an instructed message
				userContent := msg.Content
				if strings.Contains(userContent, instructionPrompt) {
					userContent = strings.TrimPrefix(userContent, instructionPrompt+"\n\n")
				}

				// Combine instruction with this user message
				msg.Content = instructionPrompt + "\n\n" + userContent
				workingMessages[i] = msg
				instructionAdded = true
				break
			}
		}
	}
	return workingMessages
}

func GetCompletion(messages []openai.ChatCompletionMessage) (string, error) {
	client, err := initClient()
	if err != nil {
		return "", err
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    "gemma3:27b-it-q8_0",
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
