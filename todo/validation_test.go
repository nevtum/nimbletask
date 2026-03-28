package todo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestValidate verifies the Validate operation for list integrity
func TestValidate(t *testing.T) {
	t.Run("returns nil for valid empty list", func(t *testing.T) {
		tl := newTestTodoList()

		err := tl.Validate()

		assert.NoError(t, err)
	})

	t.Run("returns nil for valid single root todo", func(t *testing.T) {
		tl := newTestTodoList()
		tl.Add("Root", "", -1)

		err := tl.Validate()

		assert.NoError(t, err)
	})

	t.Run("returns nil for valid hierarchical list", func(t *testing.T) {
		tl := newTestTodoList()
		parent, _ := tl.Add("Parent", "", -1)
		tl.Add("Child", parent.ID, -1)

		err := tl.Validate()

		assert.NoError(t, err)
	})

	t.Run("returns nil for valid deeply nested list", func(t *testing.T) {
		tl := newTestTodoList()
		grandparent, _ := tl.Add("Grandparent", "", -1)
		parent, _ := tl.Add("Parent", grandparent.ID, -1)
		tl.Add("Child", parent.ID, -1)

		err := tl.Validate()

		assert.NoError(t, err)
	})

	t.Run("returns error when todo references non-existent parent", func(t *testing.T) {
		tl := newTestTodoList(
			withOrphanTodo("orphan-id", "Orphan", "non-existent-parent"),
		)

		err := tl.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parent")
	})

	t.Run("returns error when root todo has ParentID but is in roots", func(t *testing.T) {
		tl := newTestTodoList(
			withRootWithParentID("invalid-root", "Invalid Root", "some-parent"),
		)

		err := tl.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "root")
	})

	t.Run("returns error when child not in parent's Children slice", func(t *testing.T) {
		tl := newTestTodoList()
		parent, _ := tl.Add("Parent", "", -1)

		// Use helper to add child to map but not to parent's Children
		opt := withChildNotInParentChildren("child-id", "Child", parent.ID)
		opt(tl)

		err := tl.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "child")
	})

	t.Run("returns error when parent references child not in map", func(t *testing.T) {
		tl := newTestTodoList()
		parent, _ := tl.Add("Parent", "", -1)

		// Use helper to add ghost child to parent's Children but not to map
		opt := withGhostChild(parent.ID, "ghost-child", "Ghost")
		opt(tl)

		err := tl.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "map")
	})

	t.Run("returns error when cycle exists in hierarchy", func(t *testing.T) {
		tl := newTestTodoList()
		// Create a 3-node chain: A -> B -> C
		nodeA, _ := tl.Add("Node A", "", -1)
		nodeB, _ := tl.Add("Node B", nodeA.ID, -1)
		nodeC, _ := tl.Add("Node C", nodeB.ID, -1)

		// Create cycle: C -> A using helper
		opt := withCycle(nodeA.ID, nodeB.ID, nodeC.ID)
		opt(tl)

		err := tl.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cycle")
	})

	t.Run("returns error when todo has itself as parent", func(t *testing.T) {
		tl := newTestTodoList(withSelfParent("self-parent", "Self Parent"))

		err := tl.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parent")
	})
}

// TestCanMove verifies the CanMove operation for validating moves without executing
func TestCanMove(t *testing.T) {
	t.Run("returns true for valid move to new parent", func(t *testing.T) {
		tl := newTestTodoList()
		parent, _ := tl.Add("Parent", "", -1)
		child, _ := tl.Add("Child", "", -1)

		canMove, err := tl.CanMove(child.ID, parent.ID)

		assert.NoError(t, err)
		assert.True(t, canMove)
	})

	t.Run("returns true for valid move to root", func(t *testing.T) {
		tl := newTestTodoList()
		parent, _ := tl.Add("Parent", "", -1)
		child, _ := tl.Add("Child", parent.ID, -1)

		canMove, err := tl.CanMove(child.ID, "")

		assert.NoError(t, err)
		assert.True(t, canMove)
	})

	t.Run("returns false when move would create direct cycle", func(t *testing.T) {
		tl := newTestTodoList()
		parent, _ := tl.Add("Parent", "", -1)
		child, _ := tl.Add("Child", parent.ID, -1)

		canMove, err := tl.CanMove(parent.ID, child.ID)

		assert.NoError(t, err)
		assert.False(t, canMove)
	})

	t.Run("returns false when move would create indirect cycle", func(t *testing.T) {
		tl := newTestTodoList()
		grandparent, _ := tl.Add("Grandparent", "", -1)
		parent, _ := tl.Add("Parent", grandparent.ID, -1)
		child, _ := tl.Add("Child", parent.ID, -1)

		canMove, err := tl.CanMove(grandparent.ID, child.ID)

		assert.NoError(t, err)
		assert.False(t, canMove)
	})

	t.Run("returns false when todo does not exist", func(t *testing.T) {
		tl := newTestTodoList()
		parent, _ := tl.Add("Parent", "", -1)

		canMove, err := tl.CanMove("non-existent", parent.ID)

		assert.Error(t, err)
		assert.False(t, canMove)
	})

	t.Run("returns false when new parent does not exist", func(t *testing.T) {
		tl := newTestTodoList()
		child, _ := tl.Add("Child", "", -1)

		canMove, err := tl.CanMove(child.ID, "non-existent")

		assert.Error(t, err)
		assert.False(t, canMove)
	})

	t.Run("returns true when moving to same parent", func(t *testing.T) {
		tl := newTestTodoList()
		parent, _ := tl.Add("Parent", "", -1)
		child, _ := tl.Add("Child", parent.ID, -1)

		canMove, err := tl.CanMove(child.ID, parent.ID)

		assert.NoError(t, err)
		assert.True(t, canMove)
	})

	t.Run("returns true when moving root to root (same parent)", func(t *testing.T) {
		tl := newTestTodoList()
		root, _ := tl.Add("Root", "", -1)

		canMove, err := tl.CanMove(root.ID, "")

		assert.NoError(t, err)
		assert.True(t, canMove)
	})

	t.Run("returns false when attempting to move under itself", func(t *testing.T) {
		tl := newTestTodoList()
		todo, _ := tl.Add("Todo", "", -1)

		canMove, err := tl.CanMove(todo.ID, todo.ID)

		assert.NoError(t, err)
		assert.False(t, canMove)
	})

	t.Run("returns true for complex valid move", func(t *testing.T) {
		tl := newTestTodoList()
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
