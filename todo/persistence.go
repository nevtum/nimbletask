package todo

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// metadataRegex extracts key:value pairs from the HTML comment
var metadataRegex = regexp.MustCompile(`<!--\s*([^>]+?)\s*-->`)

// Save serializes the TodoList to a markdown file
func (tl *TodoList) Save(file *File) error {
	return file.Save(tl.serialize())
}

func (tl *TodoList) Load(file *File) error {
	content, err := file.Load()
	if err != nil {
		if err == err.(FileDoesNotExist) {
			return nil
		} else {
			return err
		}
	}
	newTodos, err := deserialize(content)
	if err != nil {
		return err
	}
	*tl = *newTodos
	return nil
}

// serialize converts all todos in the list into their markdown string representation
func (tl *TodoList) serialize() string {
	var sb strings.Builder

	for _, todo := range tl.roots {
		sb.WriteString(todo.Serialize(0))
	}

	return sb.String()
}

// deserialize parses the markdown content and returns a TodoList
func deserialize(content string) (*TodoList, error) {
	tl := NewTodoList()
	todos := make(map[string]*Todo)
	var roots []*Todo

	// We need to parse the markdown list structure
	// Each todo is on a line starting with optional indent, then "- [ ]" or "- [x]"
	// Followed by HTML comment with metadata, then title
	// Subsequent indented lines (2 spaces per level) are description or children

	lines := strings.Split(content, "\n")
	var stack []*Todo // Stack to track parent-child relationships based on indentation

	for _, line := range lines {
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

	// Build the TodoList
	tl.todos = todos
	tl.roots = roots

	// Validate the loaded structure
	if err := tl.Validate(); err != nil {
		return nil, err
	}

	return tl, nil
}
