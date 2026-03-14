package todo

import (
	"time"
)

// Clock provides time functionality for testability
type Clock interface {
	Now() time.Time
}

// RealClock uses actual system time
type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }

// Todo represents a single task in the hierarchical todo list
type Todo struct {
	ID          string
	Title       string
	Description string
	Completed   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ParentID    string
	Children    []*Todo
	Priority    int
	DueDate     *time.Time
	Tags        []string
}

// TodoList manages a collection of todos with O(1) lookup
type TodoList struct {
	todos    map[string]*Todo
	roots    []*Todo
	modified bool
	clock    Clock
}

// TodoUpdate represents update fields for a todo
type TodoUpdate struct {
	Title       *string
	Description *string
	Priority    *int
	DueDate     *time.Time
	Tags        []string
}

// NewTodoList creates a new empty todo list
func NewTodoList() *TodoList {
	return &TodoList{
		todos:    make(map[string]*Todo),
		roots:    []*Todo{},
		modified: false,
		clock:    RealClock{},
	}
}

// NewTodoListWithClock creates a todo list with a custom clock (for testing)
func NewTodoListWithClock(clock Clock) *TodoList {
	return &TodoList{
		todos:    make(map[string]*Todo),
		roots:    []*Todo{},
		modified: false,
		clock:    clock,
	}
}
