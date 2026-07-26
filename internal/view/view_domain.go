package view

import "strings"

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

func computeMaxXOffset(mode Mode, rawContent, searchTerm, replaceTerm string, viewportWidth int) int {
	if viewportWidth <= 0 {
		return 0
	}

	maxLen := maxLineLen(rawContent)
	if mode == ModeDiff {
		replacedText := buildReplacedContent(rawContent, searchTerm, replaceTerm)
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

func buildReplacedContent(rawContent, searchTerm, replaceTerm string) string {
	return strings.ReplaceAll(rawContent, searchTerm, replaceTerm)
}

func sliceAndHighlightContent(content, target string, xOffset, limitWidth int, highlight func(string) string) string {
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
		if target != "" && highlight != nil {
			visibleLine = strings.ReplaceAll(visibleLine, target, highlight(target))
		}
		processedLines[i] = visibleLine
	}

	return strings.Join(processedLines, "\n")
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
