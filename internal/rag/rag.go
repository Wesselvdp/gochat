package rag

import (
	"context"
	"fmt"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
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
	milvusAddr         = `milvus-standalone:19530:19530`
	collectionName     = `documents`
	dim                = 1024
	relevanceThreshold = 0.5
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
		milvusAddr = "milvus-standalone:19530"
	}

	milvusClient, err := client.NewClient(ctx, client.Config{
		Address:        milvusAddr,
		Username:       "",
		Password:       "",
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
	//fmt.Printf("%#v\n", searchResult)
	for i, sr := range searchResult {
		fmt.Println(sr.Fields.GetColumn("text").GetAsString(i))
		fmt.Println(sr.Scores)
	}

	results := make([]SearchResult, 0, topK)

	for i := 0; i < len(searchResult); i++ {
		current := searchResult[i]

		// Filter chunks based on relevance threshold
		// Note: For L2 distance, lower scores indicate higher similarity
		// So we want scores below the threshold
		if current.Scores[i] > relevanceThreshold {
			fmt.Println("OUT:", current.Scores[i])
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
		// Only include results above a certain threshold (e.g., 0.8)
		//if result.Score > 0.8 {
		formattedContext.WriteString("---\n")
		formattedContext.WriteString(fmt.Sprintf("%s\n", result.Text))
		formattedContext.WriteString(fmt.Sprintf("Relevance Score: %.2f\n", result.Score))
		formattedContext.WriteString(fmt.Sprintf("Chunk Index: %d\n", result.ChunkIndex))
		//}
	}

	return formattedContext.String()
}

func Query(ctx context.Context, query string, conversationID string) (string, error) {
	queryEmbedding, err := ai.GetEmbedding(query)
	if err != nil {
		return "", err
	}
	searchResult, err := SearchSimilarChunks(ctx, queryEmbedding, conversationID, 3)
	markdown := formatSearchResultsToMarkdown(searchResult)
	return markdown, err
}
