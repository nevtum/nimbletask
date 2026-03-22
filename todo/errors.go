package todo

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidMetadata is returned when metadata is missing or malformed
	ErrInvalidMetadata = errors.New("invalid or missing metadata")
	// ErrMissingID is returned when a todo has no ID
	ErrMissingID = errors.New("todo missing required ID")
	// ErrMissingCreated is returned when a todo has no created timestamp
	ErrMissingCreated = errors.New("todo missing required created timestamp")
	// ErrDuplicateID is returned when two todos have the same ID
	ErrDuplicateID = fmt.Errorf("duplicate ID")
	// ErrInvalidDateFormat is returned when a date cannot be parsed
	ErrInvalidDateFormat = errors.New("invalid date format")
	// ErrCycleDetected is returned when a cycle exists in the hierarchy
	ErrCycleDetected = errors.New("cycle detected in hierarchy")
)

type FileDoesNotExistError struct {
	Err error
}

func (e FileDoesNotExistError) Error() string {
	return fmt.Sprintf("file does not exist: %v", e.Err)
}
