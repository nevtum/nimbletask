package todo

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	MIN_ID_LENGTH = 2
	MAX_ID_LENGTH = 12
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

// Complete marks the todo as completed and updates the UpdatedAt timestamp
func (t *Todo) Complete(now time.Time) {
	t.Completed = true
	t.UpdatedAt = now
}

// CompleteSubtree marks the todo and all its children as completed and updates the UpdatedAt timestamp
func (t *Todo) CompleteSubtree(now time.Time) {
	t.Complete(now)
	for _, child := range t.Children {
		child.CompleteSubtree(now)
	}
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
	// minIDLength int
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

// GetMinIDLength returns the minimum length required to uniquely identify tasks in the list
func (tl *TodoList) GetMinIDLength() int {
	sortedIDs := tl.sortedIDs()

	n := MIN_ID_LENGTH // Start n with a minimum length

	// Incrementally check for unique prefixes
	for n <= MAX_ID_LENGTH && hasCollisions(sortedIDs, n) {
		n++
	}

	return n
}

func (tl *TodoList) sortedIDs() []string {
	sortedIDs := make([]string, 0, len(tl.todos))
	for id := range tl.todos {
		sortedIDs = append(sortedIDs, id)
	}
	sort.Strings(sortedIDs)
	return sortedIDs
}

func hasCollisions(sortedIDs []string, n int) bool {
	for i := 0; i < len(sortedIDs)-1; i++ {
		if len(sortedIDs[i]) >= n && len(sortedIDs[i+1]) >= n {
			if sortedIDs[i][:n] == sortedIDs[i+1][:n] {
				return true // Collision found
			}
		}
	}
	return false // No collision
}
