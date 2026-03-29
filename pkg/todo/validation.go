package todo

import (
	"fmt"
)

// Validate checks the integrity of the entire todo list
// Returns nil if valid, error describing the problem otherwise
func (tl *TodoList) Validate() error {
	// Check each todo in the map
	for id, todo := range tl.todos {
		// Check that todo's ID matches its key in the map
		if todo.ID != id {
			return fmt.Errorf("todo ID mismatch: map key %q but todo.ID is %q", id, todo.ID)
		}

		// Check that a todo doesn't reference itself as parent
		if todo.ParentID == todo.ID {
			return fmt.Errorf("todo %q references itself as parent", id)
		}

		// If todo has a parent, check that parent exists
		if todo.ParentID != "" {
			parent, exists := tl.todos[todo.ParentID]
			if !exists {
				return fmt.Errorf("todo %q references non-existent parent %q", id, todo.ParentID)
			}

			// Check that this todo is in its parent's Children slice
			foundInParent := false
			for _, child := range parent.Children {
				if child.ID == id {
					foundInParent = true
					break
				}
			}
			if !foundInParent {
				return fmt.Errorf("todo %q has parent %q but is not in parent's Children slice", id, todo.ParentID)
			}
		}

		// Check that all children exist in the map and reference this todo as parent
		for _, child := range todo.Children {
			childFromMap, exists := tl.todos[child.ID]
			if !exists {
				return fmt.Errorf("todo %q references child %q which is not in the todos map", id, child.ID)
			}
			if childFromMap.ParentID != id {
				return fmt.Errorf("todo %q has child %q but child's ParentID is %q", id, child.ID, childFromMap.ParentID)
			}
		}
	}

	// Check roots slice
	for _, root := range tl.roots {
		// Check that root exists in map
		if _, exists := tl.todos[root.ID]; !exists {
			return fmt.Errorf("root todo %q is not in the todos map", root.ID)
		}

		// Check that root has empty ParentID
		if root.ParentID != "" {
			return fmt.Errorf("root todo %q has non-empty ParentID %q", root.ID, root.ParentID)
		}
	}

	// Check for cycles in the hierarchy
	for id := range tl.todos {
		if err := tl.checkCycleFromNode(id); err != nil {
			return fmt.Errorf("cycle detected: %w", err)
		}
	}

	return nil
}

// checkCycleFromNode checks if there's a cycle starting from a specific node
// by walking up the parent chain
func (tl *TodoList) checkCycleFromNode(startID string) error {
	visited := make(map[string]bool)
	currentID := startID

	for currentID != "" {
		if visited[currentID] {
			return fmt.Errorf("cycle detected at todo %q", currentID)
		}
		visited[currentID] = true

		todo := tl.todos[currentID]

		currentID = todo.ParentID
	}

	return nil
}

// CanMove checks if moving a todo to a new parent would be valid
// Returns true if the move is valid, false otherwise
// Returns error if either ID doesn't exist
func (tl *TodoList) CanMove(id string, newParentID string) (bool, error) {
	// Check that todo exists
	todo, err := tl.Get(id)
	if err != nil {
		return false, err
	}

	// Check that new parent exists (if specified)
	if newParentID != "" {
		_, err = tl.Get(newParentID)
		if err != nil {
			return false, err
		}
	}

	// Check if this would be a no-op (same parent)
	if todo.ParentID == newParentID {
		return true, nil
	}

	// Check for cycles
	if newParentID != "" {
		if err := tl.checkCycle(id, newParentID); err != nil {
			// Cycle detected - move is not valid, but not an error
			return false, nil
		}
	}

	return true, nil
}
