package textproc

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"text-swap/internal/config"
)

// compiledRule holds the pre-compiled regex to avoid recompiling it for every line.
type compiledRule struct {
	target      string
	replacement string
	ignoreCase  bool
	re          *regexp.Regexp
}

type replaceChunkJob struct {
	index int
	text  string
}

type replaceChunkResult struct {
	index int
	text  string
	count int
	err   error
}

// ReplaceAll applies multiple rules sequentially to the input text via streaming
// and writes the result to the output writer.
func ReplaceAll(r io.Reader, w io.Writer, rules []config.Rule) (int, error) {
	crules, err := compileRules(rules)
	if err != nil {
		return 0, err
	}

	// Set up buffered I/O for streaming
	reader := bufio.NewReaderSize(r, readerBufferSize)
	writer := bufio.NewWriter(w)
	totalReplacements := 0

	// Process line by line
	for {
		line, err := reader.ReadString('\n')

		// Process the line even if EOF is reached, to handle files without a trailing newline.
		if len(line) > 0 {
			line, count := applyRulesToLine(line, crules)
			totalReplacements += count

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

// ReplaceAllChunked applies multiple rules in parallel chunks split by line boundaries.
// chunkSize must be greater than 0.
func ReplaceAllChunked(r io.Reader, w io.Writer, rules []config.Rule, chunkSize int) (int, error) {
	if chunkSize <= 0 {
		return 0, fmt.Errorf("chunk size must be greater than 0")
	}

	crules, err := compileRules(rules)
	if err != nil {
		return 0, err
	}

	workerCount := runtime.NumCPU()
	if workerCount < 1 {
		workerCount = 1
	}

	jobs := make(chan replaceChunkJob, workerCount)
	results := make(chan replaceChunkResult, workerCount)
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				out, count, err := replaceChunk(job.text, crules)
				results <- replaceChunkResult{index: job.index, text: out, count: count, err: err}
			}
		}()
	}

	reader := bufio.NewReaderSize(r, readerBufferSize)
	chunkIndex := 0
	chunkByteCount := 0
	var chunkBuilder strings.Builder

	flushChunk := func() {
		if chunkBuilder.Len() == 0 {
			return
		}
		jobs <- replaceChunkJob{index: chunkIndex, text: chunkBuilder.String()}
		chunkIndex++
		chunkByteCount = 0
		chunkBuilder.Reset()
	}

	for {
		line, readErr := reader.ReadString('\n')

		if len(line) > 0 {
			if chunkByteCount > 0 && chunkByteCount+len(line) > chunkSize {
				flushChunk()
			}
			chunkBuilder.WriteString(line)
			chunkByteCount += len(line)

			if chunkByteCount >= chunkSize {
				flushChunk()
			}
		}

		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			close(jobs)
			wg.Wait()
			close(results)
			return 0, readErr
		}
	}

	flushChunk()
	totalJobs := chunkIndex
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	return writeReplacedChunksInOrder(w, results, totalJobs)
}

func writeReplacedChunksInOrder(w io.Writer, results <-chan replaceChunkResult, totalJobs int) (int, error) {
	if totalJobs == 0 {
		return 0, nil
	}

	writer := bufio.NewWriter(w)
	nextIndex := 0
	totalReplacements := 0
	pending := make(map[int]replaceChunkResult, totalJobs)

	for res := range results {
		if res.err != nil {
			return totalReplacements, res.err
		}

		pending[res.index] = res
		for {
			next, ok := pending[nextIndex]
			if !ok {
				break
			}
			if _, err := writer.WriteString(next.text); err != nil {
				return totalReplacements, err
			}
			totalReplacements += next.count
			delete(pending, nextIndex)
			nextIndex++
		}
	}

	if nextIndex != totalJobs {
		return totalReplacements, fmt.Errorf("missing chunk results: expected %d, got %d", totalJobs, nextIndex)
	}

	if err := writer.Flush(); err != nil {
		return totalReplacements, err
	}

	return totalReplacements, nil
}

func replaceChunk(chunk string, rules []compiledRule) (string, int, error) {
	reader := bufio.NewReaderSize(strings.NewReader(chunk), readerBufferSize)
	var out strings.Builder
	totalReplacements := 0

	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			replacedLine, count := applyRulesToLine(line, rules)
			totalReplacements += count
			out.WriteString(replacedLine)
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return "", 0, err
		}
	}

	return out.String(), totalReplacements, nil
}

func compileRules(rules []config.Rule) ([]compiledRule, error) {
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
				return nil, err
			}
		}

		crules = append(crules, compiledRule{
			target:      rule.Target,
			replacement: rule.Replacement,
			ignoreCase:  rule.IgnoreCase,
			re:          re,
		})
	}

	return crules, nil
}

func applyRulesToLine(line string, rules []compiledRule) (string, int) {
	totalReplacements := 0

	for _, crule := range rules {
		if crule.ignoreCase {
			count := len(crule.re.FindAllString(line, -1))
			if count > 0 {
				line = crule.re.ReplaceAllLiteralString(line, crule.replacement)
				totalReplacements += count
			}
			continue
		}

		count := strings.Count(line, crule.target)
		if count > 0 {
			line = strings.ReplaceAll(line, crule.target, crule.replacement)
			totalReplacements += count
		}
	}

	return line, totalReplacements
}
