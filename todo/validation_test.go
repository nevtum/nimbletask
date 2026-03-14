package todo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test helpers using functional options pattern

// WithOrphanTodo adds a todo with a non-existent parent ID
func WithOrphanTodo(id, title, fakeParentID string) Option {
	return func(tl *TodoList) {
		todo := &Todo{
			ID:        id,
			Title:     title,
			ParentID:  fakeParentID,
			Children:  []*Todo{},
			CreatedAt: tl.clock.Now(),
			UpdatedAt: tl.clock.Now(),
		}
		tl.todos[id] = todo
	}
}

// WithRootWithParentID adds a todo to roots slice that has a non-empty ParentID
func WithRootWithParentID(id, title, fakeParentID string) Option {
	return func(tl *TodoList) {
		todo := &Todo{
			ID:        id,
			Title:     title,
			ParentID:  fakeParentID,
			Children:  []*Todo{},
			CreatedAt: tl.clock.Now(),
			UpdatedAt: tl.clock.Now(),
		}
		tl.todos[id] = todo
		tl.roots = append(tl.roots, todo)
	}
}

// WithChildNotInParentChildren adds a child to the map but not to parent's Children slice
func WithChildNotInParentChildren(childID, childTitle, parentID string) Option {
	return func(tl *TodoList) {
		child := &Todo{
			ID:        childID,
			Title:     childTitle,
			ParentID:  parentID,
			Children:  []*Todo{},
			CreatedAt: tl.clock.Now(),
			UpdatedAt: tl.clock.Now(),
		}
		tl.todos[childID] = child
		// Intentionally NOT adding to parent.Children
	}
}

// WithGhostChild adds a child to parent's Children slice but not to the map
func WithGhostChild(parentID, ghostID, ghostTitle string) Option {
	return func(tl *TodoList) {
		parent := tl.todos[parentID]
		if parent == nil {
			return
		}
		ghost := &Todo{
			ID:        ghostID,
			Title:     ghostTitle,
			ParentID:  parentID,
			Children:  []*Todo{},
			CreatedAt: tl.clock.Now(),
			UpdatedAt: tl.clock.Now(),
		}
		parent.Children = append(parent.Children, ghost)
		// Intentionally NOT adding to tl.todos
	}
}

// WithSelfParent adds a todo that references itself as parent
func WithSelfParent(id, title string) Option {
	return func(tl *TodoList) {
		todo := &Todo{
			ID:        id,
			Title:     title,
			ParentID:  id, // Self-reference
			Children:  []*Todo{},
			CreatedAt: tl.clock.Now(),
			UpdatedAt: tl.clock.Now(),
		}
		tl.todos[id] = todo
	}
}

// WithCycle creates a 3-node cycle: A -> B -> C -> A
// Requires nodes to exist first via Add()
func WithCycle(nodeAID, nodeBID, nodeCID string) Option {
	return func(tl *TodoList) {
		nodeA := tl.todos[nodeAID]
		nodeB := tl.todos[nodeBID]
		nodeC := tl.todos[nodeCID]
		if nodeA == nil || nodeB == nil || nodeC == nil {
			return
		}

		// Remove A from roots (it's becoming a child)
		newRoots := []*Todo{}
		for _, root := range tl.roots {
			if root.ID != nodeAID {
				newRoots = append(newRoots, root)
			}
		}
		tl.roots = newRoots

		// Create cycle: C -> A
		nodeA.ParentID = nodeCID
		nodeC.Children = append(nodeC.Children, nodeA)
	}
}

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
		tl := NewTodoList(WithOrphanTodo("orphan-id", "Orphan", "non-existent-parent"))

		err := tl.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parent")
	})

	t.Run("returns error when root todo has ParentID but is in roots", func(t *testing.T) {
		tl := NewTodoList(WithRootWithParentID("invalid-root", "Invalid Root", "some-parent"))

		err := tl.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "root")
	})

	t.Run("returns error when child not in parent's Children slice", func(t *testing.T) {
		tl := NewTodoList()
		parent, _ := tl.Add("Parent", "", -1)

		// Use helper to add child to map but not to parent's Children
		opt := WithChildNotInParentChildren("child-id", "Child", parent.ID)
		opt(tl)

		err := tl.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "child")
	})

	t.Run("returns error when parent references child not in map", func(t *testing.T) {
		tl := NewTodoList()
		parent, _ := tl.Add("Parent", "", -1)

		// Use helper to add ghost child to parent's Children but not to map
		opt := WithGhostChild(parent.ID, "ghost-child", "Ghost")
		opt(tl)

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

		// Create cycle: C -> A using helper
		opt := WithCycle(nodeA.ID, nodeB.ID, nodeC.ID)
		opt(tl)

		err := tl.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cycle")
	})

	t.Run("returns error when todo in roots has non-empty ParentID", func(t *testing.T) {
		tl := NewTodoList(WithRootWithParentID("bad-root", "Bad Root", "fake-parent"))

		err := tl.Validate()

		assert.Error(t, err)
	})

	t.Run("returns error when todo has itself as parent", func(t *testing.T) {
		tl := NewTodoList(WithSelfParent("self-parent", "Self Parent"))

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
