package todo

import (
	"strings"
	"testing"
)

func TestTodoList_GetWithPrefix(t *testing.T) {
	// Setup test data using existing helper
	tl := newTestTodoList()

	// Setup test data
	todos := []*Todo{
		{ID: "V1StGXR8_Z5jd", Title: "Task 1"},
		{ID: "V1StGXR8_ABCDE", Title: "Task 2"}, // Shares prefix with Task 1
		{ID: "wH9mK2pL4nQ7", Title: "Task 3"},
	}

	for _, todo := range todos {
		tl.todos[todo.ID] = todo
		tl.roots = append(tl.roots, todo)
	}

	tests := []struct {
		name          string
		query         string
		wantErr       bool
		expectedTitle string
		errContains   string
	}{
		{
			name:          "Exact match",
			query:         "V1StGXR8_Z5jd",
			wantErr:       false,
			expectedTitle: "Task 1",
		},
		{
			name:          "Unique prefix match",
			query:         "wH9mK",
			wantErr:       false,
			expectedTitle: "Task 3",
		},
		{
			name:        "Ambiguous prefix",
			query:       "V1St",
			wantErr:     true,
			errContains: "ambiguous",
		},
		{
			name:        "No match",
			query:       "nonexistent",
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: The current implementation of Get in types.go (which is just a placeholder in the spec)
			// will likely not support prefix matching yet, causing this to fail as requested.
			got, err := tl.Get(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				// We expect an error, check if it's the right kind of error
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Get() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}
			if got.Title != tt.expectedTitle {
				t.Errorf("Get() got title = %v, want %v", got.Title, tt.expectedTitle)
			}
		})
	}
}
