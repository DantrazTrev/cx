# 🧀 Cheesebox (`cx`)

> Terminal-based note-taking app with kanban boards and semantic search

Cheesebox is a powerful, beautiful CLI tool for developers who live in the terminal. Think of it as "Notion for the terminal" with gorgeous styling, semantic search, and interactive kanban boards.

## ✨ Features

- **🚀 Lightning Fast**: Sub-100ms response times for all operations
- **🎨 Beautiful UI**: Gorgeous terminal styling with Lip Gloss
- **🧠 Semantic Search**: AI-powered search using Ollama embeddings
- **📋 Kanban Boards**: Interactive terminal kanban interface
- **🏷️ Smart Tags**: Automatic tag extraction from content (#hashtags)
- **💾 Local First**: SQLite database, works offline
- **🍎 Apple Notes Sync**: Import/export with Apple Notes (coming soon)

## 🚀 Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/your-username/cheesebox.git
cd cheesebox

# Build and install
go build -o cx .
sudo mv cx /usr/local/bin/

# Or install directly with Go
go install github.com/your-username/cheesebox@latest
```

### Basic Usage

```bash
# Add your first note
cx add "Fix authentication bug #urgent #backend"

# View recent notes (default command)
cx

# Search notes
cx search "authentication"
cx s "bug fixes"

# Open interactive kanban board
cx kanban
cx kb

# List all notes with IDs
cx list

# Edit a note
cx edit 1

# Generate embeddings for semantic search
cx embed
```

## 📖 Commands

| Command | Aliases | Description |
|---------|---------|-------------|
| `cx` | | Show recent notes (default) |
| `cx add "content"` | `cx a` | Add a new note |
| `cx search "query"` | `cx s`, `cx se` | Search notes (semantic + text) |
| `cx kanban` | `cx kb`, `cx k` | Open interactive kanban board |
| `cx list` | `cx ls`, `cx l` | List all notes with IDs |
| `cx edit <id>` | | Edit note by ID |
| `cx delete <id>` | `cx del`, `cx rm` | Delete note by ID |
| `cx embed` | | Generate embeddings for semantic search |
| `cx sync` | | Sync with Apple Notes (coming soon) |

## 🎯 Kanban Board

The interactive kanban board lets you manage notes across three columns:

- **📝 TODO**: New tasks and ideas
- **⚡ DOING**: Work in progress  
- **✅ DONE**: Completed items

### Keybindings

- `←` `→` or `h` `l`: Navigate columns
- `↑` `↓` or `k` `j`: Select notes
- `Enter` or `Space`: Move note to next column
- `r`: Refresh data
- `q`: Quit

## 🧠 Semantic Search

Cheesebox uses Ollama for semantic search, allowing you to find notes by meaning rather than just keywords.

### Setup Ollama

1. **Install Ollama**: https://ollama.ai
2. **Pull the embedding model**:
   ```bash
   ollama pull nomic-embed-text
   ```
3. **Generate embeddings**:
   ```bash
   cx embed
   ```

### Examples

```bash
# Find notes about bugs, issues, problems, etc.
cx search "authentication problems"

# Find meeting-related notes
cx search "team discussions"

# Find code-related tasks
cx search "refactoring work"
```

If Ollama isn't available, search automatically falls back to text-based search.

## 🏷️ Tags

Cheesebox automatically extracts hashtags from your notes:

```bash
cx add "Review pull request #code-review #urgent"
cx add "Weekly planning meeting #meeting #planning"
cx add "Debug memory leak #bug #performance"
```

Tags are displayed in note listings and can be used for organization and search.

## 📁 Project Structure

```
cheesebox/
├── main.go                 # Entry point
├── internal/
│   ├── cli/               # Cobra commands
│   │   └── root.go
│   ├── storage/           # SQLite operations
│   │   └── storage.go
│   ├── ui/                # Bubble Tea interfaces
│   │   ├── kanban.go
│   │   └── styles.go
│   ├── search/            # Semantic search
│   │   └── ollama.go
│   └── sync/              # Apple Notes sync (coming soon)
├── go.mod
└── README.md
```

## 🎨 Design Philosophy

- **Speed First**: Every operation should feel instantaneous
- **Beautiful**: Terminal apps can be gorgeous with proper styling
- **Local**: Your data stays on your machine
- **Intuitive**: Commands should feel natural to developers
- **Extensible**: Easy to add new features and integrations

## 🔧 Configuration

Cheesebox stores data in `~/.cheesebox/`:

- `cheesebox.db`: SQLite database with your notes
- Configuration files (coming soon)

## 🛠️ Development

### Prerequisites

- Go 1.21+
- SQLite3
- Ollama (optional, for semantic search)

### Building

```bash
git clone https://github.com/your-username/cheesebox.git
cd cheesebox
go mod download
go build -o cx .
```

### Running Tests

```bash
go test ./...
```

## 🚧 Roadmap

- [x] Basic CLI with Cobra
- [x] SQLite storage layer  
- [x] Beautiful UI with Lip Gloss
- [x] Interactive kanban board
- [x] Semantic search with Ollama
- [x] Tag extraction and management
- [ ] Apple Notes sync
- [ ] Configuration file
- [ ] Note templates
- [ ] Backup and restore
- [ ] Export to markdown
- [ ] Plugin system
- [ ] Vim keybindings
- [ ] Multiple databases/workspaces

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Development Guidelines

1. Follow Go conventions (`gofmt`, proper error handling)
2. Use Charm libraries (Bubble Tea, Lip Gloss)
3. Optimize for terminal experience
4. Add helpful comments for complex logic
5. Handle errors gracefully with user-friendly messages

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Charm](https://charm.sh/) for the amazing CLI libraries
- [Ollama](https://ollama.ai/) for local AI embeddings
- [Cobra](https://cobra.dev/) for CLI framework
- The terminal tools community for inspiration

---

**Built with ❤️ for developers who love the terminal**

*Cheesebox - because your notes deserve to be as beautiful as your code* 🧀✨