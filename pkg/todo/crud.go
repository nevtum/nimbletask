package todo

import (
	"errors"
	"fmt"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

// generateID creates a unique ID using NanoID
func generateID() string {
	id, err := gonanoid.New()
	if err != nil {
		// Panic if NanoID fails
		// This should be extremely rare
		panic(err)
	}
	return id
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
	return todo, nil
}

// Delete removes a todo from the list
// If force is false and the todo has children, returns an error
// If force is true, deletes the todo and promotes its children to root level
func (tl *TodoList) Delete(id string, force bool) error {
	todo, err := tl.Get(id)
	if err != nil {
		return err
	}

	// Check if todo has children
	if len(todo.Children) > 0 && !force {
		return errors.New("todo has children, use --force to delete anyway")
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

	// If force is true and todo has children, promote them to root level
	if force {
		for _, child := range todo.Children {
			child.ParentID = ""
			tl.roots = append(tl.roots, child)
			fmt.Printf("Child task %s promoted to root level\n", child.ID)
		}
	}

	// Remove from map
	delete(tl.todos, id)
	return nil
}

func (tl *TodoList) Complete(id string) error {
	todo, err := tl.Get(id)
	if err != nil {
		return err
	}

	todo.Complete(tl.clock.Now())
	return nil
}

// CompleteSubtree marks a todo and all its descendants as completed
func (tl *TodoList) CompleteSubtree(id string) error {
	todo, err := tl.Get(id)
	if err != nil {
		return err
	}

	todo.CompleteSubtree(tl.clock.Now())
	return nil
}

// GetRoots returns all root-level todos
func (tl *TodoList) GetRoots() []*Todo {
	return tl.roots
}

// GetChildren returns the children of a todo
func (tl *TodoList) GetChildren(parentID string) ([]*Todo, error) {
	parent, err := tl.Get(parentID)
	if err != nil {
		return nil, err
	}
	return parent.Children, nil
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
