package textproc

import (
	"bufio"
	"io"
	"strings"
)

type SearchOptions struct {
	IgnoreCase bool
}

func CountOccurrences(r io.Reader, target string, opts SearchOptions) (int, error) {
	if target == "" {
		return 0, nil
	}

	scanner := bufio.NewScanner(r)
	count := 0

	searchTarget := target
	if opts.IgnoreCase {
		searchTarget = strings.ToLower(target)
	}

	for scanner.Scan() {
		line := scanner.Text()
		if opts.IgnoreCase {
			line = strings.ToLower(line)
		}
		count += strings.Count(line, searchTarget)
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	return count, nil
}
