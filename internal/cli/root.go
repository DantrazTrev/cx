package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"cheesebox/internal/storage"
	"cheesebox/internal/ui"
	"cheesebox/internal/search"
)

var db *storage.Storage

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cx",
	Short: "Cheesebox - Terminal-based note-taking with kanban boards and semantic search",
	Long: `Cheesebox (cx) is a beautiful terminal-based note-taking app with:
• Kanban boards for task management
• Semantic search powered by AI
• Fast CLI interface with gorgeous styling
• Apple Notes sync capability

Think of it as "Notion for the terminal" - powerful, fast, and beautiful.`,
	Run: showRecentNotes,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	// Initialize storage
	var err error
	db, err = storage.New()
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer db.Close()

	return rootCmd.Execute()
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(kanbanCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(embedCmd)
	rootCmd.AddCommand(syncCmd)
	
	// Global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
}

// showRecentNotes displays the most recent notes (default command)
func showRecentNotes(cmd *cobra.Command, args []string) {
	notes, err := db.GetRecentNotes(10)
	if err != nil {
		fmt.Printf("Error fetching recent notes: %v\n", err)
		os.Exit(1)
	}

	if len(notes) == 0 {
		fmt.Println("📝 No notes yet! Add your first note with: cx add \"Your note\"")
		return
	}

	fmt.Println(ui.RenderNotesList(notes, "Recent Notes"))
}

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:     "add [content]",
	Aliases: []string{"a"},
	Short:   "Add a new note",
	Long: `Add a new note to Cheesebox. Content can be provided as an argument 
or you'll be prompted to enter it interactively.

Examples:
  cx add "Fix authentication bug #urgent"
  cx a "Team meeting tomorrow #meeting"`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var content string
		
		if len(args) > 0 {
			content = args[0]
		} else {
			// Interactive mode - get content from user
			fmt.Print("Enter note content: ")
			var input string
			fmt.Scanln(&input)
			content = input
		}

		if content == "" {
			fmt.Println("❌ Note content cannot be empty")
			os.Exit(1)
		}

		// Extract tags from content
		tags := storage.ParseTags(content)
		
		// Default status
		status := "todo"
		
		note, err := db.AddNote(content, status, tags)
		if err != nil {
			fmt.Printf("❌ Error adding note: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Note added successfully! ID: %d\n", note.ID)
		if len(tags) > 0 {
			fmt.Printf("🏷️  Tags: %s\n", strings.Join(tags, ", "))
		}
	},
}

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:     "search [query]",
	Aliases: []string{"s", "se"},
	Short:   "Search notes",
	Long: `Search through your notes using text or semantic search.
If Ollama is available, semantic search will be used for better results.

Examples:
  cx search "authentication"
  cx s "meeting notes"
  cx se "bug fixes"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		
		// Try semantic search first, fall back to text search
		notes, err := searchNotes(query)
		if err != nil {
			fmt.Printf("❌ Error searching notes: %v\n", err)
			os.Exit(1)
		}

		if len(notes) == 0 {
			fmt.Printf("🔍 No notes found for: \"%s\"\n", query)
			return
		}

		fmt.Println(ui.RenderNotesList(notes, fmt.Sprintf("Search Results for \"%s\"", query)))
	},
}

// kanbanCmd represents the kanban command
var kanbanCmd = &cobra.Command{
	Use:     "kanban",
	Aliases: []string{"kb", "k"},
	Short:   "Open interactive kanban board",
	Long: `Open an interactive kanban board to manage your notes across
todo, doing, and done columns. Use arrow keys to navigate and 
enter to move notes between columns.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := ui.StartKanban(db); err != nil {
			fmt.Printf("❌ Error starting kanban: %v\n", err)
			os.Exit(1)
		}
	},
}

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit [id]",
	Short: "Edit a note by ID",
	Long: `Edit an existing note by providing its ID.
You can find note IDs using the list or search commands.

Examples:
  cx edit 123
  cx edit 42`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf("❌ Invalid note ID: %s\n", args[0])
			os.Exit(1)
		}

		note, err := db.GetNote(id)
		if err != nil {
			fmt.Printf("❌ Error fetching note: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Current content: %s\n", note.Content)
		fmt.Print("Enter new content: ")
		
		var newContent string
		fmt.Scanln(&newContent)
		
		if newContent == "" {
			fmt.Println("❌ Content cannot be empty")
			os.Exit(1)
		}

		// Extract new tags
		tags := storage.ParseTags(newContent)
		
		err = db.UpdateNote(id, newContent, note.Status, tags)
		if err != nil {
			fmt.Printf("❌ Error updating note: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Note %d updated successfully!\n", id)
	},
}

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:     "delete [id]",
	Aliases: []string{"del", "rm"},
	Short:   "Delete a note by ID",
	Long: `Delete a note by providing its ID.
Warning: This action cannot be undone!

Examples:
  cx delete 123
  cx del 42`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf("❌ Invalid note ID: %s\n", args[0])
			os.Exit(1)
		}

		// Confirm deletion
		fmt.Printf("Are you sure you want to delete note %d? (y/N): ", id)
		var confirm string
		fmt.Scanln(&confirm)
		
		if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
			fmt.Println("❌ Deletion cancelled")
			return
		}

		err = db.DeleteNote(id)
		if err != nil {
			fmt.Printf("❌ Error deleting note: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Note %d deleted successfully!\n", id)
	},
}

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls", "l"},
	Short:   "List all notes",
	Long: `List all notes with their IDs, status, and creation date.
Useful for finding note IDs for editing or deletion.`,
	Run: func(cmd *cobra.Command, args []string) {
		notes, err := db.GetRecentNotes(50) // Get more notes for listing
		if err != nil {
			fmt.Printf("❌ Error fetching notes: %v\n", err)
			os.Exit(1)
		}

		if len(notes) == 0 {
			fmt.Println("📝 No notes yet! Add your first note with: cx add \"Your note\"")
			return
		}

		fmt.Println(ui.RenderNotesList(notes, "All Notes"))
	},
}

// searchNotes performs search with fallback from semantic to text search
func searchNotes(query string) ([]*storage.Note, error) {
	return search.SearchWithFallback(db, query, 10)
}

// Helper function to format relative time
func formatTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("Jan 2, 2006")
	}
}

// embedCmd represents the embed command for generating embeddings
var embedCmd = &cobra.Command{
	Use:   "embed",
	Short: "Generate embeddings for semantic search",
	Long: `Generate embeddings for all notes to enable semantic search.
This requires Ollama to be running with the nomic-embed-text model.

Examples:
  cx embed              # Generate embeddings for all notes
  cx embed --note 123   # Generate embedding for specific note`,
	Run: func(cmd *cobra.Command, args []string) {
		noteID, _ := cmd.Flags().GetInt("note")
		
		client := search.NewOllamaClient("")
		if !client.IsAvailable() {
			fmt.Println("❌ Ollama is not available. Please ensure Ollama is running.")
			fmt.Println("💡 Install Ollama: https://ollama.ai")
			fmt.Println("💡 Run: ollama pull nomic-embed-text")
			os.Exit(1)
		}

		if noteID > 0 {
			// Generate embedding for specific note
			fmt.Printf("🧠 Generating embedding for note %d...\n", noteID)
			if err := client.GenerateEmbeddingForNote(db, noteID); err != nil {
				fmt.Printf("❌ Error generating embedding: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✅ Embedding generated for note %d\n", noteID)
		} else {
			// Generate embeddings for all notes
			fmt.Println("🧠 Generating embeddings for all notes...")
			fmt.Println("⏳ This may take a while...")
			if err := client.GenerateEmbeddingsForAllNotes(db); err != nil {
				fmt.Printf("❌ Error generating embeddings: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✅ Embeddings generation complete!")
		}
	},
}

// syncCmd represents the sync command for Apple Notes integration
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync with Apple Notes (placeholder)",
	Long: `Sync notes with Apple Notes. This feature is planned for future releases.
Currently shows a placeholder message.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🍎 Apple Notes Sync")
		fmt.Println("📋 This feature is coming soon!")
		fmt.Println("🚧 Current status: In development")
		fmt.Println("")
		fmt.Println("💡 Planned features:")
		fmt.Println("   • Import notes from Apple Notes")
		fmt.Println("   • Export Cheesebox notes to Apple Notes")
		fmt.Println("   • Two-way synchronization")
		fmt.Println("   • Conflict resolution")
		fmt.Println("")
		fmt.Println("⭐ Star the project on GitHub for updates!")
	},
}

func init() {
	// Add flags for embed command
	embedCmd.Flags().IntP("note", "n", 0, "Generate embedding for specific note ID")
}