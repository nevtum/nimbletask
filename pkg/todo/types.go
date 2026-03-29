package todo

import (
	"fmt"
	"strings"
	"time"
)

// Todo represents a single task in the hierarchical todo list
type Todo struct {
	ID          string
	Title       string
	Description string
	Completed   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ParentID    string
	Children    []*Todo
	Priority    int
	DueDate     *time.Time
	Tags        []string
}

// Serialize converts the Todo and all its children into markdown string representation
func (t *Todo) Serialize(depth int) string {
	var sb strings.Builder

	// Build checkbox
	checkbox := "[ ]"
	if t.Completed {
		checkbox = "[x]"
	}

	// Build indent
	indent := strings.Repeat("  ", depth)

	// Build metadata
	metaParts := []string{
		fmt.Sprintf("id:%s", t.ID),
		fmt.Sprintf("parent:%s", t.ParentID),
		fmt.Sprintf("created:%s", t.CreatedAt.Format(time.RFC3339)),
	}

	if t.Priority != 0 {
		metaParts = append(metaParts, fmt.Sprintf("priority:%d", t.Priority))
	}
	if t.Completed {
		metaParts = append(metaParts, fmt.Sprintf("completed:%s", t.UpdatedAt.Format(time.RFC3339)))
	}
	if t.DueDate != nil {
		metaParts = append(metaParts, fmt.Sprintf("due:%s", t.DueDate.Format(time.RFC3339)))
	}
	if len(t.Tags) > 0 {
		metaParts = append(metaParts, fmt.Sprintf("tags:%s", strings.Join(t.Tags, ",")))
	}

	metadata := strings.Join(metaParts, "|")

	// Build the todo line
	line := fmt.Sprintf("%s- %s <!-- %s --> %s\n", indent, checkbox, metadata, t.Title)
	sb.WriteString(line)

	// Write description if present
	if t.Description != "" {
		lines := strings.Split(t.Description, "\n")
		for _, lineText := range lines {
			if trimmed := strings.TrimSpace(lineText); trimmed == "" {
				continue // Skip empty lines in description
			}
			descLine := fmt.Sprintf("%s  %s\n", indent, lineText)
			sb.WriteString(descLine)
		}
	}

	// Recursively serialize children
	for _, child := range t.Children {
		childStr := child.Serialize(depth + 1)
		sb.WriteString(childStr)
	}

	return sb.String()
}

// TodoList manages a collection of todos with O(1) lookup
type TodoList struct {
	todos map[string]*Todo
	roots []*Todo
	clock Clock
	file  AbstractFile
}

// TodoUpdate represents update fields for a todo
type TodoUpdate struct {
	Title       *string
	Description *string
	Priority    *int
	DueDate     *time.Time
	Tags        []string
}

// Option is a function that modifies a TodoList
type Option func(*TodoList)

// NewTodoList creates a new empty todo list
func NewTodoList(file AbstractFile, options ...Option) *TodoList {
	tl := &TodoList{
		todos: make(map[string]*Todo),
		roots: []*Todo{},
		clock: RealClock{},
		file:  file,
	}
	for _, opt := range options {
		opt(tl)
	}
	return tl
}
