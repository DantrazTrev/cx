package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"cheesebox/internal/storage"
)

// Color palette
var (
	// Primary colors
	primaryColor   = lipgloss.Color("#FF6B6B") // Coral red
	secondaryColor = lipgloss.Color("#4ECDC4") // Teal
	accentColor    = lipgloss.Color("#45B7D1") // Light blue
	
	// Status colors
	todoColor  = lipgloss.Color("#FFA726")   // Orange
	doingColor = lipgloss.Color("#66BB6A")   // Green
	doneColor  = lipgloss.Color("#9E9E9E")   // Gray
	
	// UI colors
	textColor      = lipgloss.Color("#2C3E50") // Dark blue-gray
	mutedColor     = lipgloss.Color("#7F8C8D") // Gray
	borderColor    = lipgloss.Color("#BDC3C7") // Light gray
	errorColor     = lipgloss.Color("#E74C3C") // Red
	successColor   = lipgloss.Color("#27AE60") // Green
	
	// Background colors
	bgColor        = lipgloss.Color("#FFFFFF") // White
	altBgColor     = lipgloss.Color("#F8F9FA") // Light gray
)

// Base styles
var (
	// Title styles
	titleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			MarginBottom(1)
	
	headerStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true).
			MarginBottom(1)
	
	// Text styles
	contentStyle = lipgloss.NewStyle().
			Foreground(textColor)
	
	mutedStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)
	
	// Status styles
	todoStyle = lipgloss.NewStyle().
			Foreground(todoColor).
			Bold(true)
	
	doingStyle = lipgloss.NewStyle().
			Foreground(doingColor).
			Bold(true)
	
	doneStyle = lipgloss.NewStyle().
			Foreground(doneColor).
			Bold(true)
	
	// UI element styles
	borderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1, 2)
	
	cardStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1, 2).
			MarginBottom(1)
	
	highlightStyle = lipgloss.NewStyle().
			Background(accentColor).
			Foreground(bgColor).
			Bold(true).
			Padding(0, 1)
	
	// Message styles
	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)
	
	successStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)
	
	warningStyle = lipgloss.NewStyle().
			Foreground(todoColor).
			Bold(true)
)

// RenderNotesList renders a formatted list of notes
func RenderNotesList(notes []*storage.Note, title string) string {
	var output strings.Builder
	
	// Title
	output.WriteString(titleStyle.Render("📋 " + title))
	output.WriteString("\n\n")
	
	// Notes
	for i, note := range notes {
		output.WriteString(renderNote(note, i == 0))
		if i < len(notes)-1 {
			output.WriteString("\n")
		}
	}
	
	// Footer
	output.WriteString("\n\n")
	output.WriteString(mutedStyle.Render(fmt.Sprintf("Total: %d notes", len(notes))))
	
	return output.String()
}

// renderNote renders a single note with beautiful formatting
func renderNote(note *storage.Note, isFirst bool) string {
	var output strings.Builder
	
	// Note header with ID and status
	header := fmt.Sprintf("#%d", note.ID)
	if note.Status != "" {
		header += " " + renderStatus(note.Status)
	}
	output.WriteString(headerStyle.Render(header))
	output.WriteString("\n")
	
	// Content
	content := note.Content
	if len(content) > 80 {
		content = content[:77] + "..."
	}
	output.WriteString(contentStyle.Render(content))
	output.WriteString("\n")
	
	// Metadata row
	var metadata []string
	
	// Time
	timeStr := formatTime(note.UpdatedAt)
	metadata = append(metadata, "⏰ "+timeStr)
	
	// Tags
	if len(note.Tags) > 0 {
		tagStr := "🏷️  " + strings.Join(note.Tags, ", ")
		metadata = append(metadata, tagStr)
	}
	
	if len(metadata) > 0 {
		output.WriteString(mutedStyle.Render(strings.Join(metadata, " • ")))
	}
	
	return cardStyle.Render(output.String())
}

// renderStatus renders a status badge with appropriate color
func renderStatus(status string) string {
	switch status {
	case "todo":
		return todoStyle.Render("TODO")
	case "doing":
		return doingStyle.Render("DOING")
	case "done":
		return doneStyle.Render("DONE")
	default:
		return mutedStyle.Render("UNKNOWN")
	}
}

// RenderKanbanBoard renders the kanban board layout
func RenderKanbanBoard(todoNotes, doingNotes, doneNotes []*storage.Note, selectedColumn int) string {
	var output strings.Builder
	
	// Title
	output.WriteString(titleStyle.Render("📊 Kanban Board"))
	output.WriteString("\n\n")
	
	// Column headers
	columns := []string{"📝 TODO", "⚡ DOING", "✅ DONE"}
	columnCounts := []int{len(todoNotes), len(doingNotes), len(doneNotes)}
	
	var headers []string
	for i, col := range columns {
		header := fmt.Sprintf("%s (%d)", col, columnCounts[i])
		if i == selectedColumn {
			header = highlightStyle.Render(header)
		} else {
			header = headerStyle.Render(header)
		}
		headers = append(headers, header)
	}
	
	output.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, headers...))
	output.WriteString("\n\n")
	
	// Render columns side by side
	todoCol := renderKanbanColumn(todoNotes, selectedColumn == 0)
	doingCol := renderKanbanColumn(doingNotes, selectedColumn == 1)
	doneCol := renderKanbanColumn(doneNotes, selectedColumn == 2)
	
	output.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, todoCol, doingCol, doneCol))
	
	// Instructions
	output.WriteString("\n\n")
	instructions := mutedStyle.Render("← → to navigate columns • ↑ ↓ to select notes • Enter to move • q to quit")
	output.WriteString(instructions)
	
	return output.String()
}

// renderKanbanColumn renders a single kanban column
func renderKanbanColumn(notes []*storage.Note, isSelected bool) string {
	const columnWidth = 30
	const columnHeight = 15
	
	var content strings.Builder
	
	for i, note := range notes {
		if i >= columnHeight-2 { // Leave space for "..." indicator
			content.WriteString(mutedStyle.Render("..."))
			break
		}
		
		// Truncate content to fit column
		noteContent := note.Content
		if len(noteContent) > columnWidth-4 {
			noteContent = noteContent[:columnWidth-7] + "..."
		}
		
		// Add note with ID
		content.WriteString(fmt.Sprintf("#%d %s\n", note.ID, noteContent))
	}
	
	// Fill remaining space
	for i := len(notes); i < columnHeight; i++ {
		content.WriteString("\n")
	}
	
	// Style the column
	style := borderStyle.Width(columnWidth).Height(columnHeight)
	if isSelected {
		style = style.BorderForeground(primaryColor)
	}
	
	return style.Render(content.String())
}

// Helper functions for consistent formatting

// formatTime formats a time.Time into a human-readable relative time
func formatTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1m ago"
		}
		return fmt.Sprintf("%dm ago", minutes)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1h ago"
		}
		return fmt.Sprintf("%dh ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1d ago"
		}
		return fmt.Sprintf("%dd ago", days)
	default:
		return t.Format("Jan 2")
	}
}

// RenderError renders an error message with styling
func RenderError(message string) string {
	return errorStyle.Render("❌ " + message)
}

// RenderSuccess renders a success message with styling
func RenderSuccess(message string) string {
	return successStyle.Render("✅ " + message)
}

// RenderWarning renders a warning message with styling
func RenderWarning(message string) string {
	return warningStyle.Render("⚠️  " + message)
}

// RenderInfo renders an info message with styling
func RenderInfo(message string) string {
	return mutedStyle.Render("ℹ️  " + message)
}

// RenderBanner renders the Cheesebox banner
func RenderBanner() string {
	banner := `
   _____ _                         _               
  / ____| |                       | |              
 | |    | |__   ___  ___  ___  ___| |__   _____  __
 | |    | '_ \ / _ \/ _ \/ __|/ _ \ '_ \ / _ \ \/ /
 | |____| | | |  __/  __/\__ \  __/ |_) | (_) >  < 
  \_____|_| |_|\___|\___||___/\___|_.__/ \___/_/\_\
`
	
	return titleStyle.Render(banner) + "\n" + 
		   mutedStyle.Render("Terminal-based notes with kanban boards & semantic search") + "\n"
}