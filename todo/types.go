package todo

import (
	"time"
)

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
	todos map[string]*Todo
	roots []*Todo
	clock Clock
}

// TodoUpdate represents update fields for a todo
type TodoUpdate struct {
	Title       *string
	Description *string
	Priority    *int
	DueDate     *time.Time
	Tags        []string
}

// Option is a function that modifies a TodoList
type Option func(*TodoList)

// NewTodoList creates a new empty todo list
func NewTodoList(options ...Option) *TodoList {
	tl := &TodoList{
		todos: make(map[string]*Todo),
		roots: []*Todo{},
		clock: RealClock{},
	}
	for _, opt := range options {
		opt(tl)
	}
	return tl
}
