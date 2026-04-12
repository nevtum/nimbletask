package todo

import (
	"fmt"
	"strings"
	"testing"
)

func TestTodoList_GetWithPrefix(t *testing.T) {
	// Setup test data using existing helper
	idx := 0
	var fakeGenerateID = func() string {
		idx++
		return fmt.Sprintf("X%02d-Y", idx)
	}
	tl := newTestTodoList(withIDGenerator(fakeGenerateID))

	t1, err := tl.Add("Task 1", "", -1)
	if err != nil {
		t.Fatalf("failed to add task 1: %v", err)
	}
	t2, err := tl.Add("Task 2", "", -1)
	if err != nil {
		t.Fatalf("failed to add task 2: %v", err)
	}

	// TODO: Add a test for ambiguous IDs

	tests := []struct {
		name          string
		prefix        string
		wantErr       bool
		expectedTitle string
		errContains   string
	}{
		{
			name:          "Exact match",
			prefix:        t1.ID,
			wantErr:       false,
			expectedTitle: "Task 1",
		},
		{
			name:          "Unique prefix match",
			prefix:        t2.ID[:3], // Use first 3 chars of t2's ID
			wantErr:       false,
			expectedTitle: "Task 2",
		},
		{
			name:        "No match",
			prefix:      "nonexistent",
			wantErr:     true,
			errContains: "not found",
		},
		{
			name:        "Ambiguous prefix match",
			prefix:      t1.ID[:2], // Use first 2 chars of t1's ID
			wantErr:     true,
			errContains: "ambiguous ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tl.Get(tt.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
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
