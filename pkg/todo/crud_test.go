package todo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// assertValid is a helper to check TodoList structural integrity
func assertValid(t *testing.T, tl *TodoList) {
	t.Helper()
	assert.NoError(t, tl.Validate(), "TodoList should be structurally valid")
}

// TestNewTodoList verifies that a new TodoList is properly initialized
func TestNewTodoList(t *testing.T) {
	t.Run("initializes empty todo list", func(t *testing.T) {
		tl := newTestTodoList()

		assert.NotNil(t, tl, "NewTodoList() returned nil")
		assert.NotNil(t, tl.todos, "todos map not initialized")
		assert.NotNil(t, tl.roots, "roots slice not initialized")
		assert.Len(t, tl.todos, 0, "expected 0 todos")
		assert.Len(t, tl.roots, 0, "expected 0 roots")
		assertValid(t, tl)
	})
}

// TestTodoCreation verifies that a Todo struct can be created with all fields
func TestTodoCreation(t *testing.T) {
	t.Run("creates todo with all fields", func(t *testing.T) {
		now := time.Now()
		due := now.Add(24 * time.Hour)

		todo := &Todo{
			ID:          "test-id",
			Title:       "Test Todo",
			Description: "Test Description",
			Completed:   false,
			CreatedAt:   now,
			UpdatedAt:   now,
			ParentID:    "",
			Children:    []*Todo{},
			Priority:    1,
			DueDate:     &due,
			Tags:        []string{"test", "todo"},
		}

		// TODO: Simplify with a validate method instead
		assert.Equal(t, "test-id", todo.ID, "ID mismatch")
		assert.Equal(t, "Test Todo", todo.Title, "Title mismatch")
		assert.Equal(t, "Test Description", todo.Description, "Description mismatch")
		assert.False(t, todo.Completed, "Expected Completed to be false")
		assert.Equal(t, 1, todo.Priority, "Priority mismatch")
		assert.Len(t, todo.Tags, 2, "Tags length mismatch")
	})
}

// TestAddTodo verifies adding todos to the list
func TestAddTodo(t *testing.T) {
	t.Run("adds root-level todo", func(t *testing.T) {
		tl := newTestTodoList()

		todo, err := tl.Add("Root Todo", "", -1)
		assert.NoError(t, err, "Add failed")
		assert.NotNil(t, todo, "Add returned nil todo")
		assert.Equal(t, "Root Todo", todo.Title, "Title mismatch")
		assert.Equal(t, "", todo.ParentID, "ParentID should be empty for root")
		assert.NotEmpty(t, todo.ID, "Todo ID should not be empty")
		assertValid(t, tl)
	})

	t.Run("adds child todo with parent", func(t *testing.T) {
		tl := newTestTodoList()

		// Add parent first
		parent, err := tl.Add("Parent", "", -1)
		assert.NoError(t, err, "Failed to create parent")

		// Add child
		child, err := tl.Add("Child", parent.ID, -1)
		assert.NoError(t, err, "Failed to add child")
		assert.Equal(t, parent.ID, child.ParentID, "Child ParentID mismatch")
		assertValid(t, tl)
	})

	t.Run("returns error for non-existent parent", func(t *testing.T) {
		tl := newTestTodoList()

		_, err := tl.Add("Child", "non-existent-id", -1)
		assert.Error(t, err, "Expected error when adding to non-existent parent")
	})

	t.Run("returns error for empty title", func(t *testing.T) {
		tl := newTestTodoList()

		_, err := tl.Add("", "", -1)
		assert.Error(t, err, "Expected error when adding todo with empty title")
	})

	t.Run("adds at position 0", func(t *testing.T) {
		tl := newTestTodoList()

		// Add at position 0 (first)
		todo1, _ := tl.Add("First", "", 0)

		// Add at position 0 again (should become new first)
		todo2, _ := tl.Add("New First", "", 0)

		assert.Equal(t, todo2.ID, tl.roots[0].ID, "New todo should be at position 0")
		assert.Equal(t, todo1.ID, tl.roots[1].ID, "Original todo should be at position 1")
	})

	t.Run("adds at end with position -1", func(t *testing.T) {
		tl := newTestTodoList()

		todo1, _ := tl.Add("First", "", -1)
		todo2, _ := tl.Add("Second", "", -1)
		todo3, _ := tl.Add("Third", "", -1)

		assert.Equal(t, todo1.ID, tl.roots[0].ID, "First should be at position 0")
		assert.Equal(t, todo2.ID, tl.roots[1].ID, "Second should be at position 1")
		assert.Equal(t, todo3.ID, tl.roots[2].ID, "Third should be at position 2")
	})

	// TODO: Add a test that inserts at other positions
	// other than 0 and -1
}

// TestGetTodo verifies retrieving todos from the list
func TestGetTodo(t *testing.T) {
	t.Run("retrieves existing todo", func(t *testing.T) {
		tl := newTestTodoList()
		todo, _ := tl.Add("Test", "", -1)

		retrieved, err := tl.Get(todo.ID)
		assert.NoError(t, err, "Get failed")
		assert.Equal(t, todo.ID, retrieved.ID, "Retrieved wrong todo")
		assert.Equal(t, "Test", retrieved.Title, "Title mismatch")
	})

	t.Run("returns error for non-existent todo", func(t *testing.T) {
		tl := newTestTodoList()

		_, err := tl.Get("non-existent-id")
		assert.Error(t, err, "Expected error when getting non-existent todo")
	})
}

// TestUpdateTodo verifies updating todo fields with deterministic clock
func TestUpdateTodo(t *testing.T) {
	t.Run("updates all fields successfully", func(t *testing.T) {
		// Use a fixed time for deterministic testing
		startTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
		clock := NewTestClock(startTime)
		tl := newTestTodoList(withClock(clock))

		todo, _ := tl.Add("Original", "", -1)
		originalID := todo.ID
		originalUpdatedAt := todo.UpdatedAt

		// Advance clock deterministically
		clock.Advance(1 * time.Second)

		newTitle := "Updated"
		newDesc := "Updated Description"
		newPriority := 2
		newDue := clock.Now().Add(48 * time.Hour)
		newTags := []string{"updated", "tags"}

		updates := TodoUpdate{
			Title:       &newTitle,
			Description: &newDesc,
			Priority:    &newPriority,
			DueDate:     &newDue,
			Tags:        newTags,
		}

		updated, err := tl.Update(todo.ID, updates)
		assert.NoError(t, err, "Update failed")
		assert.Equal(t, originalID, updated.ID, "ID should not change on update")
		assert.Equal(t, "Updated", updated.Title, "Title not updated")
		assert.Equal(t, "Updated Description", updated.Description, "Description not updated")
		assert.Equal(t, 2, updated.Priority, "Priority not updated")
		assert.True(t, updated.UpdatedAt.After(originalUpdatedAt), "UpdatedAt should be updated")
		assert.Equal(t, startTime.Add(1*time.Second), updated.UpdatedAt, "UpdatedAt should match clock")
	})

	t.Run("returns error for non-existent todo", func(t *testing.T) {
		tl := newTestTodoList()
		newTitle := "New"
		updates := TodoUpdate{Title: &newTitle}

		_, err := tl.Update("non-existent", updates)
		assert.Error(t, err, "Expected error when updating non-existent todo")
	})
}

// TestDeleteTodo verifies deleting todos from the list
func TestDeleteTodo(t *testing.T) {
	t.Run("deletes root-level todo", func(t *testing.T) {
		tl := newTestTodoList()
		todo, _ := tl.Add("To Delete", "", -1)
		id := todo.ID

		err := tl.Delete(id, false)
		assert.NoError(t, err, "Delete failed")
		assert.Len(t, tl.roots, 0, "Expected 0 roots")
		assert.NotContains(t, tl.todos, id, "Todo still exists in map after deletion")
		assertValid(t, tl)
	})

	t.Run("deletes child todo", func(t *testing.T) {
		tl := newTestTodoList()
		parent, _ := tl.Add("Parent", "", -1)
		child, _ := tl.Add("Child", parent.ID, -1)

		err := tl.Delete(child.ID, false)
		assert.NoError(t, err, "Delete failed")
		assert.Len(t, parent.Children, 0, "Expected 0 children")
		assert.NotContains(t, tl.todos, child.ID, "Child still exists in map after deletion")
		assertValid(t, tl)
	})

	t.Run("returns error for non-existent todo", func(t *testing.T) {
		tl := newTestTodoList()

		err := tl.Delete("non-existent", false)
		assert.Error(t, err, "Expected error when deleting non-existent todo")
	})

	t.Run("returns error when todo has children and force is false", func(t *testing.T) {
		tl := newTestTodoList()
		parent, _ := tl.Add("Parent", "", -1)
		tl.Add("Child", parent.ID, -1)

		err := tl.Delete(parent.ID, false)
		assert.Error(t, err, "Expected error when deleting todo with children")
		assert.Contains(t, err.Error(), "children", "Error should mention children")
	})

	t.Run("deletes todo with children when force is true", func(t *testing.T) {
		tl := newTestTodoList()
		parent, _ := tl.Add("Parent", "", -1)
		child, _ := tl.Add("Child", parent.ID, -1)

		err := tl.Delete(parent.ID, true)
		assert.NoError(t, err, "Delete with force should succeed")
		assert.NotContains(t, tl.todos, parent.ID, "Parent should be deleted")
		assert.Contains(t, tl.todos, child.ID, "Child should still exist")
		assert.Equal(t, "", child.ParentID, "Child should be promoted to root (empty ParentID)")
		assert.Contains(t, tl.roots, child, "Child should be in roots")
		assertValid(t, tl)
	})
}

// TestGetRoots verifies retrieving all root todos
func TestGetRoots(t *testing.T) {
	t.Run("retrieves all root todos", func(t *testing.T) {
		tl := newTestTodoList()

		// Add multiple roots
		tl.Add("Root 1", "", -1)
		tl.Add("Root 2", "", -1)
		tl.Add("Root 3", "", -1)

		roots := tl.GetRoots()
		assert.Len(t, roots, 3, "Expected 3 roots")
	})
}

// TestGetChildren verifies retrieving children of a todo
func TestGetChildren(t *testing.T) {
	tl := newTestTodoList()
	parent, _ := tl.Add("Parent", "", -1)
	tl.Add("Child 1", parent.ID, -1)
	tl.Add("Child 2", parent.ID, -1)

	t.Run("retrieves children of a todo", func(t *testing.T) {
		children, err := tl.GetChildren(parent.ID)
		assert.NoError(t, err, "GetChildren failed")
		assert.Len(t, children, 2, "Expected 2 children")
	})

	t.Run("fails to retrieve childrin of a non-existent todo", func(t *testing.T) {
		children, err := tl.GetChildren("non-existent")
		assert.Error(t, err, "Expected error when retrieving children of non-existent todo")
		assert.Nil(t, children, "Expected nil children for non-existent todo")
	})
}

// TestComplete verifies completing a todo
func TestComplete(t *testing.T) {
	t.Run("completes a root todo", func(t *testing.T) {
		tl := newTestTodoList()
		todo, _ := tl.Add("Root Todo", "", -1)

		err := tl.Complete(todo.ID)
		assert.NoError(t, err, "Complete failed")
		assert.True(t, todo.Completed, "Todo should be marked as completed")
		assertValid(t, tl)
	})

	t.Run("completes a child todo", func(t *testing.T) {
		tl := newTestTodoList()
		parent, _ := tl.Add("Parent", "", -1)
		child, _ := tl.Add("Child", parent.ID, -1)

		err := tl.Complete(child.ID)
		assert.NoError(t, err, "Complete failed")
		assert.True(t, child.Completed, "Child should be marked as completed")
		assertValid(t, tl)
	})

	t.Run("returns error for non-existent todo", func(t *testing.T) {
		tl := newTestTodoList()

		err := tl.Complete("non-existent-id")
		assert.Error(t, err, "Expected error when completing non-existent todo")
	})

	t.Run("updates UpdatedAt timestamp", func(t *testing.T) {
		startTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
		clock := NewTestClock(startTime)
		tl := newTestTodoList(withClock(clock))

		todo, _ := tl.Add("Test Todo", "", -1)
		originalUpdatedAt := todo.UpdatedAt

		// Advance clock
		clock.Advance(1 * time.Hour)

		err := tl.Complete(todo.ID)
		assert.NoError(t, err, "Complete failed")
		assert.True(t, todo.UpdatedAt.After(originalUpdatedAt), "UpdatedAt should be updated")
		assert.Equal(t, clock.Now(), todo.UpdatedAt, "UpdatedAt should match current clock time")
		assertValid(t, tl)
	})

	t.Run("idempotent - completing already completed todo", func(t *testing.T) {
		tl := newTestTodoList()
		todo, _ := tl.Add("Test Todo", "", -1)

		// Complete once
		err := tl.Complete(todo.ID)
		assert.NoError(t, err, "First Complete failed")
		assert.True(t, todo.Completed, "Todo should be completed after first call")

		originalUpdatedAt := todo.UpdatedAt

		// Complete again
		err = tl.Complete(todo.ID)
		assert.NoError(t, err, "Second Complete failed")
		assert.True(t, todo.Completed, "Todo should still be completed")
		assert.True(t, todo.UpdatedAt.After(originalUpdatedAt), "UpdatedAt should be updated on second Complete")
		assertValid(t, tl)
	})
}

func TestCompleteSubtree(t *testing.T) {
	t.Run("completes parent and all descendants", func(t *testing.T) {
		tl := newTestTodoList()
		parent, _ := tl.Add("Parent", "", -1)
		child1, _ := tl.Add("Child 1", parent.ID, -1)
		child2, _ := tl.Add("Child 2", parent.ID, -1)
		grandchild, _ := tl.Add("Grandchild", child1.ID, -1)

		err := tl.CompleteSubtree(parent.ID)
		assert.NoError(t, err, "CompleteSubtree should succeed")

		// All todos should be completed
		assert.True(t, parent.Completed, "Parent should be completed")
		assert.True(t, child1.Completed, "Child 1 should be completed")
		assert.True(t, child2.Completed, "Child 2 should be completed")
		assert.True(t, grandchild.Completed, "Grandchild should be completed")
		assertValid(t, tl)
	})

	t.Run("returns error for non-existent todo", func(t *testing.T) {
		tl := newTestTodoList()

		err := tl.CompleteSubtree("non-existent-id")
		assert.Error(t, err, "Expected error when completing non-existent todo subtree")
	})

	t.Run("completes single todo with no children", func(t *testing.T) {
		tl := newTestTodoList()
		todo, _ := tl.Add("Single Todo", "", -1)

		err := tl.CompleteSubtree(todo.ID)
		assert.NoError(t, err, "CompleteSubtree should succeed for single todo")
		assert.True(t, todo.Completed, "Todo should be completed")
		assertValid(t, tl)
	})

	t.Run("updates UpdatedAt for all completed todos", func(t *testing.T) {
		startTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
		clock := NewTestClock(startTime)
		tl := newTestTodoList(withClock(clock))

		parent, _ := tl.Add("Parent", "", -1)
		child, _ := tl.Add("Child", parent.ID, -1)

		originalParentUpdatedAt := parent.UpdatedAt
		originalChildUpdatedAt := child.UpdatedAt

		// Advance clock
		clock.Advance(1 * time.Hour)

		err := tl.CompleteSubtree(parent.ID)
		assert.NoError(t, err, "CompleteSubtree should succeed")

		assert.True(t, parent.UpdatedAt.After(originalParentUpdatedAt), "Parent UpdatedAt should be updated")
		assert.True(t, child.UpdatedAt.After(originalChildUpdatedAt), "Child UpdatedAt should be updated")
		assert.Equal(t, clock.Now(), parent.UpdatedAt, "Parent UpdatedAt should match current clock time")
		assert.Equal(t, clock.Now(), child.UpdatedAt, "Child UpdatedAt should match current clock time")
		assertValid(t, tl)
	})
}
