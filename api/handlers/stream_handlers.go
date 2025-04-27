// handlers/chat_handlers.go
package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"gochat/internal/ai"
	"gochat/internal/rag"
	"gochat/internal/services"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ChatStreamHandler handles SSE connections for streaming chat responses
func ChatStreamHandler(manager *services.ClientManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		threadID := c.Query("thread_id")
		if threadID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing conversation_id parameter"})
			return
		}

		// Set headers for SSE
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

		// Add explicit CORS headers that Firefox may require
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Important: prevent Gin from using its buffer
		c.Writer.Flush()

		// Get or create client channel
		clientChan := manager.RegisterClient(threadID)

		// Send initial message with the same format// Initial dummy event (this ensures a clean "connected" response)
		c.Writer.Write([]byte("data: {\"content\":\"torgonestjolie\",\"isDone\":false}\n\n"))
		c.Writer.Flush()

		// Create a closed channel to detect client disconnect
		clientGone := c.Request.Context().Done()

		// Keep connection open until client disconnects
		for {
			select {
			case msg, ok := <-clientChan:
				if !ok {
					return
				}

				// Simply pass through the message as-is, assuming it's already formatted correctly
				// by SendRawEventToConversation or other senders
				_, err := c.Writer.Write([]byte(msg))
				if err != nil {
					fmt.Printf("Error writing to client: %v\n", err)
					manager.UnregisterClient(threadID)
					return
				}
				c.Writer.Flush()

			case <-clientGone:
				// Client disconnected
				manager.UnregisterClient(threadID)
				return
			}
		}
	}
}

type ChatRequest struct {
	ThreadID string               `json:"threadId"`
	Messages []ai.IncomingMessage `json:"messages"`
}

type ModelParams struct {
	Temperature float32 `json:"temperature"`
	TopP        float32 `json:"top_p"`
}

type Message struct {
		ID          string `json:"id"`
		Role        string `json:"role"`
		Content     string `json:"content"`
		ThreadID    string `json:"threadId"`
		CreatedAt   string `json:"createdAt"`
		Status      string `json:"status"`
		ModelParams ModelParams `json:"modelParams"`
		Attachments []struct {
			ID    string `json:"id"`
			Type  string `json:"type"`
			Name  string `json:"name"`
		} `json:"attachments"`
}

type MessageHandlerRequestData struct {
	Messages  []Message `json:"messages"`
	ThreadID string `json:"threadId"`
}

func processMessages(messages []Message, c *gin.Context) ([]ai.IncomingMessage, error) {
	var processedMessages []ai.IncomingMessage

	for _, msgData := range messages {
		// Create the message structure
		message := ai.IncomingMessage{
		Role:        msgData.Role,
		Content:     msgData.Content,
		ID:          msgData.ID,
		Attachments: make([]ai.Attachment, 0, len(msgData.Attachments)),
	}

		// Process each attachment for this message
		for _, attInfo := range msgData.Attachments {
		// Construct the attachment key
		attachmentKey := fmt.Sprintf("attachment_%s_%s", msgData.ID, attInfo.ID)

		// Get the file from form data
		file, header, err := c.Request.FormFile(attachmentKey)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Failed to get attachment: %s", attachmentKey),
			})
			return nil, err
		}

		defer file.Close()

		// Read into byte array
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to read attachment: %s", attachmentKey),
			})
			return nil, err
		}

		// Add to message
		message.Attachments = append(message.Attachments, ai.Attachment{
		ID:     attInfo.ID,
		Type:   attInfo.Type,
		Name:   header.Filename,
		Binary: fileBytes,
	})
	}

		processedMessages = append(processedMessages, message)
	}
	return processedMessages, nil

}

func ChatCompletionRequestBuilder() openai.ChatCompletionRequest {
	return openai.ChatCompletionRequest{
		Model:    "gemma3:27b-it-q8_0",
		Stream:   true,
	}
}

func GetModelParamsFromMessages(messages []Message) (*float32, *float32, error) {
	lastUserMessage := messages[len(messages)-2]
	if(lastUserMessage.Role != "user") {
		return nil, nil, errors.New("Last message is not from user")
	}
	return &lastUserMessage.ModelParams.Temperature, &lastUserMessage.ModelParams.TopP, nil

}

// MessageHandler handles incoming chat messages and triggers response streaming
func MessageHandler(manager *services.ClientManager) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Parse the multipart form (32MB limit or adjust as needed)
		if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form data"})
			return
		}

		// Get the messages data
		messagesDataStr := c.Request.FormValue("messagesData")
		if messagesDataStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing messages data"})
			return
		}

		// Parse the messagesData JSON
		var requestData MessageHandlerRequestData

		if err := json.Unmarshal([]byte(messagesDataStr), &requestData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid messages data format"})
			return
		}

		// Process each message and its attachments
		processedMessages, err := processMessages(requestData.Messages, c)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process messages"})
			return
		}

		var openAIMessages []openai.ChatCompletionMessage

		for _, m := range processedMessages {
			openAIMessages = append(openAIMessages, m.ToOpenAIMessage())
		}

		temperature, topP, err := GetModelParamsFromMessages(requestData.Messages)

		openaiRequest := ChatCompletionRequestBuilder()
		openaiRequest.Messages = openAIMessages
		if temperature != nil {
			openaiRequest.Temperature = *temperature
		}
		if topP != nil {
			openaiRequest.TopP = *topP
		}

		useRag := false
		for _, message := range processedMessages {
			if len(message.Attachments) > 0 {
				for _, attachment := range message.Attachments {
					if attachment.Type != "image" {
						// Check MIME type for PDF and TXT files
						mimeType := attachment.Type
						if mimeType == "application/pdf" ||
							mimeType == "text/plain" {
							useRag = true
							break
						}
					}
				}
				if useRag {
					break
				}
			}
		}

fmt.Println("useRag: ", useRag)
		if useRag {
			go rag.GetRaggedAnswerStream(c, openAIMessages, requestData.ThreadID, openaiRequest, manager)
		} else {
			go ai.GetCompletionStream(c, requestData.ThreadID, openAIMessages, openaiRequest, manager)
		}

		// Start async goroutine to stream LLM response
		//go services.StreamLLMResponse(data.ConversationID, manager)

		// Return success immediately - actual response will stream via SSE
		c.JSON(http.StatusAccepted, gin.H{"status": "Message received, response streaming"})
	}
}
