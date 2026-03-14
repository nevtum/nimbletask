package todo

import (
	"errors"
	"time"
)

// generateID creates a simple unique ID (placeholder for NanoID)
func generateID() string {
	// Temporary simple ID generator
	// TODO: Replace with actual NanoID implementation
	return time.Now().Format("20060102150405") + randomString(6)
}

// randomString generates a random alphanumeric string of length n
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

// Add creates a new todo and adds it to the list
func (tl *TodoList) Add(title string, parentID string, position int) (*Todo, error) {
	if title == "" {
		return nil, errors.New("title cannot be empty")
	}

	// Validate parent exists if specified
	var parent *Todo
	if parentID != "" {
		var err error
		parent, err = tl.Get(parentID)
		if err != nil {
			return nil, err
		}
	}

	// Create the todo
	now := tl.clock.Now()
	todo := &Todo{
		ID:        generateID(),
		Title:     title,
		Completed: false,
		CreatedAt: now,
		UpdatedAt: now,
		ParentID:  parentID,
		Children:  []*Todo{},
		Priority:  0,
		Tags:      []string{},
	}

	// Add to the todos map
	tl.todos[todo.ID] = todo

	// Add to parent's children or roots
	if parent != nil {
		parent.Children = insertAtPosition(parent.Children, todo, position)
	} else {
		tl.roots = insertAtPosition(tl.roots, todo, position)
	}

	tl.modified = true
	return todo, nil
}

// Get retrieves a todo by ID
func (tl *TodoList) Get(id string) (*Todo, error) {
	todo, exists := tl.todos[id]
	if !exists {
		return nil, errors.New("todo not found")
	}
	return todo, nil
}

// Update modifies a todo's fields
func (tl *TodoList) Update(id string, updates TodoUpdate) (*Todo, error) {
	todo, err := tl.Get(id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if updates.Title != nil {
		todo.Title = *updates.Title
	}
	if updates.Description != nil {
		todo.Description = *updates.Description
	}
	if updates.Priority != nil {
		todo.Priority = *updates.Priority
	}
	if updates.DueDate != nil {
		todo.DueDate = updates.DueDate
	}
	if updates.Tags != nil {
		todo.Tags = updates.Tags
	}

	todo.UpdatedAt = tl.clock.Now()
	tl.modified = true
	return todo, nil
}

// Delete removes a todo from the list
func (tl *TodoList) Delete(id string) error {
	todo, err := tl.Get(id)
	if err != nil {
		return err
	}

	// Remove from parent's children or roots
	if todo.ParentID != "" {
		parent, _ := tl.Get(todo.ParentID)
		if parent != nil {
			parent.Children = removeFromSlice(parent.Children, todo)
		}
	} else {
		tl.roots = removeFromSlice(tl.roots, todo)
	}

	// Note: Children are not deleted, they become orphaned
	// This matches spec behavior - user can clean up separately

	// Remove from map
	delete(tl.todos, id)
	tl.modified = true
	return nil
}

// GetRoots returns all root-level todos
func (tl *TodoList) GetRoots() []*Todo {
	return tl.roots
}

// GetChildren returns the children of a todo
func (tl *TodoList) GetChildren(parentID string) []*Todo {
	parent, err := tl.Get(parentID)
	if err != nil {
		return []*Todo{}
	}
	return parent.Children
}

// IsModified returns whether the list has been modified
func (tl *TodoList) IsModified() bool {
	return tl.modified
}

// Helper functions

func insertAtPosition(slice []*Todo, item *Todo, position int) []*Todo {
	if position == -1 || position >= len(slice) {
		// Append
		return append(slice, item)
	}
	if position <= 0 {
		// Prepend
		return append([]*Todo{item}, slice...)
	}
	// Insert at position
	result := make([]*Todo, 0, len(slice)+1)
	result = append(result, slice[:position]...)
	result = append(result, item)
	result = append(result, slice[position:]...)
	return result
}

func removeFromSlice(slice []*Todo, item *Todo) []*Todo {
	result := make([]*Todo, 0, len(slice))
	for _, t := range slice {
		if t.ID != item.ID {
			result = append(result, t)
		}
	}
	return result
}
