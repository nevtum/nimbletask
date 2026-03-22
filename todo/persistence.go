package todo

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// ErrInvalidMetadata is returned when metadata is missing or malformed
	ErrInvalidMetadata = errors.New("invalid or missing metadata")
	// ErrMissingID is returned when a todo has no ID
	ErrMissingID = errors.New("todo missing required ID")
	// ErrMissingCreated is returned when a todo has no created timestamp
	ErrMissingCreated = errors.New("todo missing required created timestamp")
	// ErrDuplicateID is returned when two todos have the same ID
	ErrDuplicateID = fmt.Errorf("duplicate ID")
	// ErrInvalidDateFormat is returned when a date cannot be parsed
	ErrInvalidDateFormat = errors.New("invalid date format")
	// ErrCycleDetected is returned when a cycle exists in the hierarchy
	ErrCycleDetected = errors.New("cycle detected in hierarchy")
)

// metadataRegex extracts key:value pairs from the HTML comment
var metadataRegex = regexp.MustCompile(`<!--\s*([^>]+?)\s*-->`)

// Save serializes the TodoList to a markdown file
func (tl *TodoList) Save(path string) error {
	f, err := pathToWriter(path)
	if err != nil {
		return fmt.Errorf("failed to establish writer: %w", err)
	}
	writer := bufio.NewWriter(f)

	if _, err := writer.WriteString(tl.serialize()); err != nil {
		return fmt.Errorf("failed to write to buffer: %w", err)
	}

	return writer.Flush()
}

// serialize converts all todos in the list into their markdown string representation
func (tl *TodoList) serialize() string {
	var sb strings.Builder

	for _, todo := range tl.roots {
		sb.WriteString(tl.serializeTodo(todo, 0))
	}

	return sb.String()
}

// serializeTodo converts a single Todo and all its children into markdown string representation
func (tl *TodoList) serializeTodo(todo *Todo, depth int) string {
	var sb strings.Builder

	// Build checkbox
	checkbox := "[ ]"
	if todo.Completed {
		checkbox = "[x]"
	}

	// Build indent
	indent := strings.Repeat("  ", depth)

	// Build metadata
	metaParts := []string{
		fmt.Sprintf("id:%s", todo.ID),
		fmt.Sprintf("parent:%s", todo.ParentID),
		fmt.Sprintf("created:%s", todo.CreatedAt.Format(time.RFC3339)),
	}

	if todo.Priority != 0 {
		metaParts = append(metaParts, fmt.Sprintf("priority:%d", todo.Priority))
	}
	if todo.Completed {
		metaParts = append(metaParts, fmt.Sprintf("completed:%s", todo.UpdatedAt.Format(time.RFC3339)))
	}
	if todo.DueDate != nil {
		metaParts = append(metaParts, fmt.Sprintf("due:%s", todo.DueDate.Format(time.RFC3339)))
	}
	if len(todo.Tags) > 0 {
		metaParts = append(metaParts, fmt.Sprintf("tags:%s", strings.Join(todo.Tags, ",")))
	}

	metadata := strings.Join(metaParts, "|")

	// Build the todo line
	line := fmt.Sprintf("%s- %s <!-- %s --> %s\n", indent, checkbox, metadata, todo.Title)
	sb.WriteString(line)

	// Write description if present
	if todo.Description != "" {
		lines := strings.Split(todo.Description, "\n")
		for _, lineText := range lines {
			if trimmed := strings.TrimSpace(lineText); trimmed == "" {
				continue // Skip empty lines in description
			}
			descLine := fmt.Sprintf("%s  %s\n", indent, lineText)
			sb.WriteString(descLine)
		}
	}

	// Recursively serialize children
	for _, child := range todo.Children {
		sb.WriteString(tl.serializeTodo(child, depth+1))
	}

	return sb.String()
}

// LoadTodoList loads a TodoList from a markdown file
func LoadTodoList(path string) (*TodoList, error) {
	file, err := pathToReader(path)
	if err != nil {
		// If file doesn't exist, return empty list
		return NewTodoList(), nil
	}

	tl := NewTodoList()
	todos := make(map[string]*Todo)
	var roots []*Todo

	// We need to parse the markdown list structure
	// Each todo is on a line starting with optional indent, then "- [ ]" or "- [x]"
	// Followed by HTML comment with metadata, then title
	// Subsequent indented lines (2 spaces per level) are description or children

	scanner := bufio.NewScanner(file)
	var stack []*Todo // Stack to track parent-child relationships based on indentation

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Check if line starts with a list item
		// Pattern: (indent)- [status] <!-- metadata --> Title
		// We need to extract indent level, checkbox status, metadata, and title

		// Find the list marker
		dashIdx := strings.Index(line, "-")
		if dashIdx == -1 {
			// Not a list item, could be description continuation
			// Handle description: append to current parent's description
			if len(stack) > 0 {
				current := stack[len(stack)-1]
				// Remove leading spaces from continuation line
				trimmed := strings.TrimPrefix(line, "  ")
				if current.Description != "" {
					current.Description += "\n" + trimmed
				} else {
					current.Description = trimmed
				}
			}
			continue
		}

		// Calculate indent level (2 spaces per level)
		indentStr := line[:dashIdx]
		indentLevel := len(indentStr) / 2

		// Extract checkbox status
		rest := line[dashIdx+1:]
		rest = strings.TrimSpace(rest)
		if len(rest) < 3 || (rest[:3] != "[ ]" && rest[:3] != "[x]") {
			return nil, fmt.Errorf("%w: invalid checkbox format", ErrInvalidMetadata)
		}
		completed := rest[:3] == "[x]"
		rest = rest[3:]

		// Find the HTML comment
		commentStart := strings.Index(rest, "<!--")
		commentEnd := strings.Index(rest, "-->")
		if commentStart == -1 || commentEnd == -1 || commentStart > commentEnd {
			return nil, fmt.Errorf("%w: malformed comment syntax", ErrInvalidMetadata)
		}

		metadataStr := strings.TrimSpace(rest[commentStart+4 : commentEnd])
		afterComment := strings.TrimSpace(rest[commentEnd+3:])

		// Parse metadata
		metaParts := strings.Split(metadataStr, "|")
		if len(metaParts) < 3 {
			return nil, fmt.Errorf("%w: malformed metadata format", ErrInvalidMetadata)
		}

		todo := &Todo{
			Children:  []*Todo{},
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
			Tags:      []string{},
		}

		// Parse each metadata key:value pair
		hasCreated := false
		for _, part := range metaParts {
			if part == "" {
				return nil, fmt.Errorf("%w: missing key-value separator", ErrInvalidMetadata)
			}
			kv := strings.SplitN(part, ":", 2)
			if len(kv) != 2 {
				return nil, fmt.Errorf("%w: malformed metadata format", ErrInvalidMetadata)
			}
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])

			switch key {
			case "id":
				if value == "" {
					return nil, ErrMissingID
				}
				todo.ID = value
			case "parent":
				todo.ParentID = value
			case "priority":
				p, err := strconv.Atoi(value)
				if err != nil {
					return nil, fmt.Errorf("%w: invalid priority", ErrInvalidMetadata)
				}
				todo.Priority = p
			case "created":
				hasCreated = true
				t, err := time.Parse(time.RFC3339, value)
				if err != nil {
					return nil, fmt.Errorf("%w: invalid created date", ErrInvalidDateFormat)
				}
				todo.CreatedAt = t
			case "completed":
				t, err := time.Parse(time.RFC3339, value)
				if err != nil {
					return nil, fmt.Errorf("%w: invalid completed date", ErrInvalidDateFormat)
				}
				todo.Completed = completed
				todo.UpdatedAt = t
			case "due":
				t, err := time.Parse(time.RFC3339, value)
				if err != nil {
					return nil, fmt.Errorf("%w: invalid due date", ErrInvalidDateFormat)
				}
				todo.DueDate = &t
			case "updated":
				t, err := time.Parse(time.RFC3339, value)
				if err != nil {
					return nil, fmt.Errorf("%w: invalid updated date", ErrInvalidDateFormat)
				}
				todo.UpdatedAt = t
			case "tags":
				if value != "" {
					todo.Tags = strings.Split(value, ",")
				}
			}
		}

		// Check that required 'created' field was present
		if !hasCreated {
			return nil, fmt.Errorf("%w: missing created", ErrMissingCreated)
		}

		// Validate required fields
		if todo.ID == "" {
			return nil, fmt.Errorf("%w: missing id", ErrMissingID)
		}

		// If not completed via metadata, set UpdatedAt to CreatedAt
		if todo.UpdatedAt.IsZero() {
			todo.UpdatedAt = todo.CreatedAt
		}

		// Title is the remaining text after the comment
		todo.Title = afterComment

		// Check for duplicate ID
		if _, exists := todos[todo.ID]; exists {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateID, todo.ID)
		}
		todos[todo.ID] = todo

		// Adjust stack based on indent level
		if indentLevel < len(stack) {
			// Going up or staying at same level
			stack = stack[:indentLevel]
		}

		// Determine parent
		if indentLevel == 0 {
			// Root todo
			roots = append(roots, todo)
		} else if len(stack) > 0 {
			// Child of the last item at previous indent
			parent := stack[len(stack)-1]
			todo.ParentID = parent.ID
			parent.Children = append(parent.Children, todo)
		} else {
			return nil, fmt.Errorf("%w: child without parent", ErrInvalidMetadata)
		}

		// Push this todo onto stack
		stack = append(stack, todo)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Build the TodoList
	tl.todos = todos
	tl.roots = roots

	// Validate the loaded structure
	if err := tl.Validate(); err != nil {
		return nil, err
	}

	return tl, nil
}

func pathToReader(path string) (io.Reader, error) {
	// If file doesn't exist, return empty list (per test behavior)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %w", err)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}

func pathToWriter(path string) (io.Writer, error) {
	dir := filepath.Dir(path)
	if dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	return f, nil
}
