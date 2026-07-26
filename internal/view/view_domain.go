package view

import (
	"regexp"
	"strings"
)

type Rule struct {
	Target      string
	Replacement string
	IgnoreCase  bool
}

func computeNextXOffset(current int, direction int, maxOffset int) int {
	next := current + direction
	if next < 0 {
		return 0
	}
	if next > maxOffset {
		return maxOffset
	}
	return next
}

func computeMaxXOffset(mode Mode, rawContent string, rules []Rule, viewportWidth int) int {
	if viewportWidth <= 0 {
		return 0
	}

	maxLen := maxLineLen(rawContent)
	if mode == ModeDiff {
		replacedText := buildReplacedContent(rawContent, rules)
		rightMax := maxLineLen(replacedText)
		if rightMax > maxLen {
			maxLen = rightMax
		}
	}

	if maxLen <= viewportWidth {
		return 0
	}
	return maxLen - viewportWidth
}

func buildReplacedContent(rawContent string, rules []Rule) string {
	replaced := rawContent
	for _, rule := range rules {
		if rule.Target == "" {
			continue
		}

		if !rule.IgnoreCase {
			replaced = strings.ReplaceAll(replaced, rule.Target, rule.Replacement)
			continue
		}

		re, err := regexp.Compile("(?i)" + regexp.QuoteMeta(rule.Target))
		if err != nil {
			continue
		}
		replaced = re.ReplaceAllLiteralString(replaced, rule.Replacement)
	}

	return replaced
}

func sliceAndHighlightContent(content string, rules []Rule, xOffset, limitWidth int, highlight func(ruleIndex int, target string) string) string {
	lines := strings.Split(content, "\n")
	processedLines := make([]string, len(lines))

	for i, line := range lines {
		runes := []rune(line)
		if xOffset >= len(runes) {
			processedLines[i] = ""
			continue
		}

		end := xOffset + limitWidth
		if limitWidth > 0 && end < len(runes) {
			runes = runes[xOffset:end]
		} else {
			runes = runes[xOffset:]
		}

		visibleLine := string(runes)
		if highlight != nil {
			visibleLine = highlightLineByRules(visibleLine, rules, highlight)
		}
		processedLines[i] = visibleLine
	}

	return strings.Join(processedLines, "\n")
}

func highlightLineByRules(line string, rules []Rule, highlight func(ruleIndex int, target string) string) string {
	highlighted := line
	for i, rule := range rules {
		if rule.Target == "" {
			continue
		}

		if !rule.IgnoreCase {
			highlighted = strings.ReplaceAll(highlighted, rule.Target, highlight(i, rule.Target))
			continue
		}

		re, err := regexp.Compile("(?i)" + regexp.QuoteMeta(rule.Target))
		if err != nil {
			continue
		}
		highlighted = re.ReplaceAllStringFunc(highlighted, func(matched string) string {
			return highlight(i, matched)
		})
	}

	return highlighted
}

func maxLineLen(content string) int {
	maxLen := 0
	for _, line := range strings.Split(content, "\n") {
		length := len([]rune(line))
		if length > maxLen {
			maxLen = length
		}
	}
	return maxLen
}
