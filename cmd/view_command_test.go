package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	internalview "text-swap/internal/view"
)

func executeViewCmd(args ...string) (string, string, error) {
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)
	cmd := newViewCmd()

	cmd.SetOut(stdoutBuf)
	cmd.SetErr(stderrBuf)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return stdoutBuf.String(), stderrBuf.String(), err
}

func TestViewCmd_ReadFileError(t *testing.T) {
	_, _, err := executeViewCmd("missing.txt")
	if err == nil {
		t.Fatal("expected read file error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read file") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestViewOptions_Run_SearchMode(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "sample.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	opts := &viewOptions{searchTerm: "he"}
	called := false
	opts.runProgram = func(m tea.Model, _ ...tea.ProgramOption) error {
		called = true
		vm, ok := m.(internalview.Model)
		if !ok {
			t.Fatalf("expected internalview.Model, got %T", m)
		}
		if vm.GetMode() != internalview.ModeSearch {
			t.Fatalf("mode = %v, want ModeSearch", vm.GetMode())
		}
		return nil
	}

	cmd := newViewCmd()
	if err := opts.run(cmd, []string{testFile}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	if !called {
		t.Fatal("runProgram was not called")
	}
}

func TestViewOptions_Run_DiffMode(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "sample.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	opts := &viewOptions{searchTerm: "he", replaceTerm: "HE"}
	opts.runProgram = func(m tea.Model, _ ...tea.ProgramOption) error {
		vm := m.(internalview.Model)
		if vm.GetMode() != internalview.ModeDiff {
			t.Fatalf("mode = %v, want ModeDiff", vm.GetMode())
		}
		return nil
	}

	cmd := newViewCmd()
	if err := opts.run(cmd, []string{testFile}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
}

func TestViewOptions_Run_ProgramError(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "sample.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	opts := &viewOptions{}
	opts.runProgram = func(_ tea.Model, _ ...tea.ProgramOption) error {
		return fmt.Errorf("boom")
	}

	cmd := newViewCmd()
	err := opts.run(cmd, []string{testFile})
	if err == nil {
		t.Fatal("expected program error, got nil")
	}
	if !strings.Contains(err.Error(), "execution error") {
		t.Fatalf("unexpected error: %v", err)
	}
}
