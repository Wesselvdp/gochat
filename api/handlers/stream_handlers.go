// handlers/chat_handlers.go
package handlers

import (
	"fmt"
	"github.com/sashabaranov/go-openai"
	"gochat/internal/ai"
	"gochat/internal/rag"
	"gochat/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ChatStreamHandler handles SSE connections for streaming chat responses
func ChatStreamHandler(manager *services.ClientManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		conversationID := c.Query("conversation_id")
		if conversationID == "" {
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
		clientChan := manager.RegisterClient(conversationID)

		// Send initial message with the same format
		//initialEvent := fmt.Sprintf("event: message\ndata: Connected to event stream\n\n")
		//c.Writer.Write([]byte(initialEvent))
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
					manager.UnregisterClient(conversationID)
					return
				}
				c.Writer.Flush()

			case <-clientGone:
				// Client disconnected
				manager.UnregisterClient(conversationID)
				return
			}
		}
	}
}

// MessageHandler handles incoming chat messages and triggers response streaming
func MessageHandler(manager *services.ClientManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data struct {
			ConversationID string                         `json:"conversationId"`
			Messages       []openai.ChatCompletionMessage `json:"messages"`
			HasFiles       bool                           `json:"hasFiles"`
		}

		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request body"})
			return
		}
		if data.ConversationID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing conversation_id or message"})
			return
		}

		fmt.Println("hasFiles: ", data.HasFiles)

		if data.HasFiles {
			go rag.GetRaggedAnswerStream(c, data.Messages, data.ConversationID, manager)
		} else {
			go ai.GetCompletionStream(c, data.ConversationID, data.Messages, manager)
		}

		// Start async goroutine to stream LLM response
		//go services.StreamLLMResponse(data.ConversationID, manager)

		// Return success immediately - actual response will stream via SSE
		c.JSON(http.StatusAccepted, gin.H{"status": "Message received, response streaming"})
	}
}
