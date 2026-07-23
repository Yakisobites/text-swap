package config

import (
	"reflect"
	"testing"
)

func TestLoadRules(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []Rule
		wantErr bool
	}{
		{
			name: "Valid JSON - Array format",
			input: `[
				{"target": "foo", "replacement": "bar", "ignore_case": true}
			]`,
			want: []Rule{
				{Target: "foo", Replacement: "bar", IgnoreCase: true},
			},
			wantErr: false,
		},
		{
			name: "Valid JSON - Object format",
			input: `{
				"rules": [
					{"target": "hello", "replacement": "world", "ignore_case": false}
				]
			}`,
			want: []Rule{
				{Target: "hello", Replacement: "world", IgnoreCase: false},
			},
			wantErr: false,
		},
		{
			name: "Valid YAML - Object format",
			input: `
rules:
  - target: "foo"
    replacement: "bar"
    ignore_case: true
`,
			want: []Rule{
				{Target: "foo", Replacement: "bar", IgnoreCase: true},
			},
			wantErr: false,
		},
		{
			name: "Valid YAML - Array format",
			input: `
- target: "hello"
  replacement: "world"
`,
			want: []Rule{
				{Target: "hello", Replacement: "world", IgnoreCase: false},
			},
			wantErr: false,
		},
		{
			name:    "Invalid syntax",
			input:   `{invalid_json_or_yaml:`,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Empty input",
			input:   "",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "JSON with empty rules array",
			input:   `{"rules": []}`,
			want:    nil,
			wantErr: true,
		},
		{
			name: "YAML with empty rules",
			input: `
		rules:
		`,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "JSON with unrelated keys",
			input:   `{"foo": "bar"}`,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadRules([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Fatalf("LoadRules() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadRules() got = %v, want %v", got, tt.want)
			}
		})
	}
}
