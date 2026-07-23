package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper function to execute the search command in isolation.
func executeSearchCmd(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	cmd := newSearchCmd()

	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return buf.String(), err
}

func TestSearchCmd(t *testing.T) {
	// 1. Create a temporary test file (t.TempDir() will be automatically removed after the test)
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "sample.txt")
	fileContent := "Hello World\nhello Go\nHELLO Cobra"

	if err := os.WriteFile(testFile, []byte(fileContent), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 2. Define table-driven test cases
	tests := []struct {
		name         string
		args         []string
		wantCountStr string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "Normal case: Case-sensitive search (1 match)",
			args:         []string{"search", "-f", testFile, "-t", "hello"},
			wantCountStr: "Count of [hello]: 1",
			wantErr:      false,
		},
		{
			name:         "Normal case: Case-insensitive search (-i, 3 matches)",
			args:         []string{"search", "-f", testFile, "-t", "hello", "-i"},
			wantCountStr: "Count of [hello]: 3",
			wantErr:      false,
		},
		{
			name:        "Error case: Nonexistent file path",
			args:        []string{"search", "-f", filepath.Join(tmpDir, "not_exist.txt"), "-t", "hello"},
			wantErr:     true,
			errContains: "cannot open",
		},
		{
			name:    "Error case: Missing required flag (--target)",
			args:    []string{"search", "-f", testFile},
			wantErr: true,
		},
		{
			name:         "Edge case: Empty target string is explicitly provided",
			args:         []string{"search", "-f", testFile, "--target", ""},
			wantCountStr: "Count of []: 0",
			wantErr:      false,
		},
		{
			name:        "Error case: Invalid config path",
			args:        []string{"search", "-f", testFile, "-c", filepath.Join(tmpDir, "missing.yaml")},
			wantErr:     true,
			errContains: "failed to read config file",
		},
		{
			name:        "Error case: Malformed config file",
			args:        []string{"search", "-f", testFile, "-c", filepath.Join(tmpDir, "bad.yaml")},
			wantErr:     true,
			errContains: "failed to parse config file",
		},
		{
			name:         "Config case: Multiple rules are processed sequentially",
			args:         []string{"search", "-f", testFile, "-c", filepath.Join(tmpDir, "rules.yaml")},
			wantCountStr: "Count of [Hello]: 3",
			wantErr:      false,
		},
	}

	rulesFile := filepath.Join(tmpDir, "rules.yaml")
	if err := os.WriteFile(rulesFile, []byte(`rules:
  - target: "Hello"
    replacement: "World"
    ignore_case: true
  - target: "Go"
    replacement: "Gopher"
`), 0o644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
	badConfigFile := filepath.Join(tmpDir, "bad.yaml")
	if err := os.WriteFile(badConfigFile, []byte(`rules: [
`), 0o644); err != nil {
		t.Fatalf("Failed to create malformed config file: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Buffers to capture stdout and stderr
			out, err := executeSearchCmd(tt.args[1:]...)

			// Validate error result
			if (err != nil) != tt.wantErr {
				t.Fatalf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error string: %q, actual error: %q", tt.errContains, err.Error())
				}
			}

			// Validate output for normal cases
			if !tt.wantErr && tt.wantCountStr != "" {
				if !strings.Contains(out, tt.wantCountStr) {
					t.Errorf("Expected output string not found.\nExpected: %q\nActual output:\n%s", tt.wantCountStr, out)
				}
			}

			if tt.name == "Config case: Multiple rules are processed sequentially" {
				if !strings.Contains(out, "Count of [Go]: 1") {
					t.Errorf("Expected output string for second rule not found.\nActual output:\n%s", out)
				}
			}
		})
	}
}
