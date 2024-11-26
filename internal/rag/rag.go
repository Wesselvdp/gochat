package rag

import (
	"context"
	"fmt"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"gochat/internal/ai"
	"io"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"
)

type TextExtractor func(*multipart.FileHeader) (string, error)

var extractors = map[string]TextExtractor{
	".txt": getTextFromText,
}

const (
	milvusAddr     = `localhost:19530`
	collectionName = `documents`
	dim            = 1024
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
	chunkSize := 200
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

func SaveDocuments(ctx context.Context, docs []Document, fileID string) error {
	milvusClient, err := client.NewClient(ctx, client.Config{
		Address: milvusAddr,
	})
	if err != nil {
		// handling error and exit, to make example simple here
		log.Fatal("failed to connect to milvus:", err.Error())
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

	partitionName := strings.ReplaceAll(fileID, "-", "_")

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

func SearchSimilarChunks(
	ctx context.Context,
	queryEmbedding []float32,
	fileIDs []string,
	topK int64,
) ([]SearchResult, error) {
	milvusClient, err := client.NewClient(ctx, client.Config{
		Address: milvusAddr,
	})
	if err != nil {
		// handling error and exit, to make example simple here
		log.Fatal("failed to connect to milvus:", err.Error())
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

	partitions := make([]string, len(fileIDs))
	for i, fileID := range fileIDs {
		partitions[i] = strings.ReplaceAll(fileID, "-", "_")
	}
	fmt.Printf("partitions: %v\n", partitions)

	// Now also retrieving chunk_index
	searchResult, err := milvusClient.Search(
		ctx,
		collectionName,
		partitions,
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

func HandleFileEmbedding(ctx context.Context, file *multipart.FileHeader, fileID string) error {
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
	err = SaveDocuments(ctx, docs, fileID)

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

func Query(ctx context.Context, query string, fileIDs []string) (string, error) {
	queryEmbedding, err := ai.GetEmbedding(query)
	if err != nil {
		return "", err
	}
	searchResult, err := SearchSimilarChunks(ctx, queryEmbedding, fileIDs, 3)
	markdown := formatSearchResultsToMarkdown(searchResult)
	return markdown, err
}
