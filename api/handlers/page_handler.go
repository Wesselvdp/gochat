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
	"mime/multipart"
	"net/http"
	"time"
)

type Config struct {
	Router *gin.Engine
}

type UserRequestData struct {
	Messages []openai.ChatCompletionMessage `json:"messages"`
	Files    []string                       `json:"files"`
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

		// If files exist, do vector search and augment messages
		var augmentedMessages []openai.ChatCompletionMessage
		if len(data.Files) > 0 {
			fmt.Printf("FILES!")
			// Retrieve relevant document chunks based on files
			lastMessage := data.Messages[len(data.Messages)-1]
			documentContext, err := rag.Query(ctx, lastMessage.Content, data.Files)
			fmt.Println("documentContext", documentContext)
			if err != nil {
				// Log error but don't fail the request
				log.Printf("Document context retrieval error: %v", err)
			}

			// Create augmented messages with document context
			augmentedMessages = append(data.Messages[:len(data.Messages)-1], []openai.ChatCompletionMessage{
				{
					Role: "system",
					Content: `You are an AI assistant tasked with answering questions STRICTLY based on the provided document context.

				IMPORTANT RULES:
				- If document context is provided, you MUST use ONLY the information from that context to answer.
				- Do NOT use any external or general knowledge when context is present.
				- Directly quote from the context when possible.`,
				},
				{
					Role:    "system",
					Content: "If document context is provided below, use it to answer questions. If no context is provided or the context isn't relevant, respond based on your general knowledge.",
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
		//isNew := ctx.DefaultQuery("create", "f") == "true"

		isHTMX := ctx.GetHeader("HX-Request") != ""
		if isHTMX {
			// Serve partial HTML for HTMX requests
			render(ctx, http.StatusOK, views.Chat(id))
		} else {
			render(ctx, http.StatusOK, views.ChatPage(id))
		}
	}
}

func handlePDFFile(file *multipart.FileHeader) {
	// Add your PDF processing logic here
	fmt.Printf("Processing PDF file: %s\n", file.Filename)
}

func RagQueryHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		queryEmbedding, err := ai.GetEmbedding("Who is elara?")
		if err != nil {
			fmt.Printf("Error getting embedding: %v\n", err)
		}
		searchResult, err := rag.SearchSimilarChunks(c, queryEmbedding, []string{"d3656eab_1230_432d_8096_14829e1e801c"}, 2)
		if err != nil {
			fmt.Printf("Error getting embedding: %v\n", err)
		}
		fmt.Printf("Search result: %v\n", searchResult)
	}
}

func FileUploadHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
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
		err = rag.HandleFileEmbedding(c, file, fileID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("File %s uploaded successfully", file.Filename),
			"id":      fileID,
		})
	}

}

//curl https://cqzqss48xndt4b-11434.proxy.runpod.net/api/embeddings -d '{
//"model": "mxbai-embed-large",
//"prompt": "Llamas are members of the camelid family"
//}'
