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

func TestReplaceAllChunked(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		rules      []config.Rule
		chunkSize  int
		wantOutput string
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "Chunked replace with small chunk size",
			input:      "hello one\nhello two\nhello three\n",
			rules:      []config.Rule{{Target: "hello", Replacement: "hi", IgnoreCase: false}},
			chunkSize:  10,
			wantOutput: "hi one\nhi two\nhi three\n",
			wantCount:  3,
			wantErr:    false,
		},
		{
			name:       "Chunked replace ignore-case",
			input:      "Hello\nhello\nHELLO\n",
			rules:      []config.Rule{{Target: "hello", Replacement: "$1", IgnoreCase: true}},
			chunkSize:  8,
			wantOutput: "$1\n$1\n$1\n",
			wantCount:  3,
			wantErr:    false,
		},
		{
			name:       "Chunked replace without trailing newline",
			input:      "foo\nbar",
			rules:      []config.Rule{{Target: "bar", Replacement: "baz", IgnoreCase: false}},
			chunkSize:  6,
			wantOutput: "foo\nbaz",
			wantCount:  1,
			wantErr:    false,
		},
		{
			name:       "Invalid chunk size",
			input:      "foo",
			rules:      []config.Rule{{Target: "foo", Replacement: "bar", IgnoreCase: false}},
			chunkSize:  0,
			wantOutput: "",
			wantCount:  0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			outBuf := new(bytes.Buffer)

			gotCount, err := ReplaceAllChunked(r, outBuf, tt.rules, tt.chunkSize)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ReplaceAllChunked() error = %v, wantErr %v", err, tt.wantErr)
			}

			if gotCount != tt.wantCount {
				t.Errorf("ReplaceAllChunked() count = %v, want %v", gotCount, tt.wantCount)
			}

			if gotOutput := outBuf.String(); gotOutput != tt.wantOutput {
				t.Errorf("ReplaceAllChunked() output = %q, want %q", gotOutput, tt.wantOutput)
			}
		})
	}
}

func TestReplaceAllChunked_ParityWithSequential(t *testing.T) {
	input := "cat dog\nCAT dog\ncat DOG\n"
	rules := []config.Rule{
		{Target: "cat", Replacement: "dog", IgnoreCase: true},
		{Target: "dog", Replacement: "fox", IgnoreCase: false},
	}

	seqOut := new(bytes.Buffer)
	seqCount, seqErr := ReplaceAll(strings.NewReader(input), seqOut, rules)
	if seqErr != nil {
		t.Fatalf("ReplaceAll() error = %v", seqErr)
	}

	chunkOut := new(bytes.Buffer)
	chunkCount, chunkErr := ReplaceAllChunked(strings.NewReader(input), chunkOut, rules, 9)
	if chunkErr != nil {
		t.Fatalf("ReplaceAllChunked() error = %v", chunkErr)
	}

	if seqCount != chunkCount {
		t.Errorf("sequential count %d != chunked count %d", seqCount, chunkCount)
	}

	if seqOut.String() != chunkOut.String() {
		t.Errorf("sequential output %q != chunked output %q", seqOut.String(), chunkOut.String())
	}
}
