package textproc

import (
	"io"
	"regexp"
	"strings"

	"text-swap/internal/config"
)

// ReplaceAll applies multiple rules sequentially to the input text in memory and writes to the output writer.
func ReplaceAll(r io.Reader, w io.Writer, rules []config.Rule) (int, error) {
	// Read entire content into memory for multi-rule chain replacement
	content, err := io.ReadAll(r)
	if err != nil {
		return 0, err
	}

	text := string(content)
	totalReplacements := 0

	for _, rule := range rules {
		if rule.Target == "" {
			continue
		}

		var count int
		if rule.IgnoreCase {
			// Compile case-insensitive regular expression for each rule
			re, err := regexp.Compile("(?i)" + regexp.QuoteMeta(rule.Target))
			if err != nil {
				return 0, err
			}

			text = re.ReplaceAllStringFunc(text, func(_ string) string {
				count++
				return rule.Replacement
			})
		} else {
			// Standard exact match replacement
			count = strings.Count(text, rule.Target)
			text = strings.ReplaceAll(text, rule.Target, rule.Replacement)
		}

		totalReplacements += count
	}

	if _, err := io.WriteString(w, text); err != nil {
		return 0, err
	}

	return totalReplacements, nil
}
