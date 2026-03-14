package todo

import (
	"errors"
)

// Move moves a todo to a new parent with the specified position
// Position: 0 = first, -1 = append, otherwise specific index
func (tl *TodoList) Move(id string, newParentID string, position int) error {
	// Get the todo to move
	todo, err := tl.Get(id)
	if err != nil {
		return err
	}

	// Validate new parent exists if specified
	if newParentID != "" {
		_, err = tl.Get(newParentID)
		if err != nil {
			return err
		}
	}

	// Check for cycles - cannot move under itself or any of its descendants
	if newParentID != "" {
		if err := tl.checkCycle(id, newParentID); err != nil {
			return err
		}
	}

	// Remove from current parent
	if todo.ParentID != "" {
		parent, _ := tl.Get(todo.ParentID)
		if parent != nil {
			parent.Children = removeFromSlice(parent.Children, todo)
		}
	} else {
		tl.roots = removeFromSlice(tl.roots, todo)
	}

	// Update parent ID
	todo.ParentID = newParentID

	// Add to new parent
	if newParentID != "" {
		parent, _ := tl.Get(newParentID)
		parent.Children = insertAtPosition(parent.Children, todo, position)
	} else {
		tl.roots = insertAtPosition(tl.roots, todo, position)
	}

	tl.modified = true
	return nil
}

// checkCycle verifies that moving todoID under newParentID wouldn't create a cycle
func (tl *TodoList) checkCycle(todoID string, newParentID string) error {
	// Cannot move under itself
	if todoID == newParentID {
		return errors.New("cannot move todo under itself: would create cycle")
	}

	// Check if newParent is a descendant of todo (would create cycle)
	current, err := tl.Get(newParentID)
	if err != nil {
		return err
	}

	// Walk up the tree from newParent
	visited := make(map[string]bool)
	for current.ParentID != "" {
		if visited[current.ID] {
			// Shouldn't happen, but prevents infinite loop
			return errors.New("cycle detected in existing hierarchy")
		}
		visited[current.ID] = true

		if current.ParentID == todoID {
			return errors.New("cannot move todo under its own descendant: would create cycle")
		}

		current, _ = tl.Get(current.ParentID)
	}

	return nil
}

// Promote moves a todo up one level in the hierarchy
// No-op if already at root level
func (tl *TodoList) Promote(id string) error {
	todo, err := tl.Get(id)
	if err != nil {
		return err
	}

	// No-op for root todos
	if todo.ParentID == "" {
		return nil
	}

	// Get current parent
	parent, _ := tl.Get(todo.ParentID)

	// Remove from current parent
	parent.Children = removeFromSlice(parent.Children, todo)

	// Move to grandparent (or root if parent is root)
	todo.ParentID = parent.ParentID

	if todo.ParentID != "" {
		grandparent, _ := tl.Get(todo.ParentID)
		grandparent.Children = append(grandparent.Children, todo)
	} else {
		tl.roots = append(tl.roots, todo)
	}

	tl.modified = true
	return nil
}

// Demote moves a todo under a sibling
// Position: 0 = beginning, -1 = end (default)
func (tl *TodoList) Demote(id string, siblingID string, position ...int) error {
	todo, err := tl.Get(id)
	if err != nil {
		return err
	}

	// Validate sibling exists
	sibling, err := tl.Get(siblingID)
	if err != nil {
		return err
	}

	// Verify they are actually siblings (same parent)
	if todo.ParentID != sibling.ParentID {
		return errors.New("cannot demote: specified todo is not a sibling")
	}

	// Cannot demote under itself
	if todo.ID == siblingID {
		return errors.New("cannot demote todo under itself")
	}

	// Remove from current parent
	if todo.ParentID != "" {
		parent, _ := tl.Get(todo.ParentID)
		parent.Children = removeFromSlice(parent.Children, todo)
	} else {
		tl.roots = removeFromSlice(tl.roots, todo)
	}

	// Move under sibling
	todo.ParentID = siblingID

	// Determine position (default to -1 = end if not specified)
	pos := -1
	if len(position) > 0 {
		pos = position[0]
	}

	// Insert at specified position
	sibling.Children = insertAtPosition(sibling.Children, todo, pos)

	tl.modified = true
	return nil
}

// Reorder changes the position of a todo within its current parent
// Returns error for out-of-bounds positions (Option A)
func (tl *TodoList) Reorder(id string, newPosition int) error {
	todo, err := tl.Get(id)
	if err != nil {
		return err
	}

	// Get the slice to reorder
	var slice []*Todo
	if todo.ParentID != "" {
		parent, _ := tl.Get(todo.ParentID)
		slice = parent.Children
	} else {
		slice = tl.roots
	}

	// Validate position bounds (Option A: return error)
	if newPosition < 0 {
		return errors.New("position cannot be negative")
	}
	if newPosition >= len(slice) {
		return errors.New("position out of bounds")
	}

	// Find current position
	currentPos := -1
	for i, t := range slice {
		if t.ID == todo.ID {
			currentPos = i
			break
		}
	}

	if currentPos == -1 {
		return errors.New("todo not found in parent's children")
	}

	// No change needed
	if currentPos == newPosition {
		tl.modified = true
		return nil
	}

	// Remove from current position
	slice = append(slice[:currentPos], slice[currentPos+1:]...)

	// Adjust position if moving to a higher index
	// (because removal shifted elements down)
	if newPosition > currentPos {
		newPosition--
	}

	// Insert at new position
	result := make([]*Todo, 0, len(slice)+1)
	result = append(result, slice[:newPosition]...)
	result = append(result, todo)
	result = append(result, slice[newPosition:]...)

	// Update the source slice
	if todo.ParentID != "" {
		parent, _ := tl.Get(todo.ParentID)
		parent.Children = result
	} else {
		tl.roots = result
	}

	tl.modified = true
	return nil
}
