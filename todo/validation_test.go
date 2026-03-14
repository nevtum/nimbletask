package todo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestValidate verifies the Validate operation for list integrity
func TestValidate(t *testing.T) {
	t.Run("returns nil for valid empty list", func(t *testing.T) {
		tl := NewTodoList()

		err := tl.Validate()

		assert.NoError(t, err)
	})

	t.Run("returns nil for valid single root todo", func(t *testing.T) {
		tl := NewTodoList()
		tl.Add("Root", "", -1)

		err := tl.Validate()

		assert.NoError(t, err)
	})

	t.Run("returns nil for valid hierarchical list", func(t *testing.T) {
		tl := NewTodoList()
		parent, _ := tl.Add("Parent", "", -1)
		tl.Add("Child", parent.ID, -1)

		err := tl.Validate()

		assert.NoError(t, err)
	})

	t.Run("returns nil for valid deeply nested list", func(t *testing.T) {
		tl := NewTodoList()
		grandparent, _ := tl.Add("Grandparent", "", -1)
		parent, _ := tl.Add("Parent", grandparent.ID, -1)
		tl.Add("Child", parent.ID, -1)

		err := tl.Validate()

		assert.NoError(t, err)
	})

	t.Run("returns error when todo references non-existent parent", func(t *testing.T) {
		tl := NewTodoList()
		todo := &Todo{
			ID:        "orphan-id",
			Title:     "Orphan",
			ParentID:  "non-existent-parent",
			Children:  []*Todo{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		tl.todos[todo.ID] = todo

		err := tl.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parent")
	})

	t.Run("returns error when root todo has ParentID but is in roots", func(t *testing.T) {
		tl := NewTodoList()
		todo := &Todo{
			ID:        "invalid-root",
			Title:     "Invalid Root",
			ParentID:  "some-parent",
			Children:  []*Todo{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		tl.todos[todo.ID] = todo
		tl.roots = append(tl.roots, todo)

		err := tl.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "root")
	})

	t.Run("returns error when child not in parent's Children slice", func(t *testing.T) {
		tl := NewTodoList()
		parent, _ := tl.Add("Parent", "", -1)
		child := &Todo{
			ID:        "child-id",
			Title:     "Child",
			ParentID:  parent.ID,
			Children:  []*Todo{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		tl.todos[child.ID] = child
		// Intentionally NOT adding child to parent.Children

		err := tl.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "child")
	})

	t.Run("returns error when parent references child not in map", func(t *testing.T) {
		tl := NewTodoList()
		parent, _ := tl.Add("Parent", "", -1)
		child := &Todo{
			ID:        "ghost-child",
			Title:     "Ghost",
			ParentID:  parent.ID,
			Children:  []*Todo{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		parent.Children = append(parent.Children, child)
		// Intentionally NOT adding child to tl.todos

		err := tl.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "map")
	})

	t.Run("returns error when cycle exists in hierarchy", func(t *testing.T) {
		tl := NewTodoList()
		// Create a 3-node chain: A -> B -> C
		nodeA, _ := tl.Add("Node A", "", -1)
		nodeB, _ := tl.Add("Node B", nodeA.ID, -1)
		nodeC, _ := tl.Add("Node C", nodeB.ID, -1)

		// Create cycle: make C the parent of A
		// First remove A from roots
		tl.roots = []*Todo{}
		// Set A's parent to C
		nodeA.ParentID = nodeC.ID
		// Add A to C's children
		nodeC.Children = append(nodeC.Children, nodeA)

		err := tl.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cycle")
	})

	t.Run("returns error when todo in roots has non-empty ParentID", func(t *testing.T) {
		tl := NewTodoList()
		todo := &Todo{
			ID:        "bad-root",
			Title:     "Bad Root",
			ParentID:  "fake-parent",
			Children:  []*Todo{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		tl.todos[todo.ID] = todo
		tl.roots = append(tl.roots, todo)

		err := tl.Validate()

		assert.Error(t, err)
	})

	t.Run("returns error when todo has itself as parent", func(t *testing.T) {
		tl := NewTodoList()
		todo := &Todo{
			ID:        "self-parent",
			Title:     "Self Parent",
			ParentID:  "self-parent",
			Children:  []*Todo{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		tl.todos[todo.ID] = todo

		err := tl.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parent")
	})
}

// TestCanMove verifies the CanMove operation for validating moves without executing
func TestCanMove(t *testing.T) {
	t.Run("returns true for valid move to new parent", func(t *testing.T) {
		tl := NewTodoList()
		parent, _ := tl.Add("Parent", "", -1)
		child, _ := tl.Add("Child", "", -1)

		canMove, err := tl.CanMove(child.ID, parent.ID)

		assert.NoError(t, err)
		assert.True(t, canMove)
	})

	t.Run("returns true for valid move to root", func(t *testing.T) {
		tl := NewTodoList()
		parent, _ := tl.Add("Parent", "", -1)
		child, _ := tl.Add("Child", parent.ID, -1)

		canMove, err := tl.CanMove(child.ID, "")

		assert.NoError(t, err)
		assert.True(t, canMove)
	})

	t.Run("returns false when move would create direct cycle", func(t *testing.T) {
		tl := NewTodoList()
		parent, _ := tl.Add("Parent", "", -1)
		child, _ := tl.Add("Child", parent.ID, -1)

		canMove, err := tl.CanMove(parent.ID, child.ID)

		assert.NoError(t, err)
		assert.False(t, canMove)
	})

	t.Run("returns false when move would create indirect cycle", func(t *testing.T) {
		tl := NewTodoList()
		grandparent, _ := tl.Add("Grandparent", "", -1)
		parent, _ := tl.Add("Parent", grandparent.ID, -1)
		child, _ := tl.Add("Child", parent.ID, -1)

		canMove, err := tl.CanMove(grandparent.ID, child.ID)

		assert.NoError(t, err)
		assert.False(t, canMove)
	})

	t.Run("returns false when todo does not exist", func(t *testing.T) {
		tl := NewTodoList()
		parent, _ := tl.Add("Parent", "", -1)

		canMove, err := tl.CanMove("non-existent", parent.ID)

		assert.Error(t, err)
		assert.False(t, canMove)
	})

	t.Run("returns false when new parent does not exist", func(t *testing.T) {
		tl := NewTodoList()
		child, _ := tl.Add("Child", "", -1)

		canMove, err := tl.CanMove(child.ID, "non-existent")

		assert.Error(t, err)
		assert.False(t, canMove)
	})

	t.Run("returns true when moving to same parent", func(t *testing.T) {
		tl := NewTodoList()
		parent, _ := tl.Add("Parent", "", -1)
		child, _ := tl.Add("Child", parent.ID, -1)

		canMove, err := tl.CanMove(child.ID, parent.ID)

		assert.NoError(t, err)
		assert.True(t, canMove)
	})

	t.Run("returns true when moving root to root (same parent)", func(t *testing.T) {
		tl := NewTodoList()
		root, _ := tl.Add("Root", "", -1)

		canMove, err := tl.CanMove(root.ID, "")

		assert.NoError(t, err)
		assert.True(t, canMove)
	})

	t.Run("returns false when attempting to move under itself", func(t *testing.T) {
		tl := NewTodoList()
		todo, _ := tl.Add("Todo", "", -1)

		canMove, err := tl.CanMove(todo.ID, todo.ID)

		assert.NoError(t, err)
		assert.False(t, canMove)
	})

	t.Run("returns true for complex valid move", func(t *testing.T) {
		tl := NewTodoList()
		// Create two separate trees
		tree1Root, _ := tl.Add("Tree1 Root", "", -1)
		tree1Child, _ := tl.Add("Tree1 Child", tree1Root.ID, -1)
		tree2Root, _ := tl.Add("Tree2 Root", "", -1)

		// Move tree1's child under tree2 (should be valid)
		canMove, err := tl.CanMove(tree1Child.ID, tree2Root.ID)

		assert.NoError(t, err)
		assert.True(t, canMove)
	})
}
