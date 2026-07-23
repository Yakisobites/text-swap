package config

import (
	"bytes"
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// Rule represents a replacement rule configuration.
type Rule struct {
	Target      string `json:"target" yaml:"target"`
	Replacement string `json:"replacement" yaml:"replacement"`
	IgnoreCase  bool   `json:"ignore_case" yaml:"ignore_case"`
}

// Config represents the top-level configuration for YAML.
type Config struct {
	Rules []Rule `json:"rules" yaml:"rules"`
}

// LoadRules automatically detects whether the input data is JSON or YAML and loads it.
func LoadRules(data []byte) ([]Rule, error) {
	// Guard clause: Return error if input is empty or contains only whitespace
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, fmt.Errorf("input data is empty")
	}
	var rules []Rule

	// Try parsing as JSON
	if json.Valid(data) {
		// Top-level array: [{"target": "..."}]
		if err := json.Unmarshal(data, &rules); err == nil && len(rules) > 0 {
			return rules, nil
		}

		// Top-level object: {"rules": [...]}`
		var config Config
		if err := json.Unmarshal(data, &config); err == nil && len(config.Rules) > 0 {
			return config.Rules, nil
		}
	}

	// Try parsing as YAML
	// Top-level object: rules: [...]
	var config Config
	if err := yaml.Unmarshal(data, &config); err == nil && len(config.Rules) > 0 {
		return config.Rules, nil
	}

	// Top-level array: - target: "..."
	if err := yaml.Unmarshal(data, &rules); err == nil && len(rules) > 0 {
		return rules, nil
	}

	// If no rules were loaded or data format is invalid
	return nil, fmt.Errorf("no valid rules found in the input data")
}
