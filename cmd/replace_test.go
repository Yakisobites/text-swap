package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper function to execute the replace command in isolatedly
func executeReplaceCmd(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	cmd := newReplaceCmd()

	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return buf.String(), err
}

func TestReplaceCmd_MissingRequiredFlags(t *testing.T) {
	// Test missing required flags (--file or --target)
	_, err := executeReplaceCmd("-f", "input.txt")
	if err == nil {
		t.Errorf("Expected error for missing required flag --target, got nil")
	}

	_, err = executeReplaceCmd("-t", "hello")
	if err == nil {
		t.Errorf("Expected error for missing required flag --file, got nil")
	}
}

func TestReplaceCmd_Stdout(t *testing.T) {
	// Setup temporary test file
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.txt")
	content := "Hello World\nHello Go\n"
	if err := os.WriteFile(inputFile, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Test replacing text and outputting to stdout (captured by buf)
	out, err := executeReplaceCmd("-f", inputFile, "-t", "Hello", "-r", "Hi")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := "Hi World\nHi Go\n"
	if out != expected {
		t.Errorf("Expected stdout %q, got %q", expected, out)
	}
}

func TestReplaceCmd_OutFlag(t *testing.T) {
	// Setup temporary test files
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.txt")
	outputFile := filepath.Join(tmpDir, "output.txt")

	content := "apple banana apple"
	if err := os.WriteFile(inputFile, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Test replacing text with --out (-o) flag
	out, err := executeReplaceCmd("-f", inputFile, "-t", "apple", "-r", "orange", "-o", outputFile)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify command message on stdout
	expectedMsg := "Replaced [apple] -> [orange] (2 occurrences)"
	if !strings.Contains(out, expectedMsg) {
		t.Errorf("Expected message containing %q, got %q", expectedMsg, out)
	}

	// Verify output file content
	savedContent, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	expectedContent := "orange banana orange"
	if string(savedContent) != expectedContent {
		t.Errorf("Expected file content %q, got %q", expectedContent, string(savedContent))
	}
}

func TestReplaceCmd_IgnoreCaseFlag(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.txt")
	if err := os.WriteFile(inputFile, []byte("Go GO go"), 0o644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Test case-insensitive flag (-i)
	out, err := executeReplaceCmd("-f", inputFile, "-t", "go", "-r", "Rust", "-i")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := "Rust Rust Rust"
	if out != expected {
		t.Errorf("Expected stdout %q, got %q", expected, out)
	}
}

func TestReplaceCmd_SameInputAndOutputFileError(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.txt")
	_ = os.WriteFile(inputFile, []byte("test"), 0o644)

	// Test passing the same path to --file and --out
	_, err := executeReplaceCmd("-f", inputFile, "-t", "test", "-r", "pass", "-o", inputFile)
	if err == nil {
		t.Errorf("Expected error when output path equals input path, got nil")
	}
}

func TestReplaceCmd_PreservesLineEndingsAndFinalNewline_WithOutFlag(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		target     string
		repl       string
		wantOutput string
	}{
		{
			name:       "preserve CRLF and final newline",
			input:      "foo\r\nbar\r\n",
			target:     "bar",
			repl:       "baz",
			wantOutput: "foo\r\nbaz\r\n",
		},
		{
			name:       "preserve no final newline",
			input:      "foo\nbar",
			target:     "foo",
			repl:       "qux",
			wantOutput: "qux\nbar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			inputFile := filepath.Join(tmpDir, "input.txt")
			outputFile := filepath.Join(tmpDir, "output.txt")

			if err := os.WriteFile(inputFile, []byte(tt.input), 0o644); err != nil {
				t.Fatalf("Failed to create temp input file: %v", err)
			}

			_, err := executeReplaceCmd("-f", inputFile, "-t", tt.target, "-r", tt.repl, "-o", outputFile)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			got, err := os.ReadFile(outputFile)
			if err != nil {
				t.Fatalf("Failed to read output file: %v", err)
			}

			if string(got) != tt.wantOutput {
				t.Errorf("Expected output file content %q, got %q", tt.wantOutput, string(got))
			}
		})
	}
}
