package todo

import (
	"strings"
	"testing"
)

func TestTodoList_GetWithPrefix(t *testing.T) {
	// Setup test data using existing helper
	tl := newTestTodoList()

	t1, err := tl.Add("Task 1", "", -1)
	if err != nil {
		t.Fatalf("failed to add task 1: %v", err)
	}
	_, err = tl.Add("Task 2", "", -1)
	if err != nil {
		t.Fatalf("failed to add task 2: %v", err)
	}
	t3, err := tl.Add("Task 3", "", -1)
	if err != nil {
		t.Fatalf("failed to add task 3: %v", err)
	}

	// TODO: Add a test for ambiguous IDs

	tests := []struct {
		name          string
		query         string
		wantErr       bool
		expectedTitle string
		errContains   string
	}{
		{
			name:          "Exact match",
			query:         t1.ID,
			wantErr:       false,
			expectedTitle: "Task 1",
		},
		{
			name:          "Unique prefix match",
			query:         t3.ID[:3], // Use first 3 chars of t3's ID
			wantErr:       false,
			expectedTitle: "Task 3",
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
			got, err := tl.Get(tt.query)
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
