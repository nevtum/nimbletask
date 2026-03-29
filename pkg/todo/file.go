package todo

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type AbstractFile interface {
	Load() (string, error)
	Save(string) error
}

type File struct {
	path string
}

func NewFile(path string) *File {
	return &File{path: path}
}

func (f *File) Load() (string, error) {
	file, err := pathToReader(f.path)
	if err != nil {
		return "", err
	}

	// Read all content from the reader
	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

func (f *File) Save(content string) error {
	w, err := pathToWriter(f.path)
	if err != nil {
		return fmt.Errorf("failed to establish writer: %w", err)
	}
	writer := bufio.NewWriter(w)

	if _, err := writer.WriteString(content); err != nil {
		return fmt.Errorf("failed to write to buffer: %w", err)
	}

	return writer.Flush()
}

func pathToReader(path string) (io.Reader, error) {
	// If file doesn't exist, return empty list (per test behavior)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, FileDoesNotExist{
			Err: err,
		}
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}

func pathToWriter(path string) (io.Writer, error) {
	dir := filepath.Dir(path)
	if dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	return f, nil
}

// FakeFile is a fake implementation of File for testing purposes.
type FakeFile struct {
	content string
}

func NewFakeFile() *FakeFile {
	return &FakeFile{}
}

func (f *FakeFile) Load() (string, error) {
	return f.content, nil
}

func (f *FakeFile) Save(content string) error {
	f.content = content
	return nil
}
