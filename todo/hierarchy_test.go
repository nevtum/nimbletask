package todo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestMove verifies moving todos between parents
func TestMove(t *testing.T) {
	t.Run("moves child to new parent", func(t *testing.T) {
		tl := NewTodoList(withClock(NewTestClock(time.Now())))
		parent1, _ := tl.Add("Parent 1", "", -1)
		parent2, _ := tl.Add("Parent 2", "", -1)
		child, _ := tl.Add("Child", parent1.ID, -1)

		tl.modified = false

		err := tl.Move(child.ID, parent2.ID, -1)

		assert.NoError(t, err)
		assert.Equal(t, parent2.ID, child.ParentID)
		assert.Len(t, parent1.Children, 0)
		assert.Len(t, parent2.Children, 1)
		assert.Equal(t, child.ID, parent2.Children[0].ID)
		assert.True(t, tl.modified)
	})

	t.Run("moves to root level", func(t *testing.T) {
		tl := NewTodoList(withClock(NewTestClock(time.Now())))
		parent, _ := tl.Add("Parent", "", -1)
		child, _ := tl.Add("Child", parent.ID, -1)

		err := tl.Move(child.ID, "", -1)

		assert.NoError(t, err)
		assert.Equal(t, "", child.ParentID)
		assert.Len(t, parent.Children, 0)
		assert.Len(t, tl.roots, 2)
	})

	t.Run("moves root to child", func(t *testing.T) {
		tl := NewTodoList(withClock(NewTestClock(time.Now())))
		root, _ := tl.Add("Root", "", -1)
		parent, _ := tl.Add("Parent", "", -1)

		err := tl.Move(root.ID, parent.ID, -1)

		assert.NoError(t, err)
		assert.Equal(t, parent.ID, root.ParentID)
		assert.Len(t, tl.roots, 1)
		assert.Len(t, parent.Children, 1)
	})

	t.Run("moves with specific position", func(t *testing.T) {
		tl := NewTodoList(withClock(NewTestClock(time.Now())))
		parent, _ := tl.Add("Parent", "", -1)
		child1, _ := tl.Add("Child 1", parent.ID, -1)
		child2, _ := tl.Add("Child 2", parent.ID, -1)
		movingChild, _ := tl.Add("Moving", "", -1)

		err := tl.Move(movingChild.ID, parent.ID, 1)

		assert.NoError(t, err)
		assert.Equal(t, child1.ID, parent.Children[0].ID)
		assert.Equal(t, movingChild.ID, parent.Children[1].ID)
		assert.Equal(t, child2.ID, parent.Children[2].ID)
	})

	t.Run("no-op when moving to same parent at same position", func(t *testing.T) {
		tl := NewTodoList(withClock(NewTestClock(time.Now())))
		parent, _ := tl.Add("Parent", "", -1)
		child, _ := tl.Add("Child", parent.ID, -1)

		tl.modified = false

		err := tl.Move(child.ID, parent.ID, -1)

		assert.NoError(t, err)
		assert.False(t, tl.modified)
	})

	t.Run("returns error for non-existent todo", func(t *testing.T) {
		tl := NewTodoList(withClock(NewTestClock(time.Now())))
		parent, _ := tl.Add("Parent", "", -1)

		err := tl.Move("non-existent", parent.ID, -1)

		assert.Error(t, err)
	})

	t.Run("returns error for non-existent new parent", func(t *testing.T) {
		tl := NewTodoList(withClock(NewTestClock(time.Now())))
		todo, _ := tl.Add("Todo", "", -1)

		err := tl.Move(todo.ID, "non-existent", -1)

		assert.Error(t, err)
	})

	t.Run("prevents direct circular reference", func(t *testing.T) {
		tl := NewTodoList(withClock(NewTestClock(time.Now())))
		parent, _ := tl.Add("Parent", "", -1)
		child, _ := tl.Add("Child", parent.ID, -1)

		// Try to move parent under its own child
		err := tl.Move(parent.ID, child.ID, -1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cycle")
	})

	t.Run("prevents indirect circular reference", func(t *testing.T) {
		tl := NewTodoList(withClock(NewTestClock(time.Now())))
		grandparent, _ := tl.Add("Grandparent", "", -1)
		parent, _ := tl.Add("Parent", grandparent.ID, -1)
		child, _ := tl.Add("Child", parent.ID, -1)

		// Try to move grandparent under its great-grandchild (which would create cycle)
		err := tl.Move(grandparent.ID, child.ID, -1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cycle")
	})

	t.Run("prevents moving todo under itself", func(t *testing.T) {
		tl := NewTodoList(withClock(NewTestClock(time.Now())))
		todo, _ := tl.Add("Todo", "", -1)

		err := tl.Move(todo.ID, todo.ID, -1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cycle")
	})
}

// TestPromote verifies promoting todos up the hierarchy
func TestPromote(t *testing.T) {
	t.Run("promotes child to root", func(t *testing.T) {
		tl := NewTodoList(withClock(NewTestClock(time.Now())))
		parent, _ := tl.Add("Parent", "", -1)
		child, _ := tl.Add("Child", parent.ID, -1)

		tl.modified = false

		err := tl.Promote(child.ID)

		assert.NoError(t, err)
		assert.Equal(t, "", child.ParentID)
		assert.Len(t, parent.Children, 0)
		assert.Len(t, tl.roots, 2)
		assert.True(t, tl.modified)
	})

	t.Run("promotes grandchild to sibling of parent", func(t *testing.T) {
		tl := NewTodoList(withClock(NewTestClock(time.Now())))
		grandparent, _ := tl.Add("Grandparent", "", -1)
		parent, _ := tl.Add("Parent", grandparent.ID, -1)
		grandchild, _ := tl.Add("Grandchild", parent.ID, -1)

		err := tl.Promote(grandchild.ID)

		assert.NoError(t, err)
		assert.Equal(t, grandparent.ID, grandchild.ParentID)
		assert.Len(t, parent.Children, 0)
		assert.Len(t, grandparent.Children, 2)
	})

	t.Run("no-op for root todo", func(t *testing.T) {
		tl := NewTodoList(withClock(NewTestClock(time.Now())))
		root, _ := tl.Add("Root", "", -1)

		tl.modified = false

		err := tl.Promote(root.ID)

		assert.NoError(t, err)
		assert.Equal(t, "", root.ParentID)
		assert.Len(t, tl.roots, 1)
		assert.False(t, tl.modified) // No change, should not mark modified
	})

	t.Run("returns error for non-existent todo", func(t *testing.T) {
		tl := NewTodoList(withClock(NewTestClock(time.Now())))

		err := tl.Promote("non-existent")

		assert.Error(t, err)
	})
}

// TestDemote verifies demoting todos under siblings
func TestDemote(t *testing.T) {
	t.Run("demotes root under sibling at end", func(t *testing.T) {
		tl := NewTodoList(withClock(NewTestClock(time.Now())))
		sibling, _ := tl.Add("Sibling", "", -1)
		todo, _ := tl.Add("Todo", "", -1)

		tl.modified = false

		err := tl.Demote(todo.ID, sibling.ID)

		assert.NoError(t, err)
		assert.Equal(t, sibling.ID, todo.ParentID)
		assert.Len(t, tl.roots, 1)
		assert.Len(t, sibling.Children, 1)
		assert.Equal(t, todo.ID, sibling.Children[0].ID)
		assert.True(t, tl.modified)
	})

	t.Run("demotes root under sibling at beginning", func(t *testing.T) {
		tl := NewTodoList(withClock(NewTestClock(time.Now())))
		sibling, _ := tl.Add("Sibling", "", -1)
		_, _ = tl.Add("Sibling2", "", -1) // Ensure multiple siblings exist
		todo, _ := tl.Add("Todo", "", -1)

		err := tl.Demote(todo.ID, sibling.ID)

		assert.NoError(t, err)
		assert.Equal(t, sibling.ID, todo.ParentID)
		assert.Len(t, sibling.Children, 1)
		assert.Equal(t, todo.ID, sibling.Children[0].ID)
	})

	t.Run("demotes child under sibling", func(t *testing.T) {
		tl := NewTodoList(withClock(NewTestClock(time.Now())))
		parent, _ := tl.Add("Parent", "", -1)
		sibling, _ := tl.Add("Sibling", parent.ID, -1)
		todo, _ := tl.Add("Todo", parent.ID, -1)

		err := tl.Demote(todo.ID, sibling.ID)

		assert.NoError(t, err)
		assert.Equal(t, sibling.ID, todo.ParentID)
		assert.Len(t, parent.Children, 1) // Only sibling remains
		assert.Len(t, sibling.Children, 1)
	})

	t.Run("returns error for non-existent todo", func(t *testing.T) {
		tl := NewTodoList(withClock(NewTestClock(time.Now())))
		sibling, _ := tl.Add("Sibling", "", -1)

		err := tl.Demote("non-existent", sibling.ID)

		assert.Error(t, err)
	})

	t.Run("returns error for non-existent sibling", func(t *testing.T) {
		tl := NewTodoList(withClock(NewTestClock(time.Now())))
		todo, _ := tl.Add("Todo", "", -1)

		err := tl.Demote(todo.ID, "non-existent")

		assert.Error(t, err)
	})

	t.Run("returns error for non-existent sibling", func(t *testing.T) {
		tl := NewTodoList(withClock(NewTestClock(time.Now())))
		todo, _ := tl.Add("Todo", "", -1)

		err := tl.Demote(todo.ID, "any-id")

		assert.Error(t, err)
	})

	t.Run("returns error when sibling is not actually a sibling", func(t *testing.T) {
		tl := NewTodoList(withClock(NewTestClock(time.Now())))
		parent, _ := tl.Add("Parent", "", -1)
		child, _ := tl.Add("Child", parent.ID, -1)
		unrelated, _ := tl.Add("Unrelated", "", -1)

		err := tl.Demote(child.ID, unrelated.ID)

		assert.Error(t, err)
	})
}

// TestReorder verifies reordering todos within their parent
func TestReorder(t *testing.T) {
	t.Run("reorders within same parent", func(t *testing.T) {
		tl := NewTodoList(withClock(NewTestClock(time.Now())))
		parent, _ := tl.Add("Parent", "", -1)
		child1, _ := tl.Add("Child 1", parent.ID, -1)
		child2, _ := tl.Add("Child 2", parent.ID, -1)
		child3, _ := tl.Add("Child 3", parent.ID, -1)

		tl.modified = false

		// Move child3 to position 0
		err := tl.Reorder(child3.ID, 0)

		assert.NoError(t, err)
		assert.Equal(t, child3.ID, parent.Children[0].ID)
		assert.Equal(t, child1.ID, parent.Children[1].ID)
		assert.Equal(t, child2.ID, parent.Children[2].ID)
		assert.True(t, tl.modified)
	})

	t.Run("reorders root todos", func(t *testing.T) {
		tl := NewTodoList(withClock(NewTestClock(time.Now())))
		root1, _ := tl.Add("Root 1", "", -1)
		root2, _ := tl.Add("Root 2", "", -1)
		root3, _ := tl.Add("Root 3", "", -1)

		err := tl.Reorder(root3.ID, 0)

		assert.NoError(t, err)
		assert.Equal(t, root3.ID, tl.roots[0].ID)
		assert.Equal(t, root1.ID, tl.roots[1].ID)
		assert.Equal(t, root2.ID, tl.roots[2].ID)
	})

	t.Run("reorders to end", func(t *testing.T) {
		tl := NewTodoList(withClock(NewTestClock(time.Now())))
		parent, _ := tl.Add("Parent", "", -1)
		child1, _ := tl.Add("Child 1", parent.ID, -1)
		child2, _ := tl.Add("Child 2", parent.ID, -1)
		_ = child1 // silence unused warning

		// Move child1 to position 1 (end)
		err := tl.Reorder(child1.ID, 1)

		assert.NoError(t, err)
		assert.Equal(t, child2.ID, parent.Children[0].ID)
		assert.Equal(t, child1.ID, parent.Children[1].ID)
	})

	t.Run("returns error for non-existent todo", func(t *testing.T) {
		tl := NewTodoList(withClock(NewTestClock(time.Now())))

		err := tl.Reorder("non-existent", 0)

		assert.Error(t, err)
	})

	t.Run("returns error for negative position", func(t *testing.T) {
		tl := NewTodoList(withClock(NewTestClock(time.Now())))
		todo, _ := tl.Add("Todo", "", -1)

		err := tl.Reorder(todo.ID, -1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "position")
	})

	t.Run("returns error for position out of bounds", func(t *testing.T) {
		tl := NewTodoList(withClock(NewTestClock(time.Now())))
		parent, _ := tl.Add("Parent", "", -1)
		child, _ := tl.Add("Child", parent.ID, -1)
		_ = parent // silence unused warning

		err := tl.Reorder(child.ID, 5)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "position")
	})

	t.Run("no-op when reordering to current position", func(t *testing.T) {
		tl := NewTodoList(withClock(NewTestClock(time.Now())))
		parent, _ := tl.Add("Parent", "", -1)
		child, _ := tl.Add("Child", parent.ID, -1)

		tl.modified = false

		err := tl.Reorder(child.ID, 0)

		assert.NoError(t, err)
		assert.False(t, tl.modified)
	})
}
