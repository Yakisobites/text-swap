package config

import (
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
	if json.Valid(data) {
		var rules []Rule
		if err := json.Unmarshal(data, &rules); err == nil {
			return rules, nil
		}

		var config Config
		if err := json.Unmarshal(data, &config); err == nil {
			return config.Rules, nil
		}
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err == nil && len(config.Rules) > 0 {
		return config.Rules, nil
	}

	var rules []Rule
	if err := yaml.Unmarshal(data, &rules); err == nil {
		return rules, nil
	}

	return nil, fmt.Errorf("failed to parse data as JSON or YAML")
}
