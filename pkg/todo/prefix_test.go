package todo

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTodoList_GetWithPrefixSuccess(t *testing.T) {
	idx := 0
	var fakeGenerateID = func() string {
		idx++
		return fmt.Sprintf("X%02d-Y", idx)
	}
	tl := newTestTodoList(withIDGenerator(fakeGenerateID))

	t1, err := tl.Add("Task 1", "", -1)
	assert.NoError(t, err, "failed to add task 1")

	t2, err := tl.Add("Task 2", "", -1)
	assert.NoError(t, err, "failed to add task 2")

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tl.Get(tt.prefix)
			assert.NoError(t, err, "expected no error but got an error")
			assert.Equal(t, tt.expectedTitle, got.Title)
		})
	}
}

func TestTodoList_GetWithPrefixFailed(t *testing.T) {
	idx := 0
	var fakeGenerateID = func() string {
		idx++
		return fmt.Sprintf("X%02d-Y", idx)
	}
	tl := newTestTodoList(withIDGenerator(fakeGenerateID))

	t1, err := tl.Add("Task 1", "", -1)
	assert.NoError(t, err, "failed to add task 1")

	_, err = tl.Add("Task 2", "", -1)
	assert.NoError(t, err, "failed to add task 2")

	tests := []struct {
		name          string
		prefix        string
		wantErr       bool
		expectedTitle string
		errContains   string
	}{
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
			_, err := tl.Get(tt.prefix)
			assert.Error(t, err)
			assert.ErrorContains(t, err, tt.errContains)
		})
	}
}

func TestTodoList_GetMinIDLength(t *testing.T) {
	idx := 0

	tests := []struct {
		name        string
		generator   func() string
		expectedLen int
	}{
		{
			name: "prefix length 7",
			generator: func() string {
				idx++
				return fmt.Sprintf("ABCDE%02d-XYZ", idx)
			},
			expectedLen: 7,
		},
		{
			name: "prefix length MIN_ID_LENGTH",
			generator: func() string {
				idx++
				return fmt.Sprintf("%d-XYZ", idx)
			},
			expectedLen: MIN_ID_LENGTH,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tl := newTestTodoList(withIDGenerator(tt.generator))

			_, err := tl.Add("Task 1", "", -1)
			assert.NoError(t, err, "failed to add task 1")

			_, err = tl.Add("Task 2", "", -1)
			assert.NoError(t, err, "failed to add task 2")

			length := tl.GetMinIDLength()
			assert.Equal(t, tt.expectedLen, length, "expected min ID length to be %d", tt.expectedLen)
		})
	}
}
