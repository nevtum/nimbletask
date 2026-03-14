package todo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }

// roundTrip saves tl1 to a temp file, loads it back, and returns the loaded list
func roundTrip(t *testing.T, tl1 *TodoList) *TodoList {
	t.Helper()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "roundtrip.md")

	err := tl1.Save(filePath)
	assert.NoError(t, err, "Save failed during round-trip")

	tl2, err := LoadTodoList(filePath)
	assert.NoError(t, err, "Load failed during round-trip")

	return tl2
}

// assertEqualTodos deep-compares two Todo structs (including children recursively)
func assertEqualTodos(t *testing.T, t1, t2 *Todo) {
	t.Helper()
	assert.Equal(t, t1.ID, t2.ID, "ID mismatch")
	assert.Equal(t, t1.Title, t2.Title, "Title mismatch")
	assert.Equal(t, t1.Description, t2.Description, "Description mismatch")
	assert.Equal(t, t1.Completed, t2.Completed, "Completed mismatch")
	assert.Equal(t, t1.Priority, t2.Priority, "Priority mismatch")
	assert.Equal(t, t1.CreatedAt, t2.CreatedAt, "CreatedAt mismatch")
	assert.Equal(t, t1.UpdatedAt, t2.UpdatedAt, "UpdatedAt mismatch")
	assert.Equal(t, t1.Tags, t2.Tags, "Tags mismatch")

	// Compare DueDate (both nil or equal)
	if t1.DueDate != nil && t2.DueDate != nil {
		assert.Equal(t, *t1.DueDate, *t2.DueDate, "DueDate mismatch")
	} else {
		assert.Nil(t, t1.DueDate, "t1.DueDate should be nil")
		assert.Nil(t, t2.DueDate, "t2.DueDate should be nil")
	}

	// Compare children recursively
	assert.Equal(t, len(t1.Children), len(t2.Children), "Children count mismatch for todo %s", t1.ID)
	for i, child1 := range t1.Children {
		child2 := t2.Children[i]
		assertEqualTodos(t, child1, child2)
	}
}

// assertEqualTodoLists deep-compares two TodoList structs
func assertEqualTodoLists(t *testing.T, tl1, tl2 *TodoList) {
	t.Helper()

	// Compare todos map
	assert.Equal(t, len(tl1.todos), len(tl2.todos), "Todo count mismatch")
	for id, todo1 := range tl1.todos {
		todo2, exists := tl2.todos[id]
		assert.True(t, exists, "Missing todo ID %s in loaded list", id)
		assertEqualTodos(t, todo1, todo2)
	}

	// Compare roots order
	assert.Equal(t, len(tl1.roots), len(tl2.roots), "Root count mismatch")
	for i, root1 := range tl1.roots {
		root2 := tl2.roots[i]
		assert.Equal(t, root1.ID, root2.ID, "Root order mismatch at index %d", i)
	}

	// Validate loaded list structure
	assertValid(t, tl2)
}

// saveToString saves tl to a string (for format testing without file I/O)
func saveToString(t *testing.T, tl *TodoList) string {
	t.Helper()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "temp.md")

	err := tl.Save(filePath)
	assert.NoError(t, err)

	content, err := os.ReadFile(filePath)
	assert.NoError(t, err)
	return string(content)
}

// ============================================================================
// SAVE/LOAD ROUND-TRIP TESTS (Data Fidelity)
// ============================================================================

func TestSaveLoadRoundTrip(t *testing.T) {
	t.Run("simple todo with all fields", func(t *testing.T) {
		startTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
		clock := NewTestClock(startTime)
		tl1 := NewTodoList(withClock(clock))

		todo, _ := tl1.Add("Complete Task", "", -1)
		due := startTime.Add(24 * time.Hour)
		tl1.Update(todo.ID, TodoUpdate{
			Description: strPtr("Full description with details"),
			Priority:    intPtr(3),
			DueDate:     &due,
			Tags:        []string{"work", "urgent"},
		})
		tl1.Complete(todo.ID)

		tl2 := roundTrip(t, tl1)
		assertEqualTodoLists(t, tl1, tl2)
	})

	t.Run("deep hierarchy with mixed field combinations", func(t *testing.T) {
		startTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
		clock := NewTestClock(startTime)
		tl1 := NewTodoList(withClock(clock))

		// Build a complex tree
		root1, _ := tl1.Add("Root 1", "", -1)
		root2, _ := tl1.Add("Root 2", "", -1)

		child1, _ := tl1.Add("Child 1.1", root1.ID, -1)
		child2, _ := tl1.Add("Child 1.2", root1.ID, -1)
		_, _ = tl1.Add("Child 2.1", root2.ID, -1)

		grand1, _ := tl1.Add("Grand 1.1.1", child1.ID, -1)
		grand2, _ := tl1.Add("Grand 1.1.2", child1.ID, -1)
		_, _ = tl1.Add("Grand 1.2.1", child2.ID, -1)

		// Mix of field updates
		due1 := startTime.Add(48 * time.Hour)
		_ = startTime.Add(72 * time.Hour)
		tl1.Update(child1.ID, TodoUpdate{Priority: intPtr(2), DueDate: &due1})
		tl1.Update(child2.ID, TodoUpdate{Priority: intPtr(1)})
		tl1.Update(grand1.ID, TodoUpdate{Tags: []string{"important"}})
		tl1.Complete(grand2.ID)
		tl1.Update(root2.ID, TodoUpdate{Description: strPtr("Root 2 description")})

		tl2 := roundTrip(t, tl1)
		assertEqualTodoLists(t, tl1, tl2)
	})

	t.Run("order preservation with specific positions", func(t *testing.T) {
		startTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
		clock := NewTestClock(startTime)
		tl1 := NewTodoList(withClock(clock))

		parent, _ := tl1.Add("Parent", "", -1)

		// Add in specific order: First (pos 0), Third (pos -1), Second (pos 1)
		first, _ := tl1.Add("First", parent.ID, 0)
		third, _ := tl1.Add("Third", parent.ID, -1)
		second, _ := tl1.Add("Second", parent.ID, 1)

		tl2 := roundTrip(t, tl1)

		// Verify order is preserved
		assert.Equal(t, first.ID, parent.Children[0].ID)
		assert.Equal(t, second.ID, parent.Children[1].ID)
		assert.Equal(t, third.ID, parent.Children[2].ID)

		// Verify deep equality
		assertEqualTodoLists(t, tl1, tl2)
	})

	t.Run("edge cases: empty strings, zero values, special characters", func(t *testing.T) {
		startTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
		clock := NewTestClock(startTime)
		tl1 := NewTodoList(withClock(clock))

		// Various edge cases
		t1, _ := tl1.Add("Minimal", "", -1) // no optional fields
		t2, _ := tl1.Add("With Bracket [Test]", "", -1)
		t3, _ := tl1.Add("With Asterisk * Bold", "", -1)
		_, _ = tl1.Add("With Backtick `code`", "", -1)
		_, _ = tl1.Add("With Pipe | and Ampersand &", "", -1)

		// Zero priority (should be omitted)
		tl1.Update(t1.ID, TodoUpdate{Priority: intPtr(0)})

		// Empty description (explicitly set to empty)
		tl1.Update(t2.ID, TodoUpdate{Description: strPtr("")})

		// Multi-line with special chars
		tl1.Update(t3.ID, TodoUpdate{
			Description: strPtr("Line 1\nLine 2 with *stars*\nLine 3 with `code`"),
		})

		tl2 := roundTrip(t, tl1)
		assertEqualTodoLists(t, tl1, tl2)
	})

	t.Run("modified flag is false after load", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "mod-flag.md")

		tl1 := NewTodoList()
		tl1.Add("Task", "", -1)
		tl1.Save(filePath)

		tl2, _ := LoadTodoList(filePath)
		assert.False(t, tl2.IsModified())
	})
}

// ============================================================================
// SAVE FORMAT TESTS (File Content Verification)
// ============================================================================

func TestSaveFormat(t *testing.T) {
	t.Run("empty list produces empty file", func(t *testing.T) {
		tl := NewTodoList()
		content := saveToString(t, tl)
		assert.Empty(t, strings.TrimSpace(content))
	})

	t.Run("omits priority when zero", func(t *testing.T) {
		tl := NewTodoList()
		tl.Add("Task", "", -1) // default priority = 0

		content := saveToString(t, tl)
		assert.NotContains(t, content, "priority:")
	})

	t.Run("includes priority when non-zero", func(t *testing.T) {
		tl := NewTodoList()
		todo, _ := tl.Add("Task", "", -1)
		tl.Update(todo.ID, TodoUpdate{Priority: intPtr(5)})

		content := saveToString(t, tl)
		assert.Contains(t, content, "priority:5")
	})

	t.Run("omits due when nil", func(t *testing.T) {
		tl := NewTodoList()
		todo, _ := tl.Add("Task", "", -1)
		tl.Update(todo.ID, TodoUpdate{Priority: intPtr(1)}) // no DueDate

		content := saveToString(t, tl)
		assert.NotContains(t, content, "due:")
	})

	t.Run("includes due when set", func(t *testing.T) {
		startTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
		clock := NewTestClock(startTime)
		tl := NewTodoList(withClock(clock))

		todo, _ := tl.Add("Task", "", -1)
		due := startTime.Add(24 * time.Hour)
		tl.Update(todo.ID, TodoUpdate{DueDate: &due})

		content := saveToString(t, tl)
		assert.Contains(t, content, "due:2024-01-16T10:00:00Z")
	})

	t.Run("omits completed date when not completed", func(t *testing.T) {
		tl := NewTodoList()
		tl.Add("Task", "", -1)

		content := saveToString(t, tl)
		assert.NotContains(t, content, "completed:")
	})

	t.Run("includes completed date when completed", func(t *testing.T) {
		startTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
		clock := NewTestClock(startTime)
		tl := NewTodoList(withClock(clock))

		todo, _ := tl.Add("Task", "", -1)
		tl.Complete(todo.ID)

		content := saveToString(t, tl)
		assert.Contains(t, content, "completed:2024-01-15T10:00:00Z")
	})

	t.Run("formats tags as comma-separated", func(t *testing.T) {
		tl := NewTodoList()
		todo, _ := tl.Add("Task", "", -1)
		tl.Update(todo.ID, TodoUpdate{Tags: []string{"tag1", "tag2", "tag3"}})

		content := saveToString(t, tl)
		assert.Contains(t, content, "tags:tag1,tag2,tag3")
	})

	t.Run("handles multi-line descriptions with 2-space indent", func(t *testing.T) {
		tl := NewTodoList()
		todo, _ := tl.Add("Task", "", -1)
		tl.Update(todo.ID, TodoUpdate{Description: strPtr("Line 1\nLine 2\nLine 3")})

		content := saveToString(t, tl)
		lines := strings.Split(strings.TrimSpace(content), "\n")
		assert.Len(t, lines, 4)
		assert.Contains(t, lines[1], "  Line 1")
		assert.Contains(t, lines[2], "  Line 2")
		assert.Contains(t, lines[3], "  Line 3")
	})

	t.Run("indents children by 2 spaces per level", func(t *testing.T) {
		tl := NewTodoList()
		parent, _ := tl.Add("Parent", "", -1)
		child, _ := tl.Add("Child", parent.ID, -1)
		_, _ = tl.Add("Grandchild", child.ID, -1)

		content := saveToString(t, tl)
		lines := strings.Split(strings.TrimSpace(content), "\n")

		// Parent: no indent
		assert.True(t, strings.HasPrefix(lines[0], "- [ ] "))
		// Child: 2 spaces
		assert.True(t, strings.HasPrefix(lines[1], "  - [ ] "))
		// Grandchild: 4 spaces
		assert.True(t, strings.HasPrefix(lines[2], "    - [ ] "))
	})

	t.Run("preserves order of multiple roots", func(t *testing.T) {
		tl := NewTodoList()
		tl.Add("First", "", -1)
		tl.Add("Second", "", -1)
		tl.Add("Third", "", -1)

		content := saveToString(t, tl)
		lines := strings.Split(strings.TrimSpace(content), "\n")
		assert.Contains(t, lines[0], "First")
		assert.Contains(t, lines[1], "Second")
		assert.Contains(t, lines[2], "Third")
	})
}

// ============================================================================
// LOAD PARSING TESTS (Happy Path)
// ============================================================================

func TestLoad(t *testing.T) {
	t.Run("empty file returns empty TodoList", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "empty.md")
		os.WriteFile(filePath, []byte{}, 0644)

		tl, err := LoadTodoList(filePath)
		assert.NoError(t, err)
		assert.NotNil(t, tl)
		assertValid(t, tl)
		assert.Len(t, tl.todos, 0)
		assert.Len(t, tl.roots, 0)
	})

	t.Run("single root todo with basic fields", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "single.md")

		markdown := `- [ ] <!-- id:abc123|parent:|created:2024-01-15T10:00:00Z --> Task Title
  Description
`
		os.WriteFile(filePath, []byte(markdown), 0644)

		tl, err := LoadTodoList(filePath)
		assert.NoError(t, err)
		assert.Len(t, tl.todos, 1)
		assert.Len(t, tl.roots, 1)

		todo := tl.roots[0]
		assert.Equal(t, "abc123", todo.ID)
		assert.Equal(t, "Task Title", todo.Title)
		assert.False(t, todo.Completed)
		assert.Equal(t, "", todo.ParentID)
		assert.Equal(t, "Description", todo.Description)
	})

	t.Run("nested hierarchy with 3 levels", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "nested.md")

		markdown := `- [ ] <!-- id:root|parent:|priority:0|created:2024-01-15T10:00:00Z --> Root
  - [ ] <!-- id:child|parent:root|created:2024-01-15T10:00:00Z --> Child
    - [ ] <!-- id:grandchild|parent:child|created:2024-01-15T10:00:00Z --> Grandchild
`
		os.WriteFile(filePath, []byte(markdown), 0644)

		tl, err := LoadTodoList(filePath)
		assert.NoError(t, err)
		assert.Len(t, tl.todos, 3)
		assert.Len(t, tl.roots, 1)

		root := tl.roots[0]
		assert.Equal(t, "root", root.ID)
		assert.Len(t, root.Children, 1)
		assert.Equal(t, "child", root.Children[0].ID)
		assert.Len(t, root.Children[0].Children, 1)
		assert.Equal(t, "grandchild", root.Children[0].Children[0].ID)
	})

	t.Run("all optional fields", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "full.md")

		markdown := `- [x] <!-- id:full|parent:|priority:3|created:2024-01-15T10:00:00Z|completed:2024-01-16T14:30:00Z|due:2024-02-01T00:00:00Z|tags:tag1,tag2,tag3 --> Complete Task
  Full description with details
`
		os.WriteFile(filePath, []byte(markdown), 0644)

		tl, err := LoadTodoList(filePath)
		assert.NoError(t, err)
		assert.Len(t, tl.todos, 1)

		todo := tl.roots[0]
		assert.True(t, todo.Completed)
		assert.Equal(t, 3, todo.Priority)
		assert.NotNil(t, todo.DueDate)
		assert.Equal(t, "2024-02-01T00:00:00Z", (*todo.DueDate).Format(time.RFC3339))
		assert.Len(t, todo.Tags, 3)
		assert.Contains(t, todo.Tags, "tag1")
		assert.Contains(t, todo.Tags, "tag2")
		assert.Contains(t, todo.Tags, "tag3")
	})

	t.Run("preserves timestamps from file", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "timestamps.md")

		markdown := `- [ ] <!-- id:ts|parent:|created:2020-05-01T12:00:00Z|updated:2020-05-02T13:00:00Z --> Timestamped Task
`
		os.WriteFile(filePath, []byte(markdown), 0644)

		tl, err := LoadTodoList(filePath)
		assert.NoError(t, err)

		todo := tl.roots[0]
		expectedCreated := time.Date(2020, 5, 1, 12, 0, 0, 0, time.UTC)
		expectedUpdated := time.Date(2020, 5, 2, 13, 0, 0, 0, time.UTC)
		assert.Equal(t, expectedCreated, todo.CreatedAt)
		assert.Equal(t, expectedUpdated, todo.UpdatedAt)
	})

	t.Run("multiple roots maintain order", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "order.md")

		markdown := `- [ ] <!-- id:root1|parent:|created:2024-01-15T10:00:00Z --> First
- [ ] <!-- id:root2|parent:|created:2024-01-15T10:00:00Z --> Second
- [ ] <!-- id:root3|parent:|created:2024-01-15T10:00:00Z --> Third
`
		os.WriteFile(filePath, []byte(markdown), 0644)

		tl, err := LoadTodoList(filePath)
		assert.NoError(t, err)
		assert.Len(t, tl.roots, 3)
		assert.Equal(t, "root1", tl.roots[0].ID)
		assert.Equal(t, "root2", tl.roots[1].ID)
		assert.Equal(t, "root3", tl.roots[2].ID)
	})

	t.Run("multi-line description", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "desc.md")

		markdown := `- [ ] <!-- id:desc|parent:|created:2024-01-15T10:00:00Z --> Task
  First line
  Second line
  Third line
`
		os.WriteFile(filePath, []byte(markdown), 0644)

		tl, err := LoadTodoList(filePath)
		assert.NoError(t, err)

		todo := tl.roots[0]
		expectedDesc := "First line\nSecond line\nThird line"
		assert.Equal(t, expectedDesc, todo.Description)
	})

	t.Run("empty description", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "no-desc.md")

		markdown := `- [ ] <!-- id:nodesc|parent:|created:2024-01-15T10:00:00Z --> Task
`
		os.WriteFile(filePath, []byte(markdown), 0644)

		tl, err := LoadTodoList(filePath)
		assert.NoError(t, err)

		todo := tl.roots[0]
		assert.Empty(t, todo.Description)
	})
}

// ============================================================================
// ERROR HANDLING TESTS
// ============================================================================

func TestLoadErrors(t *testing.T) {
	t.Run("missing metadata", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "bad.md")

		markdown := `- [ ] Task without metadata
`
		os.WriteFile(filePath, []byte(markdown), 0644)

		_, err := LoadTodoList(filePath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "metadata")
	})

	t.Run("invalid ID", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "bad-id.md")

		markdown := `- [ ] <!-- id:|parent:|created:2024-01-15T10:00:00Z --> Task
`
		os.WriteFile(filePath, []byte(markdown), 0644)

		_, err := LoadTodoList(filePath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ID")
	})

	t.Run("missing required created timestamp", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "bad-time.md")

		markdown := `- [ ] <!-- id:abc|parent:| --> Task
`
		os.WriteFile(filePath, []byte(markdown), 0644)

		_, err := LoadTodoList(filePath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "created")
	})

	t.Run("invalid date format", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "bad-date.md")

		markdown := `- [ ] <!-- id:abc|parent:|created:not-a-date|completed:2024-01-15T10:00:00Z --> Task
`
		os.WriteFile(filePath, []byte(markdown), 0644)

		_, err := LoadTodoList(filePath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "date")
	})

	t.Run("duplicate ID", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "dup.md")

		markdown := `- [ ] <!-- id:dup|parent:|created:2024-01-15T10:00:00Z --> Task 1
- [ ] <!-- id:dup|parent:|created:2024-01-15T10:00:00Z --> Task 2
`
		os.WriteFile(filePath, []byte(markdown), 0644)

		_, err := LoadTodoList(filePath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate")
		assert.Contains(t, err.Error(), "dup")
	})

	t.Run("malformed comment syntax", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "syntax.md")

		markdown := `- [ ] <!-- id:abc parent:|created:2024-01-15T10:00:00Z --> Task
`
		os.WriteFile(filePath, []byte(markdown), 0644)

		_, err := LoadTodoList(filePath)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "format")
	})

}

// ============================================================================
// MODIFIED FLAG TESTS
// ============================================================================

func TestModifiedFlag(t *testing.T) {
	t.Run("modified after add", func(t *testing.T) {
		tl := NewTodoList()
		assert.False(t, tl.IsModified())
		tl.Add("Task", "", -1)
		assert.True(t, tl.IsModified())
	})

	t.Run("modified after update", func(t *testing.T) {
		tl := NewTodoList()
		todo, _ := tl.Add("Task", "", -1)
		tl.modified = false
		tl.Update(todo.ID, TodoUpdate{Title: strPtr("New")})
		assert.True(t, tl.IsModified())
	})

	t.Run("modified after delete", func(t *testing.T) {
		tl := NewTodoList()
		todo, _ := tl.Add("Task", "", -1)
		tl.modified = false
		tl.Delete(todo.ID)
		assert.True(t, tl.IsModified())
	})

	t.Run("modified after move", func(t *testing.T) {
		tl := NewTodoList()
		parent, _ := tl.Add("Parent", "", -1)
		child, _ := tl.Add("Child", "", -1)
		tl.modified = false
		tl.Move(child.ID, parent.ID, -1)
		assert.True(t, tl.IsModified())
	})

	t.Run("not modified after save", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "mod-save.md")

		tl := NewTodoList()
		tl.Add("Task", "", -1)
		assert.True(t, tl.IsModified())

		err := tl.Save(filePath)
		assert.NoError(t, err)
		assert.False(t, tl.IsModified())
	})

	t.Run("not modified after load", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "mod-load.md")

		tl1 := NewTodoList()
		tl1.Add("Task", "", -1)
		tl1.Save(filePath)

		tl2, _ := LoadTodoList(filePath)
		assert.False(t, tl2.IsModified())
	})
}

// ============================================================================
// NON-EXISTENT FILE & DIRECTORY TESTS
// ============================================================================

func TestLoadNonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistent := filepath.Join(tmpDir, "does-not-exist.md")

	tl, err := LoadTodoList(nonExistent)
	assert.NoError(t, err)
	assert.NotNil(t, tl)
	assertValid(t, tl)
	assert.Len(t, tl.todos, 0)
	assert.Len(t, tl.roots, 0)
}

func TestSaveToNonExistentDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "newdir")
	filePath := filepath.Join(subDir, "test.md")

	tl := NewTodoList()
	tl.Add("Task", "", -1)

	err := tl.Save(filePath)
	assert.NoError(t, err)

	_, err = os.Stat(filePath)
	assert.NoError(t, err)
}

// ============================================================================
// ATOMIC WRITE TESTS
// ============================================================================

func TestSaveAtomic(t *testing.T) {
	t.Run("creates file atomically", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "atomic.md")

		tl := NewTodoList()
		tl.Add("Task", "", -1)

		err := tl.Save(filePath)
		assert.NoError(t, err)

		_, err = os.Stat(filePath)
		assert.NoError(t, err, "File should exist after save")
	})

	t.Run("overwrites existing file atomically", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "overwrite.md")

		// Create initial file
		os.WriteFile(filePath, []byte("old content"), 0644)

		tl := NewTodoList()
		tl.Add("New Task", "", -1)

		err := tl.Save(filePath)
		assert.NoError(t, err)

		content, _ := os.ReadFile(filePath)
		assert.Contains(t, string(content), "New Task")
		assert.NotContains(t, string(content), "old content")
	})
}
