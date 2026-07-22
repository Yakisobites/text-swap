package textproc

import (
	"bytes"
	"strings"
	"testing"
)

func TestReplaceWords(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		oldWord    string
		newWord    string
		opts       ReplaceOptions
		wantOutput string
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "Normal case: Replace with case sensitivity",
			input:      "Hello World\nhello Go\nHELLO Cobra",
			oldWord:    "hello",
			newWord:    "hi",
			opts:       ReplaceOptions{IgnoreCase: false},
			wantOutput: "Hello World\nhi Go\nHELLO Cobra",
			wantCount:  1,
			wantErr:    false,
		},
		{
			name:       "Normal case: Replace all occurrences ignoring case (-i)",
			input:      "Hello World\nhello Go\nHELLO Cobra",
			oldWord:    "hello",
			newWord:    "hi",
			opts:       ReplaceOptions{IgnoreCase: true},
			wantOutput: "hi World\nhi Go\nhi Cobra",
			wantCount:  3,
			wantErr:    false,
		},
		{
			name:       "Edge case: Ignore-case replacement treats $ literally",
			input:      "Hello hello",
			oldWord:    "hello",
			newWord:    "$1",
			opts:       ReplaceOptions{IgnoreCase: true},
			wantOutput: "$1 $1",
			wantCount:  2,
			wantErr:    false,
		},
		{
			name:       "Normal case: Replace multiple occurrences in a single line",
			input:      "apple banana apple",
			oldWord:    "apple",
			newWord:    "orange",
			opts:       ReplaceOptions{IgnoreCase: false},
			wantOutput: "orange banana orange",
			wantCount:  2,
			wantErr:    false,
		},
		{
			name:       "Boundary case: Target word is an empty string",
			input:      "Hello World",
			oldWord:    "",
			newWord:    "X",
			opts:       ReplaceOptions{IgnoreCase: false},
			wantOutput: "Hello World",
			wantCount:  0,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			outBuf := new(bytes.Buffer)

			gotCount, err := ReplaceWords(r, outBuf, tt.oldWord, tt.newWord, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ReplaceWords() error = %v, wantErr %v", err, tt.wantErr)
			}

			if gotCount != tt.wantCount {
				t.Errorf("ReplaceWords() count = %v, want %v", gotCount, tt.wantCount)
			}

			if gotOutput := outBuf.String(); gotOutput != tt.wantOutput {
				t.Errorf("ReplaceWords() output = %q, want %q", gotOutput, tt.wantOutput)
			}
		})
	}
}
