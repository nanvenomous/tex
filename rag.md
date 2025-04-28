# Simple RAG System for Code Understanding

This guide will help you set up a simple RAG (Retrieval-Augmented Generation) system using Docker, Golang, and Ollama to help understand large codebases.

## Project Structure

```
code-rag/
├── docker-compose.yml
├── go.mod
├── go.sum
├── cmd/
│   ├── index/
│   │   └── main.go
│   └── server/
│       └── main.go
├── internal/
│   ├── embedder/
│   │   └── embedder.go
│   ├── chunker/
│   │   └── chunker.go
│   ├── store/
│   │   └── vector_store.go
│   └── rag/
│       └── rag.go
└── Dockerfile
```

## Step 1: Set up the project

```bash
mkdir -p code-rag/cmd/{index,server} code-rag/internal/{embedder,chunker,store,rag}
cd code-rag
go mod init github.com/yourusername/code-rag
```

## Step 2: Create a Docker Compose file

```yaml
# docker-compose.yml
version: '3'

services:
  qdrant:
    image: qdrant/qdrant:latest
    ports:
      - "6333:6333"
      - "6334:6334"
    volumes:
      - qdrant_data:/qdrant/storage

  code-rag-server:
    build: .
    ports:
      - "8080:8080"
    environment:
      - QDRANT_URL=http://qdrant:6333
      - OLLAMA_URL=http://host.docker.internal:11434
    volumes:
      - ./:/app
      - ${CODE_PATH}:/code:ro

volumes:
  qdrant_data:
```

## Step 3: Create Dockerfile

```dockerfile
# Dockerfile
FROM golang:1.21-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /server ./cmd/server

EXPOSE 8080

CMD ["/server"]
```

## Step 4: Code Chunker

```go
// internal/chunker/chunker.go
package chunker

import (
	"bufio"
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type CodeChunk struct {
	Content     string
	FilePath    string
	StartLine   int
	EndLine     int
	Language    string
	Imports     []string
	Definitions []string
}

type Chunker struct {
	MinChunkSize int
	MaxChunkSize int
}

func NewChunker(minSize, maxSize int) *Chunker {
	return &Chunker{
		MinChunkSize: minSize,
		MaxChunkSize: maxSize,
	}
}

func (c *Chunker) ChunkDirectory(dirPath string) ([]CodeChunk, error) {
	var chunks []CodeChunk

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Skip binary files, hidden files, etc.
		if !isTextFile(path) || strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}

		fileChunks, err := c.ChunkFile(path)
		if err != nil {
			return err
		}

		chunks = append(chunks, fileChunks...)
		return nil
	})

	return chunks, err
}

func (c *Chunker) ChunkFile(filePath string) ([]CodeChunk, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	ext := filepath.Ext(filePath)
	language := getLanguageFromExt(ext)

	scanner := bufio.NewScanner(bytes.NewReader(content))
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	chunks := c.splitIntoChunks(lines, filePath, language)
	return chunks, nil
}

func (c *Chunker) splitIntoChunks(lines []string, filePath, language string) []CodeChunk {
	var chunks []CodeChunk
	var currentChunk []string
	var startLine int = 1
	
	imports := extractImports(lines, language)
	definitions := extractDefinitions(lines, language)

	for i, line := range lines {
		currentChunk = append(currentChunk, line)
		
		// Create a new chunk if we've reached the max size or found a logical boundary
		if len(currentChunk) >= c.MaxChunkSize || isChunkBoundary(line, language) {
			if len(currentChunk) >= c.MinChunkSize {
				chunks = append(chunks, CodeChunk{
					Content:     strings.Join(currentChunk, "\n"),
					FilePath:    filePath,
					StartLine:   startLine,
					EndLine:     startLine + len(currentChunk) - 1,
					Language:    language,
					Imports:     imports,
					Definitions: definitions,
				})
			}
			
			currentChunk = nil
			startLine = i + 2
		}
	}
	
	// Add the final chunk if it's not empty
	if len(currentChunk) >= c.MinChunkSize {
		chunks = append(chunks, CodeChunk{
			Content:     strings.Join(currentChunk, "\n"),
			FilePath:    filePath,
			StartLine:   startLine,
			EndLine:     startLine + len(currentChunk) - 1,
			Language:    language,
			Imports:     imports,
			Definitions: definitions,
		})
	}
	
	return chunks
}

func isChunkBoundary(line, language string) bool {
	// This is a simplified implementation that should be enhanced
	// for better boundary detection based on language
	trimmed := strings.TrimSpace(line)
	
	if language == "go" {
		return strings.HasPrefix(trimmed, "func ") || 
			   strings.HasPrefix(trimmed, "type ") ||
			   strings.HasPrefix(trimmed, "var ") ||
			   strings.HasPrefix(trimmed, "const ") ||
			   trimmed == "}"
	}
	
	// Generic chunk boundaries
	return len(trimmed) == 0 || trimmed == "}" || strings.HasPrefix(trimmed, "function ") 
}

func extractImports(lines []string, language string) []string {
	var imports []string
	
	if language == "go" {
		inImportBlock := false
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			
			if strings.HasPrefix(trimmed, "import (") {
				inImportBlock = true
				continue
			}
			
			if inImportBlock && trimmed == ")" {
				inImportBlock = false
				continue
			}
			
			if inImportBlock && len(trimmed) > 0 {
				imports = append(imports, trimmed)
			} else if strings.HasPrefix(trimmed, "import ") {
				imports = append(imports, strings.TrimPrefix(trimmed, "import "))
			}
		}
	}
	
	return imports
}

func extractDefinitions(lines []string, language string) []string {
	var definitions []string
	
	if language == "go" {
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			
			if strings.HasPrefix(trimmed, "func ") || 
			   strings.HasPrefix(trimmed, "type ") {
				definitions = append(definitions, trimmed)
			}
		}
	}
	
	return definitions
}

func isTextFile(path string) bool {
	// Simple check for text files by extension
	ext := strings.ToLower(filepath.Ext(path))
	textExtensions := map[string]bool{
		".go":   true,
		".py":   true,
		".js":   true,
		".ts":   true,
		".java": true,
		".c":    true,
		".cpp":  true,
		".h":    true,
		".hpp":  true,
		".rs":   true,
		".txt":  true,
		".md":   true,
	}
	
	return textExtensions[ext]
}

func getLanguageFromExt(ext string) string {
	extToLang := map[string]string{
		".go":   "go",
		".py":   "python",
		".js":   "javascript",
		".ts":   "typescript",
		".java": "java",
		".c":    "c",
		".cpp":  "cpp",
		".h":    "c",
		".hpp":  "cpp",
		".rs":   "rust",
	}
	
	lang, ok := extToLang[strings.ToLower(ext)]
	if !ok {
		return "text"
	}
	return lang
}
```

## Step 5: Embedder using Ollama

```go
// internal/embedder/embedder.go
package embedder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yourusername/code-rag/internal/chunker"
)

type Embedder struct {
	OllamaURL  string
	ModelName  string
	HttpClient *http.Client
}

type OllamaEmbedRequest struct {
	Model    string `json:"model"`
	Prompt   string `json:"prompt"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

type OllamaEmbedResponse struct {
	Embedding []float32 `json:"embedding"`
}

func NewEmbedder(ollamaURL, modelName string) *Embedder {
	return &Embedder{
		OllamaURL:  ollamaURL,
		ModelName:  modelName,
		HttpClient: &http.Client{},
	}
}

func (e *Embedder) GetEmbedding(text string) ([]float32, error) {
	reqBody := OllamaEmbedRequest{
		Model:  e.ModelName,
		Prompt: text,
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	
	resp, err := e.HttpClient.Post(
		fmt.Sprintf("%s/api/embeddings", e.OllamaURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama API error: %s", resp.Status)
	}
	
	var embedResp OllamaEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, err
	}
	
	return embedResp.Embedding, nil
}

func (e *Embedder) GetCodeChunkEmbedding(chunk chunker.CodeChunk) ([]float32, error) {
	// Build a prompt that captures the code's context
	prompt := fmt.Sprintf(
		"File: %s\nLanguage: %s\nLines: %d-%d\n\nCode:\n%s",
		chunk.FilePath,
		chunk.Language,
		chunk.StartLine,
		chunk.EndLine,
		chunk.Content,
	)
	
	return e.GetEmbedding(prompt)
}
```

## Step 6: Vector Store Interface

```go
// internal/store/vector_store.go
package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yourusername/code-rag/internal/chunker"
)

type VectorStore struct {
	QdrantURL   string
	CollectionName string
	HttpClient  *http.Client
}

type Point struct {
	ID        string              `json:"id"`
	Vector    []float32           `json:"vector"`
	Payload   map[string]interface{} `json:"payload"`
}

type SearchRequest struct {
	Vector     []float32 `json:"vector"`
	Limit      int       `json:"limit"`
	WithPayload bool      `json:"with_payload"`
}

type SearchResult struct {
	Points []struct {
		ID      string                 `json:"id"`
		Score   float64                `json:"score"`
		Payload map[string]interface{} `json:"payload"`
	} `json:"points"`
}

func NewVectorStore(qdrantURL, collectionName string) *VectorStore {
	return &VectorStore{
		QdrantURL:     qdrantURL,
		CollectionName: collectionName,
		HttpClient:    &http.Client{},
	}
}

func (vs *VectorStore) Initialize(dimension int) error {
	// Create collection if it doesn't exist
	createCollectionURL := fmt.Sprintf("%s/collections/%s", 
		vs.QdrantURL, vs.CollectionName)
	
	collectionConfig := map[string]interface{}{
		"vectors": map[string]interface{}{
			"size":     dimension,
			"distance": "Cosine",
		},
	}
	
	jsonData, err := json.Marshal(collectionConfig)
	if err != nil {
		return err
	}
	
	req, err := http.NewRequest("PUT", createCollectionURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	
	req.Header.Set("Content-Type", "application/json")
	resp, err := vs.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	// 200 OK or 409 Conflict (already exists) are both acceptable
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusConflict {
		return fmt.Errorf("failed to create collection: %s", resp.Status)
	}
	
	return nil
}

func (vs *VectorStore) StoreChunk(id string, vector []float32, chunk chunker.CodeChunk) error {
	upsertURL := fmt.Sprintf("%s/collections/%s/points", 
		vs.QdrantURL, vs.CollectionName)
	
	point := Point{
		ID:     id,
		Vector: vector,
		Payload: map[string]interface{}{
			"content":     chunk.Content,
			"file_path":   chunk.FilePath,
			"start_line":  chunk.StartLine,
			"end_line":    chunk.EndLine,
			"language":    chunk.Language,
			"imports":     chunk.Imports,
			"definitions": chunk.Definitions,
		},
	}
	
	requestBody := map[string]interface{}{
		"points": []Point{point},
	}
	
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}
	
	resp, err := vs.HttpClient.Post(upsertURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to store vector: %s", resp.Status)
	}
	
	return nil
}

func (vs *VectorStore) Search(vector []float32, limit int) ([]map[string]interface{}, error) {
	searchURL := fmt.Sprintf("%s/collections/%s/points/search", 
		vs.QdrantURL, vs.CollectionName)
	
	searchRequest := SearchRequest{
		Vector:      vector,
		Limit:       limit,
		WithPayload: true,
	}
	
	jsonData, err := json.Marshal(searchRequest)
	if err != nil {
		return nil, err
	}
	
	resp, err := vs.HttpClient.Post(searchURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed: %s", resp.Status)
	}
	
	var result SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	var hits []map[string]interface{}
	for _, point := range result.Points {
		hit := point.Payload
		hit["score"] = point.Score
		hits = append(hits, hit)
	}
	
	return hits, nil
}
```

## Step 7: RAG Service

```go
// internal/rag/rag.go
package rag

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yourusername/code-rag/internal/embedder"
	"github.com/yourusername/code-rag/internal/store"
)

type RagService struct {
	Embedder    *embedder.Embedder
	VectorStore *store.VectorStore
	OllamaURL   string
	ModelName   string
	HttpClient  *http.Client
}

type OllamaCompletionRequest struct {
	Model    string `json:"model"`
	Prompt   string `json:"prompt"`
	Stream   bool   `json:"stream"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

type OllamaCompletionResponse struct {
	Response string `json:"response"`
}

func NewRagService(
	embedder *embedder.Embedder,
	vectorStore *store.VectorStore,
	ollamaURL string,
	modelName string,
) *RagService {
	return &RagService{
		Embedder:    embedder,
		VectorStore: vectorStore,
		OllamaURL:   ollamaURL,
		ModelName:   modelName,
		HttpClient:  &http.Client{},
	}
}

func (rs *RagService) Query(query string, numResults int) (string, error) {
	// Get embedding for the query
	queryEmbedding, err := rs.Embedder.GetEmbedding(query)
	if err != nil {
		return "", fmt.Errorf("failed to get query embedding: %w", err)
	}
	
	// Search for similar code chunks
	results, err := rs.VectorStore.Search(queryEmbedding, numResults)
	if err != nil {
		return "", fmt.Errorf("search failed: %w", err)
	}
	
	// Build context from search results
	var contextBuilder bytes.Buffer
	contextBuilder.WriteString("Here are code snippets that might help answer your question:\n\n")
	
	for i, result := range results {
		contextBuilder.WriteString(fmt.Sprintf("Snippet %d (from %s, lines %v-%v):\n```%s\n%v\n```\n\n",
			i+1,
			result["file_path"],
			result["start_line"],
			result["end_line"],
			result["language"],
			result["content"],
		))
	}
	
	// Generate response using Ollama
	prompt := fmt.Sprintf(
		"You are an expert programmer helping to understand a codebase.\n\n"+
		"USER QUERY: %s\n\n"+
		"CONTEXT:\n%s\n\n"+
		"Based on the code snippets above, please answer the user's query. "+
		"Explain the relevant parts of the code and how they relate to the question. "+
		"If the provided context doesn't contain enough information, say so.",
		query,
		contextBuilder.String(),
	)
	
	reqBody := OllamaCompletionRequest{
		Model:  rs.ModelName,
		Prompt: prompt,
		Stream: false,
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	
	resp, err := rs.HttpClient.Post(
		fmt.Sprintf("%s/api/generate", rs.OllamaURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama API error: %s", resp.Status)
	}
	
	var completionResp OllamaCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&completionResp); err != nil {
		return "", err
	}
	
	return completionResp.Response, nil
}
```

## Step 8: Indexer Command

```go
// cmd/index/main.go
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/yourusername/code-rag/internal/chunker"
	"github.com/yourusername/code-rag/internal/embedder"
	"github.com/yourusername/code-rag/internal/store"
)

func main() {
	codeDir := flag.String("dir", "", "Directory containing code to index")
	ollamaURL := flag.String("ollama", "http://localhost:11434", "Ollama API URL")
	qdrantURL := flag.String("qdrant", "http://localhost:6333", "Qdrant API URL")
	collection := flag.String("collection", "code", "Qdrant collection name")
	embeddingModel := flag.String("model", "codellama", "Ollama model to use for embeddings")
	minChunkSize := flag.Int("min-chunk", 10, "Minimum chunk size (lines)")
	maxChunkSize := flag.Int("max-chunk", 100, "Maximum chunk size (lines)")
	
	flag.Parse()
	
	if *codeDir == "" {
		log.Fatal("Please specify a directory to index with -dir")
	}
	
	// Initialize components
	chunker := chunker.NewChunker(*minChunkSize, *maxChunkSize)
	embedder := embedder.NewEmbedder(*ollamaURL, *embeddingModel)
	
	// Test embedding to get dimension
	testEmbed, err := embedder.GetEmbedding("test")
	if err != nil {
		log.Fatalf("Failed to get test embedding: %v", err)
	}
	
	vectorStore := store.NewVectorStore(*qdrantURL, *collection)
	if err := vectorStore.Initialize(len(testEmbed)); err != nil {
		log.Fatalf("Failed to initialize vector store: %v", err)
	}
	
	// Chunk and index the code
	chunks, err := chunker.ChunkDirectory(*codeDir)
	if err != nil {
		log.Fatalf("Failed to chunk directory: %v", err)
	}
	
	log.Printf("Found %d code chunks to index", len(chunks))
	
	for i, chunk := range chunks {
		// Generate a stable ID for the chunk
		idHash := sha256.Sum256([]byte(fmt.Sprintf("%s:%d-%d", 
			chunk.FilePath, chunk.StartLine, chunk.EndLine)))
		id := hex.EncodeToString(idHash[:])
		
		// Get embedding
		embedding, err := embedder.GetCodeChunkEmbedding(chunk)
		if err != nil {
			log.Printf("Warning: Failed to get embedding for chunk %d: %v", i, err)
			continue
		}
		
		// Store in vector database
		if err := vectorStore.StoreChunk(id, embedding, chunk); err != nil {
			log.Printf("Warning: Failed to store chunk %d: %v", i, err)
			continue
		}
		
		if i > 0 && i%100 == 0 {
			log.Printf("Indexed %d/%d chunks...", i, len(chunks))
		}
	}
	
	log.Printf("Successfully indexed %d code chunks", len(chunks))
}
```

## Step 9: API Server

```go
// cmd/server/main.go
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/yourusername/code-rag/internal/embedder"
	"github.com/yourusername/code-rag/internal/rag"
	"github.com/yourusername/code-rag/internal/store"
)

type QueryRequest struct {
	Query      string `json:"query"`
	NumResults int    `json:"num_results"`
}

type QueryResponse struct {
	Answer string `json:"answer"`
}

func main() {
	// Get configuration from environment
	ollamaURL := getEnv("OLLAMA_URL", "http://localhost:11434")
	qdrantURL := getEnv("QDRANT_URL", "http://localhost:6333")
	collection := getEnv("COLLECTION_NAME", "code")
	embeddingModel := getEnv("EMBEDDING_MODEL", "codellama")
	llmModel := getEnv("LLM_MODEL", "codellama")
	
	// Initialize components
	embedder := embedder.NewEmbedder(ollamaURL, embeddingModel)
	vectorStore := store.NewVectorStore(qdrantURL, collection)
	ragService := rag.NewRagService(embedder, vectorStore, ollamaURL, llmModel)
	
	// Set up HTTP handlers
	http.HandleFunc("/api/query", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		var req QueryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		if req.Query == "" {
			http.Error(w, "Query cannot be empty", http.StatusBadRequest)
			return
		}
		
		if req.NumResults <= 0 {
			req.NumResults = 3 // Default to 3 results
		}
		
		answer, err := ragService.Query(req.Query, req.NumResults)
		if err != nil {
			log.Printf("Query error: %v", err)
			http.Error(w, "Failed to process query", http.StatusInternalServerError)
			return
		}
		
		resp := QueryResponse{
			Answer: answer,
		}
		
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("Response encoding error: %v", err)
		}
	})
	
	// Simple health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	// Start the server
	port := getEnv("PORT", "8080")
	log.Printf("Starting server on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	return value
}
```

## Step 10: Run the System

1. Build and start the containers:

```bash
# Set the path to your codebase
export CODE_PATH=/path/to/your/codebase
docker-compose up -d
```

2. Index your codebase:

```bash
docker exec -it code-rag-server go run cmd/index/main.go -dir /code
```

3. Query your codebase:

```bash
curl -X POST http://localhost:8080/api/query \
  -H "Content-Type: application/json" \
  -d '{"query": "How does the authentication system work?", "num_results": 5}'
```

## Step 11: Neovim Integration

Create a simple plugin or script to connect to your RAG API:

```lua
-- ~/.config/nvim/lua/coderag.lua
local api = vim.api
local M = {}

function M.query_code_rag()
  -- Prompt for query
  local query = vim.fn.input("Code question: ")
  if query == "" then return end
  
  -- Call API
  local curl_cmd = string.format(
    'curl -s -X POST "http://localhost:8080/api/query" ' ..
    '-H "Content-Type: application/json" ' ..
    '-d \'{"query": "%s", "num_results": 5}\'',
    query:gsub('"', '\\"')
  )
  
  local handle = io.popen(curl_cmd)
  local result = handle:read("*a")
  handle:close()
  
  -- Parse response
  local success, parsed = pcall(vim.json.decode, result)
  if not success then
    print("Error parsing response")
    return
  end
  
  -- Show in a floating window
  local buf = api.nvim_create_buf(false, true)
  api.nvim_buf_set_lines(buf, 0, -1, true, vim.split(parsed.answer, "\n"))
  
  local width = math.min(120, api.nvim_get_option("columns"))
  local height = math.min(30, api.nvim_get_option("lines"))
  
  local win = api.nvim_open_win(buf, true, {
    relative = "editor",
    width = width,
    height = height,
    col = (api.nvim_get_option("columns") - width) / 2,
    row = (api.nvim_get_option("lines") - height) / 2,
    style = "minimal",
    border = "rounded"
  })
  
  -- Close window with 'q'
  api.nvim_buf_set_keymap(buf, 'n', 'q', ':close<CR>', {noremap = true, silent = true})
end

-- Map to a keybinding
function M.setup()
  vim.keymap.set('n', '<leader>cr', M.query_code_rag, {noremap = true, desc = "Ask about code"})
end

return M
```

Add to your Neovim config:

```lua
-- init.lua or init.vim equivalent
require('coderag').setup()
```

## Step 12: CodeCompanion Integration

Since you're already using CodeCompanion with Neovim, you can create a simple shell script to integrate your RAG system with it:

```bash
#!/bin/bash
# ~/bin/code-rag-query.sh

query="$1"
if [ -z "$query" ]; then
  echo "Usage: code-rag-query.sh 'your question about the code'"
  exit 1
fi

curl -s -X POST "http://localhost:8080/api/query" \
  -H "Content-Type: application/json" \
  -d "{\"query\": \"$query\", \"num_results\": 5}" | jq -r '.answer'
```

Make it executable:

```bash
chmod +x ~/bin/code-rag-query.sh
```

Then you can configure CodeCompanion to use this script as a custom command.

## Optional Improvements

1. **Better Code Chunking**: Enhance the chunker to use AST parsing for more intelligent code splitting

2. **Metadata Extraction**: Extract more code metadata like function signatures, class hierarchies, etc.

3. **Caching**: Add response caching for better performance

4. **Multiple Models**: Support different models for different languages

5. **History Tracking**: Save query history for context-aware follow-up questions

6. **Context Window Management**: Limit context to fit within model's context window

7. **Web UI**: Add a simple web interface for browsing code and asking questions
