package textproc

import (
	"strings"
	"testing"
)

func TestCountOccurrences(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		target    string
		opts      SearchOptions
		wantCount int
		wantErr   bool
	}{
		{
			name:      "Normal case: Case-sensitive (1 match)",
			input:     "Hello World\nhello Go\nHELLO Cobra",
			target:    "hello",
			opts:      SearchOptions{IgnoreCase: false},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "Normal case: Case-insensitive (3 matches)",
			input:     "Hello World\nhello Go\nHELLO Cobra",
			target:    "hello",
			opts:      SearchOptions{IgnoreCase: true},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "Normal case: Multiple occurrences in a single line",
			input:     "hello world hello go hello",
			target:    "hello",
			opts:      SearchOptions{IgnoreCase: false},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "Normal case: Target word does not exist",
			input:     "Hello World",
			target:    "Python",
			opts:      SearchOptions{IgnoreCase: false},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "Boundary case: Empty search word",
			input:     "Hello World",
			target:    "",
			opts:      SearchOptions{IgnoreCase: false},
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a reader from the input text
			r := strings.NewReader(tt.input)

			// Execute CountOccurrences
			got, err := CountOccurrences(r, tt.target, tt.opts)

			// Validate error result
			if (err != nil) != tt.wantErr {
				t.Fatalf("CountOccurrences() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Validate count result
			if got != tt.wantCount {
				t.Errorf("CountOccurrences() = %v, want %v", got, tt.wantCount)
			}
		})
	}
}

func TestCountOccurrencesChunked(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		target    string
		opts      SearchOptions
		chunkSize int
		wantCount int
		wantErr   bool
	}{
		{
			name:      "Line boundary chunk split",
			input:     "hello one\nhello two\nhello three\n",
			target:    "hello",
			opts:      SearchOptions{IgnoreCase: false},
			chunkSize: 12,
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "Case-insensitive chunked search",
			input:     "Hello\nhello\nHELLO\n",
			target:    "hello",
			opts:      SearchOptions{IgnoreCase: true},
			chunkSize: 8,
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "Last line without trailing newline",
			input:     "alpha\nalpha",
			target:    "alpha",
			opts:      SearchOptions{IgnoreCase: false},
			chunkSize: 6,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "Invalid chunk size",
			input:     "hello",
			target:    "hello",
			opts:      SearchOptions{IgnoreCase: false},
			chunkSize: 0,
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			got, err := CountOccurrencesChunked(r, tt.target, tt.opts, tt.chunkSize)

			if (err != nil) != tt.wantErr {
				t.Fatalf("CountOccurrencesChunked() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got != tt.wantCount {
				t.Errorf("CountOccurrencesChunked() = %v, want %v", got, tt.wantCount)
			}
		})
	}
}

func TestCountOccurrencesChunked_ParityWithSequential(t *testing.T) {
	input := "Hello hello\nHELLO\nhello world hello\n"
	target := "hello"
	opts := SearchOptions{IgnoreCase: true}

	sequential, err := CountOccurrences(strings.NewReader(input), target, opts)
	if err != nil {
		t.Fatalf("CountOccurrences() error = %v", err)
	}

	chunked, err := CountOccurrencesChunked(strings.NewReader(input), target, opts, 9)
	if err != nil {
		t.Fatalf("CountOccurrencesChunked() error = %v", err)
	}

	if sequential != chunked {
		t.Errorf("sequential count %d != chunked count %d", sequential, chunked)
	}
}
