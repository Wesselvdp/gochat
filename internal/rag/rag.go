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
	"regexp"
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
	relevanceThreshold = 10
)

type Document struct {
	Text      string    // Original text
	Embedding []float32 // Vector embedding
	ID        int64
	fileID    string
}

type SearchResult struct {
	Text  string  // The text chunk
	Score float32 // Similarity score
}

// SplitText splits the input text into strings based on new lines or sentence-ending punctuation.
func SplitText(text string) []string {

	re := regexp.MustCompile(`[\.\?]\s+|[\.\?]$`)
	// Use the regex to split the text
	sentences := re.Split(text, -1)

	// Trim spaces from each resulting string and filter out empty results
	var result []string
	for _, sentence := range sentences {
		trimmed := strings.TrimSpace(sentence)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

type EmbeddingWithOriginal struct {
	openai.Embedding // Embedding is embedded, so all its fields are accessible directly
	Text             string
}

func CreateChunkDocuments(ctx context.Context, text string, fileID string) ([]Document, error) {

	texts := SplitText(text)
	docs := make([]Document, 0, len(texts))
	embeddings, err := ai.GetEmbeddings(ctx, texts)
	if err != nil {
		return nil, err
	}

	for i, embedding := range embeddings {
		originalText := texts[embedding.Index]

		doc := Document{
			Text:      originalText,
			Embedding: embedding.Embedding,
			ID:        int64(i + 1),
			fileID:    fileID,
		}
		docs = append(docs, doc)
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
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
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
		texts[i] = strings.ToValidUTF8(doc.Text, "")
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

	err = milvusClient.LoadCollection(ctx, collectionName, true)
	if err != nil {
		return nil, fmt.Errorf("failed to load collection: %w", err)
	}
	//sp, _ := entity.NewIndexFlatSearchParam()
	//sp, err := entity.NewIndexIvfFlatSearchParam(200)
	sp, err := entity.NewIndexHNSWSearchParam(74)
	if err != nil {
		return nil, fmt.Errorf("failed to create search parameters: %w", err)
	}

	vectors := []entity.Vector{
		entity.FloatVector(queryEmbedding),
	}

	cols := []string{"text"}

	// Now also retrieving chunk_index
	sr, err := milvusClient.Search(
		ctx,
		collectionName,
		[]string{conversationID},
		"",
		cols,
		vectors,
		"embedding",
		entity.L2,
		5,
		sp,
	)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	firstResult := sr[0]
	results := make([]SearchResult, sr[0].ResultCount)

	// This is a bit brittle and doesn't scale well if we add more cols to the search
	for i := range results {
		text, err := firstResult.Fields[0].GetAsString(i)
		if err != nil {
			fmt.Println("error", err.Error())
			continue
		}
		results[i].Text = text
		results[i].Score = firstResult.Scores[i]
	}

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

	docs, err := CreateChunkDocuments(ctx, extractedText, fileID)
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
		formattedContext.WriteString(fmt.Sprintf("%s\n", result.Text))
		formattedContext.WriteString(fmt.Sprintf("Relevance Score: %.2f\n", result.Score))
	}

	return formattedContext.String()
}

func GetDocumentsFromQuery(ctx context.Context, query string, conversationID string) (string, error) {
	queryEmbedding, err := ai.GetEmbeddings(ctx, []string{query})
	if err != nil {
		fmt.Println("err", err.Error())
		return "", err
	}

	searchResult, err := SearchSimilarChunks(ctx, queryEmbedding[0].Embedding, conversationID, 5)
	markdown := formatSearchResultsToMarkdown(searchResult)

	return markdown, err
}

func determineRAGWithContext(userQuery string, documentContext string) (bool, error) {
	prompt := RAGDeterminationPrompt(userQuery, documentContext)
	response, err := ai.SingleQuery(prompt)

	if err != nil {
		return false, err
	}
	return strings.Contains(response, "YES"), nil
}

// GetRaggedAnswer Returns a ragged answer if the LLM deems RAG is required, returns a normal answer if not.
func GetRaggedAnswer(ctx context.Context, messages []openai.ChatCompletionMessage, conversationID string) (string, error) {
	query := messages[len(messages)-1].Content
	documentContext, err := GetDocumentsFromQuery(ctx, query, conversationID)

	if err != nil {
		return "", err
	}

	// LLM will decide whether RAG is required, given the question and the document context
	useRAG, err := determineRAGWithContext(query, documentContext)
	if err != nil {
		return "", err
	}

	var response string
	if useRAG {
		prompt := RagPrompt2(documentContext, query)
		response, err = ai.SingleQuery(prompt)
	} else {
		response, err = ai.GetCompletion(messages)
	}

	return response, nil
}
