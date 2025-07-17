package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Note represents a note in the system
type Note struct {
	ID        int       `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Status    string    `json:"status"` // "todo", "doing", "done"
	Tags      []string  `json:"tags"`
	Embedding []float64 `json:"embedding,omitempty"`
}

// Storage handles all database operations
type Storage struct {
	db *sql.DB
}

// New creates a new Storage instance
func New() (*Storage, error) {
	dbPath, err := getDBPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get database path: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	storage := &Storage{db: db}
	if err := storage.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return storage, nil
}

// Close closes the database connection
func (s *Storage) Close() error {
	return s.db.Close()
}

// AddNote adds a new note to the database
func (s *Storage) AddNote(content, status string, tags []string) (*Note, error) {
	if status == "" {
		status = "todo"
	}

	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tags: %w", err)
	}

	query := `
		INSERT INTO notes (content, status, tags, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`
	now := time.Now()
	result, err := s.db.Exec(query, content, status, string(tagsJSON), now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to insert note: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return &Note{
		ID:        int(id),
		Content:   content,
		Status:    status,
		Tags:      tags,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// GetNote retrieves a note by ID
func (s *Storage) GetNote(id int) (*Note, error) {
	query := `SELECT id, content, status, tags, created_at, updated_at FROM notes WHERE id = ?`
	row := s.db.QueryRow(query, id)

	var note Note
	var tagsJSON string
	err := row.Scan(&note.ID, &note.Content, &note.Status, &tagsJSON, &note.CreatedAt, &note.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("note with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to scan note: %w", err)
	}

	if err := json.Unmarshal([]byte(tagsJSON), &note.Tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
	}

	return &note, nil
}

// GetRecentNotes retrieves the most recent notes
func (s *Storage) GetRecentNotes(limit int) ([]*Note, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT id, content, status, tags, created_at, updated_at 
		FROM notes 
		ORDER BY updated_at DESC 
		LIMIT ?
	`
	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent notes: %w", err)
	}
	defer rows.Close()

	var notes []*Note
	for rows.Next() {
		var note Note
		var tagsJSON string
		err := rows.Scan(&note.ID, &note.Content, &note.Status, &tagsJSON, &note.CreatedAt, &note.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note row: %w", err)
		}

		if err := json.Unmarshal([]byte(tagsJSON), &note.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}

		notes = append(notes, &note)
	}

	return notes, nil
}

// SearchNotes performs a text-based search on notes
func (s *Storage) SearchNotes(query string) ([]*Note, error) {
	searchQuery := `
		SELECT id, content, status, tags, created_at, updated_at 
		FROM notes 
		WHERE content LIKE ? 
		ORDER BY updated_at DESC
	`
	
	rows, err := s.db.Query(searchQuery, "%"+query+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to search notes: %w", err)
	}
	defer rows.Close()

	var notes []*Note
	for rows.Next() {
		var note Note
		var tagsJSON string
		err := rows.Scan(&note.ID, &note.Content, &note.Status, &tagsJSON, &note.CreatedAt, &note.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note row: %w", err)
		}

		if err := json.Unmarshal([]byte(tagsJSON), &note.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}

		notes = append(notes, &note)
	}

	return notes, nil
}

// UpdateNote updates an existing note
func (s *Storage) UpdateNote(id int, content, status string, tags []string) error {
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	query := `
		UPDATE notes 
		SET content = ?, status = ?, tags = ?, updated_at = ?
		WHERE id = ?
	`
	_, err = s.db.Exec(query, content, status, string(tagsJSON), time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update note: %w", err)
	}

	return nil
}

// UpdateNoteStatus updates only the status of a note
func (s *Storage) UpdateNoteStatus(id int, status string) error {
	query := `UPDATE notes SET status = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.Exec(query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update note status: %w", err)
	}
	return nil
}

// DeleteNote deletes a note by ID
func (s *Storage) DeleteNote(id int) error {
	query := `DELETE FROM notes WHERE id = ?`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}
	return nil
}

// GetNotesByStatus retrieves notes by status for kanban board
func (s *Storage) GetNotesByStatus(status string) ([]*Note, error) {
	query := `
		SELECT id, content, status, tags, created_at, updated_at 
		FROM notes 
		WHERE status = ? 
		ORDER BY created_at ASC
	`
	
	rows, err := s.db.Query(query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to query notes by status: %w", err)
	}
	defer rows.Close()

	var notes []*Note
	for rows.Next() {
		var note Note
		var tagsJSON string
		err := rows.Scan(&note.ID, &note.Content, &note.Status, &tagsJSON, &note.CreatedAt, &note.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note row: %w", err)
		}

		if err := json.Unmarshal([]byte(tagsJSON), &note.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}

		notes = append(notes, &note)
	}

	return notes, nil
}

// SaveEmbedding saves an embedding for a note
func (s *Storage) SaveEmbedding(noteID int, embedding []float64) error {
	embeddingJSON, err := json.Marshal(embedding)
	if err != nil {
		return fmt.Errorf("failed to marshal embedding: %w", err)
	}

	query := `UPDATE notes SET embedding = ? WHERE id = ?`
	_, err = s.db.Exec(query, string(embeddingJSON), noteID)
	if err != nil {
		return fmt.Errorf("failed to save embedding: %w", err)
	}

	return nil
}

// GetNotesWithEmbeddings retrieves all notes that have embeddings
func (s *Storage) GetNotesWithEmbeddings() ([]*Note, error) {
	query := `
		SELECT id, content, status, tags, created_at, updated_at, embedding
		FROM notes 
		WHERE embedding IS NOT NULL AND embedding != ''
	`
	
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query notes with embeddings: %w", err)
	}
	defer rows.Close()

	var notes []*Note
	for rows.Next() {
		var note Note
		var tagsJSON, embeddingJSON string
		err := rows.Scan(&note.ID, &note.Content, &note.Status, &tagsJSON, &note.CreatedAt, &note.UpdatedAt, &embeddingJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note row: %w", err)
		}

		if err := json.Unmarshal([]byte(tagsJSON), &note.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}

		if embeddingJSON != "" {
			if err := json.Unmarshal([]byte(embeddingJSON), &note.Embedding); err != nil {
				return nil, fmt.Errorf("failed to unmarshal embedding: %w", err)
			}
		}

		notes = append(notes, &note)
	}

	return notes, nil
}

// migrate creates the necessary database tables
func (s *Storage) migrate() error {
	query := `
		CREATE TABLE IF NOT EXISTS notes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			content TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'todo',
			tags TEXT DEFAULT '[]',
			embedding TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_notes_status ON notes(status);
		CREATE INDEX IF NOT EXISTS idx_notes_updated_at ON notes(updated_at);
		CREATE INDEX IF NOT EXISTS idx_notes_content ON notes(content);
	`

	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

// getDBPath returns the path to the database file
func getDBPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".cheesebox", "cheesebox.db"), nil
}

// ParseTags extracts tags from content (words starting with #)
func ParseTags(content string) []string {
	words := strings.Fields(content)
	var tags []string
	for _, word := range words {
		if strings.HasPrefix(word, "#") && len(word) > 1 {
			tag := strings.TrimPrefix(word, "#")
			tag = strings.ToLower(tag)
			// Remove punctuation from end of tag
			tag = strings.TrimRight(tag, ".,!?;:")
			if tag != "" {
				tags = append(tags, tag)
			}
		}
	}
	return tags
}