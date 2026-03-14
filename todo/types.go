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

// Option is a function that modifies a TodoList
type Option func(*TodoList)

// WithClock sets the clock for the todo list (for testing)
func WithClock(clock Clock) Option {
	return func(tl *TodoList) {
		tl.clock = clock
	}
}

// NewTodoList creates a new empty todo list
func NewTodoList(options ...Option) *TodoList {
	tl := &TodoList{
		todos:    make(map[string]*Todo),
		roots:    []*Todo{},
		modified: false,
		clock:    RealClock{},
	}
	for _, opt := range options {
		opt(tl)
	}
	return tl
}
