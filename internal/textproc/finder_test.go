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
