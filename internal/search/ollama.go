package search

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
	
	"cheesebox/internal/storage"
)

// OllamaClient handles communication with Ollama API
type OllamaClient struct {
	baseURL    string
	httpClient *http.Client
	model      string
}

// EmbedRequest represents the request structure for Ollama embeddings
type EmbedRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// EmbedResponse represents the response structure from Ollama embeddings
type EmbedResponse struct {
	Embedding []float64 `json:"embedding"`
}

// SearchResult represents a search result with similarity score
type SearchResult struct {
	Note       *storage.Note
	Similarity float64
}

// NewOllamaClient creates a new Ollama client
func NewOllamaClient(baseURL string) *OllamaClient {
	if baseURL == "" {
		baseURL = "http://localhost:11434" // Default Ollama URL
	}

	return &OllamaClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		model: "nomic-embed-text", // Default embedding model
	}
}

// IsAvailable checks if Ollama is running and accessible
func (c *OllamaClient) IsAvailable() bool {
	resp, err := c.httpClient.Get(c.baseURL + "/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	
	return resp.StatusCode == http.StatusOK
}

// GetEmbedding generates an embedding for the given text
func (c *OllamaClient) GetEmbedding(text string) ([]float64, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	request := EmbedRequest{
		Model:  c.model,
		Prompt: text,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(
		c.baseURL+"/api/embeddings",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response EmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Embedding) == 0 {
		return nil, fmt.Errorf("received empty embedding")
	}

	return response.Embedding, nil
}

// SearchSemantic performs semantic search using embeddings
func (c *OllamaClient) SearchSemantic(storage *storage.Storage, query string, limit int) ([]*SearchResult, error) {
	// Get query embedding
	queryEmbedding, err := c.GetEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get query embedding: %w", err)
	}

	// Get all notes with embeddings
	notes, err := storage.GetNotesWithEmbeddings()
	if err != nil {
		return nil, fmt.Errorf("failed to get notes: %w", err)
	}

	// Calculate similarities
	var results []*SearchResult
	for _, note := range notes {
		if len(note.Embedding) == 0 {
			continue
		}

		similarity := cosineSimilarity(queryEmbedding, note.Embedding)
		
		// Only include results above threshold
		if similarity > 0.3 {
			results = append(results, &SearchResult{
				Note:       note,
				Similarity: similarity,
			})
		}
	}

	// Sort by similarity (highest first)
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Similarity < results[j].Similarity {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// Limit results
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// GenerateEmbeddingsForAllNotes generates embeddings for all notes that don't have them
func (c *OllamaClient) GenerateEmbeddingsForAllNotes(storage *storage.Storage) error {
	// Get all notes
	notes, err := storage.GetRecentNotes(1000) // Get a large number to cover all notes
	if err != nil {
		return fmt.Errorf("failed to get notes: %w", err)
	}

	successCount := 0
	errorCount := 0

	for _, note := range notes {
		// Skip if note already has embedding
		if len(note.Embedding) > 0 {
			continue
		}

		// Generate embedding
		embedding, err := c.GetEmbedding(note.Content)
		if err != nil {
			fmt.Printf("Failed to generate embedding for note %d: %v\n", note.ID, err)
			errorCount++
			continue
		}

		// Save embedding
		if err := storage.SaveEmbedding(note.ID, embedding); err != nil {
			fmt.Printf("Failed to save embedding for note %d: %v\n", note.ID, err)
			errorCount++
			continue
		}

		successCount++
		fmt.Printf("Generated embedding for note %d\n", note.ID)
		
		// Small delay to avoid overwhelming Ollama
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("Embedding generation complete: %d success, %d errors\n", successCount, errorCount)
	return nil
}

// GenerateEmbeddingForNote generates an embedding for a specific note
func (c *OllamaClient) GenerateEmbeddingForNote(storage *storage.Storage, noteID int) error {
	note, err := storage.GetNote(noteID)
	if err != nil {
		return fmt.Errorf("failed to get note: %w", err)
	}

	embedding, err := c.GetEmbedding(note.Content)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	if err := storage.SaveEmbedding(noteID, embedding); err != nil {
		return fmt.Errorf("failed to save embedding: %w", err)
	}

	return nil
}

// cosineSimilarity calculates the cosine similarity between two vectors
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// SearchWithFallback performs semantic search with fallback to text search
func SearchWithFallback(s *storage.Storage, query string, limit int) ([]*storage.Note, error) {
	client := NewOllamaClient("")
	
	// Try semantic search first
	if client.IsAvailable() {
		results, err := client.SearchSemantic(s, query, limit)
		if err == nil && len(results) > 0 {
			// Convert SearchResults to Notes
			var notes []*storage.Note
			for _, result := range results {
				notes = append(notes, result.Note)
			}
			return notes, nil
		}
	}

	// Fallback to text search
	return s.SearchNotes(query)
}