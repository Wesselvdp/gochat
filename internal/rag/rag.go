package rag

import (
	"context"
	"fmt"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/sashabaranov/go-openai"
	"gochat/internal/ai"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type TextExtractor func(*multipart.FileHeader) (string, error)

var extractors = map[string]TextExtractor{
	".txt": getTextFromText,
}

const (
	collectionName     = `documents`
	dim                = 1024
	relevanceThreshold = 90
)

type Document struct {
	Text      string    // Original text
	Embedding []float32 // Vector embedding
	ID        int64
	fileID    string
}

type SearchResult struct {
	Text       string  // The text chunk
	ChunkIndex int     // Position in original document
	Score      float32 // Similarity score
	Distance   float32 // Vector distance
}

func CreateChunkDocuments(text string, fileID string) ([]Document, error) {
	// Define chunk size and overlap (adjust as needed)
	chunkSize := 400
	overlap := 50
	//var chunks []string
	start := 0

	var docs []Document

	for start < len(text) {
		end := start + chunkSize
		if end > len(text) {
			end = len(text)
		}
		chunk := text[start:end]
		vector, err := ai.GetEmbedding(chunk)
		if err != nil {
			fmt.Println("GetEmbedding err:", err.Error())
			return nil, err
		}
		doc := Document{
			Text:      chunk,
			Embedding: vector,
			ID:        int64(chunkSize - overlap),
			fileID:    fileID,
		}
		docs = append(docs, doc)
		start += chunkSize - overlap
	}

	return docs, nil
}

func initMilvusClient(ctx context.Context) (client.Client, error) {
	milvusAddr := os.Getenv("MILVUS_ADDRESS")
	if milvusAddr == "" {
		milvusAddr = "standalone:19530"
	}

	milvusPw := os.Getenv("MILVUS_PW")
	if milvusPw == "" {
		return nil, fmt.Errorf("milvus password is empty")
	}

	milvusClient, err := client.NewClient(ctx, client.Config{
		Address:        milvusAddr,
		Username:       "root",
		Password:       milvusPw,
		DBName:         "",
		Identifier:     "",
		EnableTLSAuth:  false,
		APIKey:         "",
		ServerVersion:  "",
		DialOptions:    nil,
		RetryRateLimit: nil,
		DisableConn:    false,
	})

	if err != nil {
		// handling error and exit, to make example simple here
		fmt.Println("NewClient error:", err.Error())
		return nil, err
	}

	return milvusClient, nil
}

// SaveDocuments Saves new documents to the Vector DB's conversation partition
func SaveDocuments(ctx context.Context, docs []Document, fileID string, conversationID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	milvusClient, err := initMilvusClient(ctx)
	if err != nil {
		return err
	}
	// Prepare the data columns
	numDocs := len(docs)

	texts := make([]string, numDocs)
	ids := make([]string, numDocs)
	embeddings := make([][]float32, numDocs)

	// Split the data into columns
	for i, doc := range docs {
		texts[i] = doc.Text
		embeddings[i] = doc.Embedding
		ids[i] = fileID
	}

	partitionName := conversationID

	err = milvusClient.CreatePartition(ctx, collectionName, partitionName)
	if err != nil {
		return err
	}

	// Create column-based data
	textCol := entity.NewColumnVarChar("text", texts)
	fileIdCol := entity.NewColumnVarChar("fileId", ids)
	embeddingCol := entity.NewColumnFloatVector("embedding", dim, embeddings)

	// Insert data
	_, err = milvusClient.Insert(
		ctx,
		collectionName,
		partitionName,
		textCol,
		embeddingCol,
		fileIdCol,
	)

	if err != nil {
		fmt.Println("insert error:", err)
		return err
	}

	// Optional: Flush to make the data immediately searchable
	err = milvusClient.Flush(ctx, collectionName, false)
	if err != nil {
		return err
	}

	return nil
}

func RemoveDocumentsByFileId(ctx context.Context, fileID string, conversationID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	milvusClient, err := initMilvusClient(ctx)
	if err != nil {
		return err
	}

	expr := fmt.Sprintf("fileId == \"%s\"", fileID)

	err = milvusClient.Delete(ctx, collectionName, conversationID, expr)
	if err != nil {
		fmt.Println("Delete err:", err.Error())
		return err
	}

	return nil
}

func RemovePartition(ctx context.Context, conversationID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	milvusClient, err := initMilvusClient(ctx)
	if err != nil {
		return err
	}

	if err != nil {
		// handling error and exit, to make example simple here
		fmt.Println("NewClient error:", err.Error())
		return err
	}

	err = milvusClient.DropPartition(ctx, collectionName, conversationID)
	if err != nil {
		fmt.Println("Delete err:", err.Error())
		return err
	}

	return nil
}

func SearchSimilarChunks(
	ctx context.Context,
	queryEmbedding []float32,
	conversationID string,
	topK int64,
) ([]SearchResult, error) {
	milvusClient, err := initMilvusClient(ctx)
	if err != nil {
		return nil, err
	}

	err = milvusClient.LoadCollection(ctx, collectionName, false)
	if err != nil {
		return nil, fmt.Errorf("failed to load collection: %w", err)
	}

	sp, err := entity.NewIndexIvfFlatSearchParam(10)
	if err != nil {
		return nil, fmt.Errorf("failed to create search parameters: %w", err)
	}

	vectors := []entity.Vector{
		entity.FloatVector(queryEmbedding),
	}

	// Now also retrieving chunk_index
	searchResult, err := milvusClient.Search(
		ctx,
		collectionName,
		[]string{conversationID},
		"",
		[]string{"text"},
		vectors,
		"embedding",
		entity.L2,
		int(topK),
		sp,
	)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	results := make([]SearchResult, 0, topK)

	for i := 0; i < len(searchResult); i++ {
		current := searchResult[i]

		// Filter chunks based on relevance threshold
		// Note: For L2 distance, lower scores indicate higher similarity
		// So we want scores below the threshold
		if current.Scores[i] > relevanceThreshold {
			continue // Skip this chunk if it's not relevant enough
		}

		text, err := current.Fields.GetColumn("text").GetAsString(i)
		if err != nil {
			return nil, fmt.Errorf("failed to get text: %w", err)
		}

		results = append(results, SearchResult{
			Text:  text,
			Score: current.Scores[i],
		})
	}
	//
	return results, nil
}

func HandleFileEmbedding(ctx context.Context, file *multipart.FileHeader, fileID string, conversationID string) error {
	ext := filepath.Ext(file.Filename)

	extractor, exists := extractors[ext]
	if !exists {
		fmt.Errorf("unsupported file type: %s", ext)
	}

	extractedText, err := extractor(file)
	if err != nil {
		fmt.Println("extractor failed:", err.Error())
	}

	docs, err := CreateChunkDocuments(extractedText, fileID)
	err = SaveDocuments(ctx, docs, fileID, conversationID)

	if err != nil {
		return err
	}

	return nil
}

func getTextFromText(file *multipart.FileHeader) (string, error) {
	// Open the file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("error opening file: %v", err)
	}
	defer src.Close()

	// Use io.ReadAll for simpler reading
	content, err := io.ReadAll(src)
	if err != nil {
		return "", fmt.Errorf("error reading file: %v", err)
	}

	return string(content), nil
}

func formatSearchResultsToMarkdown(results []SearchResult) string {
	if len(results) == 0 {
		return ""
	}

	var formattedContext strings.Builder
	formattedContext.WriteString("Document Context:\n")

	for _, result := range results {
		formattedContext.WriteString("---\n")
		formattedContext.WriteString(fmt.Sprintf("%s\n", result))
		formattedContext.WriteString(fmt.Sprintf("Relevance Score: %.2f\n", result.Score))
		formattedContext.WriteString(fmt.Sprintf("Chunk Index: %d\n", result.ChunkIndex))
	}

	return formattedContext.String()
}

func GetDocumentsFromQuery(ctx context.Context, query string, conversationID string) (string, error) {
	queryEmbedding, err := ai.GetEmbedding(query)
	if err != nil {
		return "", err
	}
	searchResult, err := SearchSimilarChunks(ctx, queryEmbedding, conversationID, 3)
	markdown := formatSearchResultsToMarkdown(searchResult)
	return markdown, err
}

func DetermineRAG(userQuery string, tryAgain bool) (bool, error) {
	var additional string
	if tryAgain {
		additional = "Note: A similarity search in the uploaded document yielded no relevant results for the user's query."
	} else {
		additional = ""
	}

	prompt := fmt.Sprintf(`
You are an intelligent assistant. The user has uploaded a document. Your task is to determine if the user's query likely pertains to this document or if the answer can be derived from your general knowledge. 
%s
Respond with either "true" or "false" and explain your reasoning briefly:
- Respond "true" if you believe the query likely requires information from the uploaded document or other sources outside your training.
- Respond "false" only if you are very confident that the query can be fully answered with your existing knowledge and does not require the document.
- If the similarity search in the document yielded no results, consider this in your reasoning but do not assume it conclusively rules out the document's relevance.

User Query: %s
`, additional, userQuery)

	response, err := ai.SingleQuery(prompt)
	fmt.Println("toRag", response)
	if err != nil {
		return false, err
	}
	return strings.Contains(response, "true"), nil
}

func GetRaggedAnswer(ctx context.Context, messages []openai.ChatCompletionMessage, conversationID string) (string, error) {
	query := messages[len(messages)-1].Content
	documentContext, err := GetDocumentsFromQuery(ctx, query, conversationID)
	fmt.Println("documentContext", documentContext)
	if err != nil {
		return "No satisfactory answer was found in the document.", err
	}
	if documentContext == "" {
		isRagRequired, err := DetermineRAG(query, true)
		if err != nil {
			return "No satisfactory answer was found in the document.", err
		}
		if !isRagRequired {
			return ai.GetCompletion(messages)
		} else {
			return "No satisfactory answer was found in the document.", err

		}
	}

	systemPrompt := fmt.Sprintf(`# CONTEXT # 
I am a researcher. In the realm of society and government.

#########

# OBJECTIVE #
Your task is to help me efficiently go through data. This involves answering my questions with the help of provided data, always answer in a methodical and never make answers up, directly quote from the source when possible

#########

# STYLE #
Write in an informative and instructional style, resembling a research assistant.

#########

# Tone #
Maintain a positive and motivational tone throughout, It should feel like a friendly guide offering valuable insights.

# AUDIENCE #
The target audience is researchers looking to speed up their document analysis. Assume a readership that seeks practical advice and insights into the data they've provided you with'

#########

# RESPONSE FORMAT #
Provide a clear and consise answer where you quoote from the source if you have found a feasible answer. When you can't find a suitable answer to the researchers question, don't make things up but state that the provided data is insufficient for answering the question

#############

# START ANALYSIS #
If you understand, answer the user question given the provided data.

# RESEARCHER QUESTION #
%s

# RAG result #
%s
`, query, documentContext)

	response, err := ai.SingleQuery(systemPrompt)
	if err != nil {
		return "No satisfactory answer was found in the document.", err
	}

	return response, nil

}
