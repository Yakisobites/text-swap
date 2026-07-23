package textproc

import (
	"bytes"
	"strings"
	"testing"

	"text-swap/internal/config"
)

func TestReplaceAll(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		rules      []config.Rule
		wantOutput string
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "Normal case: Replace with case sensitivity",
			input:      "Hello World\nhello Go\nHELLO Cobra",
			rules:      []config.Rule{{Target: "hello", Replacement: "hi", IgnoreCase: false}},
			wantOutput: "Hello World\nhi Go\nHELLO Cobra",
			wantCount:  1,
			wantErr:    false,
		},
		{
			name:       "Normal case: Replace all occurrences ignoring case (-i)",
			input:      "Hello World\nhello Go\nHELLO Cobra",
			rules:      []config.Rule{{Target: "hello", Replacement: "hi", IgnoreCase: true}},
			wantOutput: "hi World\nhi Go\nhi Cobra",
			wantCount:  3,
			wantErr:    false,
		},
		{
			name:       "Edge case: Ignore-case replacement treats $ literally",
			input:      "Hello hello",
			rules:      []config.Rule{{Target: "hello", Replacement: "$1", IgnoreCase: true}},
			wantOutput: "$1 $1",
			wantCount:  2,
			wantErr:    false,
		},
		{
			name:       "Normal case: Replace multiple occurrences in a single line",
			input:      "apple banana apple",
			rules:      []config.Rule{{Target: "apple", Replacement: "orange", IgnoreCase: false}},
			wantOutput: "orange banana orange",
			wantCount:  2,
			wantErr:    false,
		},
		{
			name:       "Boundary case: Target word is an empty string",
			input:      "Hello World",
			rules:      []config.Rule{{Target: "", Replacement: "X", IgnoreCase: false}},
			wantOutput: "Hello World",
			wantCount:  0,
			wantErr:    false,
		},
		{
			name:  "Normal case: Apply multiple rules sequentially",
			input: "cat dog",
			rules: []config.Rule{
				{Target: "cat", Replacement: "dog", IgnoreCase: false},
				{Target: "dog", Replacement: "fox", IgnoreCase: false},
			},
			wantOutput: "fox fox",
			wantCount:  3,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			outBuf := new(bytes.Buffer)

			gotCount, err := ReplaceAll(r, outBuf, tt.rules)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ReplaceAll() error = %v, wantErr %v", err, tt.wantErr)
			}

			if gotCount != tt.wantCount {
				t.Errorf("ReplaceAll() count = %v, want %v", gotCount, tt.wantCount)
			}

			if gotOutput := outBuf.String(); gotOutput != tt.wantOutput {
				t.Errorf("ReplaceAll() output = %q, want %q", gotOutput, tt.wantOutput)
			}
		})
	}
}

func TestReplaceAll_PreservesLineEndingsAndFinalNewline(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		rules      []config.Rule
		wantOutput string
		wantCount  int
	}{
		{
			name:       "Preserve CRLF and final newline",
			input:      "foo\r\nbar\r\n",
			rules:      []config.Rule{{Target: "bar", Replacement: "baz", IgnoreCase: false}},
			wantOutput: "foo\r\nbaz\r\n",
			wantCount:  1,
		},
		{
			name:       "Preserve no final newline with LF",
			input:      "foo\nbar",
			rules:      []config.Rule{{Target: "bar", Replacement: "baz", IgnoreCase: false}},
			wantOutput: "foo\nbaz",
			wantCount:  1,
		},
		{
			name:       "Preserve no final newline with CRLF line before last line",
			input:      "foo\r\nbar",
			rules:      []config.Rule{{Target: "foo", Replacement: "qux", IgnoreCase: false}},
			wantOutput: "qux\r\nbar",
			wantCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			outBuf := new(bytes.Buffer)

			gotCount, err := ReplaceAll(r, outBuf, tt.rules)
			if err != nil {
				t.Fatalf("ReplaceAll() error = %v", err)
			}

			if gotCount != tt.wantCount {
				t.Errorf("ReplaceAll() count = %v, want %v", gotCount, tt.wantCount)
			}

			if gotOutput := outBuf.String(); gotOutput != tt.wantOutput {
				t.Errorf("ReplaceAll() output = %q, want %q", gotOutput, tt.wantOutput)
			}
		})
	}
}
