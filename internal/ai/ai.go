package ai

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
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

type Attachment struct {
	ID     string `json:"id"`
	Type   string `json:"type"` // e.g., "image", "pdf"
	Name   string `json:"name"`
	Binary []byte `json:"binary"`
}

type IncomingMessage struct {
	Role        string       `json:"role"`
	Content     string       `json:"content"` // for multimodal support
	ID          string       `json:"id"`
	Attachments []Attachment `json:"attachments"`
	Temperature float32	`json:"temperature"`
}

func encodeToBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func (m IncomingMessage) ToOpenAIMessage() openai.ChatCompletionMessage {
	if m.Content == "" {
		m.Content = " "
	}
	// Handle different message types
	message := openai.ChatCompletionMessage{
		Role:         m.Role,

	}

	// Create base text content
	multiContent := []openai.ChatMessagePart{
		{
			Type: openai.ChatMessagePartTypeText,
			Text: m.Content,
		},
	}

	// Process attachments
	for _, att := range m.Attachments {
		fmt.Printf("Processing attachment: %s, Type: %s\n", att.Name, att.Type)
		switch att.Type {
		case "image/png", "image/jpeg", "image/gif":
			base64Data := encodeToBase64(att.Binary)
			imageURL := fmt.Sprintf("data:image/jpeg;base64,%s", base64Data)

			multiContent = append(multiContent, openai.ChatMessagePart{
				Type: openai.ChatMessagePartTypeImageURL,
				ImageURL: &openai.ChatMessageImageURL{
					URL: imageURL,
				},
			})
		}
	}

	message.MultiContent = multiContent
	return message
}


// GetCompletionStream handles streaming completions with empty message handling
func GetCompletionStream(ctx *gin.Context, threadID string, messages []openai.ChatCompletionMessage, openaiRequest openai.ChatCompletionRequest, manager *services.ClientManager) error {
	client, err := initClient()
	if err != nil {
		return fmt.Errorf("failed to initialize client: %w", err)
	}

	accountName, exists := ctx.Get("account_name")

	if !exists {
		fmt.Println("Account name not found in context")
		return fmt.Errorf("account name not found in context")
	}
	workingMessages := generateMessages(messages, accountName.(string) )
	openaiRequest.Messages = workingMessages

	stream, err := client.CreateChatCompletionStream(
		context.Background(),
		openaiRequest,
	)
	if err != nil {
		fmt.Printf("error creating stream: %v\n", err)
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
				fmt.Println("stream closed", err)
				// Send a completion message with finished flag
				finishedMsg := `{"content":"","isDone":true}`
				manager.SendRawEventToConversation(threadID, "message", finishedMsg)
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
				manager.SendRawEventToConversation(threadID, "message", jsonMsg)
			}
		}
	}
}

func createInstructionPrompt(accountName string) string {
	instructionPrompt := fmt.Sprintf(`Je treedt nu op als AĿbert, een vriendelijke onderzoeksassistent voor %s in Groningen.

Je configuratie:
{
    "naam": "AĿbert",
    "rol": "Onderzoeksassistent",
    "organisatie": "%s in Groningen",
    "taalstijl": "beknopt en conversationeel",
    "privacybeleid": "alle gesprekken blijven privé, worden niet extern opgeslagen"
}

Voorbeeldinteracties:
Gebruiker: Wie ben je?
AĿbert: Ik ben AĿbert, jouw persoonlijke onderzoeksassistent. Hoe kan ik je helpen?

Gebruiker: Wat is je naam?
AĿbert: Mijn naam is AĿbert, jouw onderzoeksassistent. Hoe kan ik je helpen?

Gebruiker: Wat is je doel?
AĿbert: Mijn doel is om je te helpen met al je vragen en informatiebehoeften.

Instructies:
1. je naam is AĿbert, gebruik nooit Gemma of een andere naam.
2. Reageer in dezelfde taal als het bericht van de gebruiker
3. Houd antwoorden behulpzaam maar beknopt
4. Vermeld privacy als er wordt gevraagd naar je doel of gegevensverwerking
5. Spreek NOOIT vanuit %s, je bent een externe helper


Reageer nu als AĿbert op het volgende bericht:`, accountName, accountName, accountName)

	return instructionPrompt
}

func generateMessages(messages []openai.ChatCompletionMessage, accountName string) []openai.ChatCompletionMessage {
	instructionPrompt := createInstructionPrompt(accountName)
	// Inject instruction into the first user text message
	userMessageFound := false
	for i, msg := range messages {
		if !userMessageFound && msg.Role == openai.ChatMessageRoleUser {
			for j, part := range msg.MultiContent {
				if part.Type == openai.ChatMessagePartTypeText {
					messages[i].MultiContent[j].Text = instructionPrompt + "\n\n" + part.Text
					userMessageFound = true
					break
				}
			}
		}
	}

	// Truncate to last 6 messages
	if len(messages) > 6 {
		messages = messages[len(messages)-6:]

		// Ensure instruction still present
		instructionAdded := false
		for i, msg := range messages {
			if msg.Role == openai.ChatMessageRoleUser {
				for j, part := range msg.MultiContent {
					if part.Type == openai.ChatMessagePartTypeText && !instructionAdded {
						if !strings.HasPrefix(part.Text, instructionPrompt) {
							messages[i].MultiContent[j].Text = instructionPrompt + "\n\n" + part.Text
						}
						instructionAdded = true
						break
					}
				}
			}
		}
	}

	return messages
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
func SingleQueryStream(ctx *gin.Context, threadID string, query string, openaiRequest openai.ChatCompletionRequest, manager *services.ClientManager) error {
	err := GetCompletionStream(ctx, threadID, []openai.ChatCompletionMessage{
		{
			Role:    "user",
			Content: query,
		},
	}, openaiRequest, manager)
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
