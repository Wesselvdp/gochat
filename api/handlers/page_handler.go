package handlers

import (
	"context"
	"fmt"
	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	openai "github.com/sashabaranov/go-openai"
	"gochat/internal/ai"
	"gochat/internal/auth"
	"gochat/internal/rag"
	"gochat/internal/services"
	views "gochat/views"
	"gochat/views/components"
	"log"
	"net/http"
	"time"
)

type Config struct {
	Router *gin.Engine
}

type UserRequestData struct {
	Messages       []openai.ChatCompletionMessage `json:"messages"`
	HasFiles       bool                           `json:"hasFiles"`
	ConversationID string                         `json:"conversationId"`
}

const appTimeout = time.Second * 10

func render(ctx *gin.Context, status int, template templ.Component) error {
	ctx.Status(status)
	return template.Render(ctx.Request.Context(), ctx.Writer)
}

func IndexPageHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), appTimeout)
		defer cancel()

		render(ctx, http.StatusOK, views.Index())
	}
}

func LoginPageHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), appTimeout)
		defer cancel()

		render(ctx, http.StatusOK, views.LoginPage())
	}
}

func LogoutPageHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		auth.UnsetTokenCookie(ctx)
		_, cancel := context.WithTimeout(context.Background(), appTimeout)
		defer cancel()

		render(ctx, http.StatusOK, views.LoginPage())
	}
}

func ComponentHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), appTimeout)
		defer cancel()

		componentName := ctx.Param("componentName")
		render(ctx, http.StatusOK, components.Component(componentName))
	}
}

func SendMessageHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), appTimeout)
		defer cancel()

		var data UserRequestData
		if err := ctx.ShouldBindJSON(&data); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userID, exists := ctx.Get("user")
		if exists {
			eventService := services.NewEventService(userID.(string))
			eventService.Create(services.EventMessage)
		}

		// If files exist, do vector search and augment messages
		var augmentedMessages []openai.ChatCompletionMessage
		if data.HasFiles {

			// Retrieve relevant document chunks based on files
			lastMessage := data.Messages[len(data.Messages)-1]
			documentContext, err := rag.Query(ctx, lastMessage.Content, data.ConversationID)

			if err != nil {
				// Log error but don't fail the request
				log.Printf("Document context retrieval error: %v", err)
			}

			// Create augmented messages with document context
			augmentedMessages = append(data.Messages[:len(data.Messages)-1], []openai.ChatCompletionMessage{
				{
					Role: "system",
					Content: `
					You are an AI assistant tasked with answering questions STRICTLY based on the provided document context.

					IMPORTANT RULES:
					- If document context is provided, you MUST use ONLY the information from that context to answer.
					- If no context is provided or the context isn't relevant, respond by saying that the answer to the question was not found in the document. Remember to answer in the user's language
					- Do NOT use any external or general knowledge when context is present.
					- Directly quote from the context when possible.
					- ALWAYS respond in the language of the user.
					
					DOCUMENT CONTEXT:
					` + documentContext,
				},

				{
					Role:    "system",
					Content: documentContext, // Retrieved document chunks
				},
				lastMessage,
			}...)

		} else {
			// No files, use original messages
			augmentedMessages = data.Messages
		}
		//for _, message := range augmentedMessages {
		//	fmt.Printf("%+v\n", message)
		//}
		aiResponse, err := ai.GetCompletion(augmentedMessages)

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"content": "Oeps, er is iets mis. We sturen er een ontwikkelaar op af"})
		} else {
			//fmt.Println(aiResponse)
			response := gin.H{
				"content": aiResponse,
				"data":    augmentedMessages,
			}
			ctx.JSON(http.StatusOK, response)

		}

	}
}

func ChatPageHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, cancel := context.WithTimeout(context.Background(), appTimeout)
		defer cancel()

		id := ctx.Param("id")

		isHTMX := ctx.GetHeader("HX-Request") != ""
		if isHTMX {
			// Serve partial HTML for HTMX requests
			render(ctx, http.StatusOK, views.Chat(id))
		} else {
			render(ctx, http.StatusOK, views.ChatPage(id))
		}
	}
}

func FileUploadHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get conversationId from form data
		conversationID := c.PostForm("conversationId")
		// Get file from form data
		file, err := c.FormFile("file")

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "No file uploaded",
			})
			return
		}
		// Save file entry locally
		fileService, err := services.NewFileService(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		dbEntry, err := fileService.Create(c, file.Filename)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		fileID := dbEntry.ID

		// Create embeddings and save to vector DB
		err = rag.HandleFileEmbedding(c, file, fileID, conversationID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("File %s uploaded successfully", file.Filename),
			"id":      fileID,
		})
	}

}
func FileDeleteHandler() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Get conversationId from form data
		conversationID := c.PostForm("conversationId")
		fileID := c.PostForm("fileId")

		// Get file from form data
		// Save file entry locally
		err := rag.RemoveDocumentsByFileId(c, fileID, conversationID)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("File %s successfully deleted", fileID),
		})
	}

}

func PartitionDeleteHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get conversationId from form data
		conversationID := c.PostForm("conversationId")

		// Get file from form data
		// Save file entry locally
		err := rag.RemovePartition(c, conversationID)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("Partition %s successfully deleted", conversationID),
		})
	}

}
