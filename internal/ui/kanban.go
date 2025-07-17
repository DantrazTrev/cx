package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"cheesebox/internal/storage"
)

// KanbanModel represents the state of the kanban board
type KanbanModel struct {
	storage        *storage.Storage
	todoNotes      []*storage.Note
	doingNotes     []*storage.Note
	doneNotes      []*storage.Note
	selectedColumn int // 0 = todo, 1 = doing, 2 = done
	selectedNote   int // Index within the selected column
	width          int
	height         int
	quitting       bool
}

// StartKanban initializes and starts the kanban board interface
func StartKanban(storage *storage.Storage) error {
	model := &KanbanModel{
		storage:        storage,
		selectedColumn: 0,
		selectedNote:   0,
	}

	// Load initial data
	if err := model.loadNotes(); err != nil {
		return fmt.Errorf("failed to load notes: %w", err)
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// Init implements tea.Model
func (m *KanbanModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *KanbanModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "left", "h":
			if m.selectedColumn > 0 {
				m.selectedColumn--
				m.selectedNote = 0 // Reset note selection when changing columns
			}
			return m, nil

		case "right", "l":
			if m.selectedColumn < 2 {
				m.selectedColumn++
				m.selectedNote = 0 // Reset note selection when changing columns
			}
			return m, nil

		case "up", "k":
			notes := m.getNotesForColumn(m.selectedColumn)
			if len(notes) > 0 && m.selectedNote > 0 {
				m.selectedNote--
			}
			return m, nil

		case "down", "j":
			notes := m.getNotesForColumn(m.selectedColumn)
			if len(notes) > 0 && m.selectedNote < len(notes)-1 {
				m.selectedNote++
			}
			return m, nil

		case "enter", " ":
			return m, m.moveSelectedNote()

		case "r":
			// Refresh data
			return m, m.refresh()
		}
	}

	return m, nil
}

// View implements tea.Model
func (m *KanbanModel) View() string {
	if m.quitting {
		return "Thanks for using Cheesebox! 🧀\n"
	}

	return m.renderKanbanBoard()
}

// loadNotes loads notes from storage into the kanban columns
func (m *KanbanModel) loadNotes() error {
	var err error

	m.todoNotes, err = m.storage.GetNotesByStatus("todo")
	if err != nil {
		return err
	}

	m.doingNotes, err = m.storage.GetNotesByStatus("doing")
	if err != nil {
		return err
	}

	m.doneNotes, err = m.storage.GetNotesByStatus("done")
	if err != nil {
		return err
	}

	return nil
}

// getNotesForColumn returns the notes for a specific column
func (m *KanbanModel) getNotesForColumn(column int) []*storage.Note {
	switch column {
	case 0:
		return m.todoNotes
	case 1:
		return m.doingNotes
	case 2:
		return m.doneNotes
	default:
		return nil
	}
}

// getStatusForColumn returns the status string for a column
func (m *KanbanModel) getStatusForColumn(column int) string {
	switch column {
	case 0:
		return "todo"
	case 1:
		return "doing"
	case 2:
		return "done"
	default:
		return "todo"
	}
}

// moveSelectedNote moves the selected note to the next column
func (m *KanbanModel) moveSelectedNote() tea.Cmd {
	notes := m.getNotesForColumn(m.selectedColumn)
	if len(notes) == 0 || m.selectedNote >= len(notes) {
		return nil
	}

	selectedNote := notes[m.selectedNote]
	var newStatus string

	// Determine new status based on current column
	switch m.selectedColumn {
	case 0: // todo -> doing
		newStatus = "doing"
	case 1: // doing -> done
		newStatus = "done"
	case 2: // done -> todo (cycle back)
		newStatus = "todo"
	}

	return tea.Cmd(func() tea.Msg {
		err := m.storage.UpdateNoteStatus(selectedNote.ID, newStatus)
		if err != nil {
			return err
		}
		return refreshMsg{}
	})
}

// refresh reloads data from storage
func (m *KanbanModel) refresh() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		err := m.loadNotes()
		if err != nil {
			return err
		}
		return refreshMsg{}
	})
}

// refreshMsg is a custom message for refreshing the view
type refreshMsg struct{}

// renderKanbanBoard renders the kanban board with current state
func (m *KanbanModel) renderKanbanBoard() string {
	// Calculate column width based on terminal width
	columnWidth := 30
	if m.width > 0 {
		columnWidth = (m.width - 10) / 3 // Leave some margin
		if columnWidth < 25 {
			columnWidth = 25
		}
		if columnWidth > 40 {
			columnWidth = 40
		}
	}

	// Render title
	title := titleStyle.Render("📊 Cheesebox Kanban Board")
	
	// Render column headers
	headers := m.renderColumnHeaders()
	
	// Render columns
	columns := m.renderColumns(columnWidth)
	
	// Render instructions
	instructions := m.renderInstructions()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		headers,
		"",
		columns,
		"",
		instructions,
	)
}

// renderColumnHeaders renders the column headers with counts
func (m *KanbanModel) renderColumnHeaders() string {
	columns := []string{"📝 TODO", "⚡ DOING", "✅ DONE"}
	columnCounts := []int{len(m.todoNotes), len(m.doingNotes), len(m.doneNotes)}
	
	var headers []string
	for i, col := range columns {
		header := fmt.Sprintf("%s (%d)", col, columnCounts[i])
		
		style := headerStyle
		if i == m.selectedColumn {
			style = highlightStyle
		}
		
		headers = append(headers, style.Render(header))
	}
	
	return lipgloss.JoinHorizontal(lipgloss.Top, headers...)
}

// renderColumns renders all three kanban columns side by side
func (m *KanbanModel) renderColumns(columnWidth int) string {
	todoCol := m.renderColumn(m.todoNotes, 0, columnWidth)
	doingCol := m.renderColumn(m.doingNotes, 1, columnWidth)
	doneCol := m.renderColumn(m.doneNotes, 2, columnWidth)
	
	return lipgloss.JoinHorizontal(lipgloss.Top, todoCol, doingCol, doneCol)
}

// renderColumn renders a single kanban column
func (m *KanbanModel) renderColumn(notes []*storage.Note, columnIndex, width int) string {
	const maxHeight = 20
	
	var content []string
	
	for i, note := range notes {
		if i >= maxHeight-1 { // Leave space for "..." indicator
			content = append(content, mutedStyle.Render("..."))
			break
		}
		
		// Truncate content to fit column
		noteContent := note.Content
		maxContentWidth := width - 8 // Account for padding and ID
		if len(noteContent) > maxContentWidth {
			noteContent = noteContent[:maxContentWidth-3] + "..."
		}
		
		// Format note
		noteText := fmt.Sprintf("#%d %s", note.ID, noteContent)
		
		// Highlight selected note
		if columnIndex == m.selectedColumn && i == m.selectedNote {
			noteText = highlightStyle.Render(noteText)
		} else {
			noteText = contentStyle.Render(noteText)
		}
		
		content = append(content, noteText)
	}
	
	// Fill remaining space
	for len(content) < maxHeight {
		content = append(content, "")
	}
	
	// Join content
	columnContent := lipgloss.JoinVertical(lipgloss.Left, content...)
	
	// Style the column
	style := borderStyle.Width(width).Height(maxHeight + 2)
	if columnIndex == m.selectedColumn {
		style = style.BorderForeground(primaryColor)
	}
	
	return style.Render(columnContent)
}

// renderInstructions renders the control instructions
func (m *KanbanModel) renderInstructions() string {
	instructions := []string{
		"← → or h l: Navigate columns",
		"↑ ↓ or k j: Select notes",
		"Enter/Space: Move note",
		"r: Refresh",
		"q: Quit",
	}
	
	return mutedStyle.Render(lipgloss.JoinVertical(lipgloss.Left, instructions...))
}