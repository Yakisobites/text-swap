package textproc

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

// ReplaceOptions represents options for the replacement process.
type ReplaceOptions struct {
	IgnoreCase bool
}

// ReplaceWords reads text from r, replaces oldWord with newWord, and writes the result to w.
// It returns the total number of replacements made.
func ReplaceWords(r io.Reader, w io.Writer, oldWord, newWord string, opts ReplaceOptions) (int, error) {
	if oldWord == "" {
		_, err := io.Copy(w, r)
		return 0, err
	}

	reader := bufio.NewReaderSize(r, 64*1024)
	writer := bufio.NewWriter(w)

	totalReplaced := 0

	var re *regexp.Regexp
	var err error
	if opts.IgnoreCase {
		re, err = regexp.Compile("(?i)" + regexp.QuoteMeta(oldWord))
		if err != nil {
			return 0, err
		}
	}

	// isFirstLine := true
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			var newLine string
			var count int

			if opts.IgnoreCase {
				count = 0
				newLine = re.ReplaceAllStringFunc(line, func(_ string) string {
					count++
					return newWord
				})
			} else {
				count = strings.Count(line, oldWord)
				newLine = strings.ReplaceAll(line, oldWord, newWord)
			}

			totalReplaced += count

			if _, err := writer.WriteString(newLine); err != nil {
				return 0, err
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return 0, err
		}
	}

	if err := writer.Flush(); err != nil {
		return 0, err
	}

	return totalReplaced, nil
}
