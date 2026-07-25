package progress

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewFileProgressReader(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "input.txt")
	content := "hello\nworld\n"

	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	f, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer func() {
		_ = f.Close()
	}()

	stderr := new(bytes.Buffer)
	reader, finish, err := NewFileProgressReader(stderr, f, "test")
	if err != nil {
		t.Fatalf("NewFileProgressReader() returned error: %v", err)
	}
	if reader == nil {
		t.Fatal("reader is nil")
	}
	if finish == nil {
		t.Fatal("finish function is nil")
	}

	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read from progress reader: %v", err)
	}
	if string(got) != content {
		t.Fatalf("unexpected reader output: got %q want %q", string(got), content)
	}

	if err := finish(); err != nil {
		t.Fatalf("finish() returned error: %v", err)
	}

	if !strings.Contains(stderr.String(), "test") {
		t.Fatalf("progress output did not contain description: %q", stderr.String())
	}
}

func TestNewFileProgressReader_StatError(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "closed-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	name := f.Name()
	if err := f.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}
	if err := os.Remove(name); err != nil {
		t.Fatalf("failed to remove temp file: %v", err)
	}

	_, _, err = NewFileProgressReader(io.Discard, f, "test")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "cannot stat input file") {
		t.Fatalf("unexpected error: %v", err)
	}
}
