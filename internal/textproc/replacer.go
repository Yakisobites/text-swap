package textproc

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"text-swap/internal/config"
)

// compiledRule holds the pre-compiled regex to avoid recompiling it for every line.
type compiledRule struct {
	target      string
	replacement string
	ignoreCase  bool
	re          *regexp.Regexp
}

// ReplaceAll applies multiple rules sequentially to the input text via streaming
// and writes the result to the output writer.
func ReplaceAll(r io.Reader, w io.Writer, rules []config.Rule) (int, error) {
	// Pre-compile rules before processing the stream
	var crules []compiledRule
	for _, rule := range rules {
		if rule.Target == "" {
			continue
		}

		var re *regexp.Regexp
		var err error
		if rule.IgnoreCase {
			re, err = regexp.Compile("(?i)" + regexp.QuoteMeta(rule.Target))
			if err != nil {
				return 0, err
			}
		}

		crules = append(crules, compiledRule{
			target:      rule.Target,
			replacement: rule.Replacement,
			ignoreCase:  rule.IgnoreCase,
			re:          re,
		})
	}

	// Set up buffered I/O for streaming
	reader := bufio.NewReaderSize(r, 64*1024)
	writer := bufio.NewWriter(w)
	totalReplacements := 0

	// Process line by line
	for {
		line, err := reader.ReadString('\n')

		// Process the line even if EOF is reached, to handle files without a trailing newline.
		if len(line) > 0 {
			for _, crule := range crules {
				if crule.ignoreCase {
					// Count occurrences before replacing to avoid closures
					count := len(crule.re.FindAllString(line, -1))
					if count > 0 {
						line = crule.re.ReplaceAllLiteralString(line, crule.replacement)
						totalReplacements += count
					}
				} else {
					count := strings.Count(line, crule.target)
					if count > 0 {
						line = strings.ReplaceAll(line, crule.target, crule.replacement)
						totalReplacements += count
					}
				}
			}

			if _, wErr := writer.WriteString(line); wErr != nil {
				return totalReplacements, wErr
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return totalReplacements, err
		}
	}

	// Flush the buffered writer to ensure all data is written to the underlying io.Writer
	if err := writer.Flush(); err != nil {
		return totalReplacements, err
	}

	return totalReplacements, nil
}
