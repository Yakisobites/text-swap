package textproc

import (
	"bufio"
	"fmt"
	"io"
	"runtime"
	"strings"
	"sync"
)

type SearchOptions struct {
	IgnoreCase bool
}

const readerBufferSize = 64 * 1024

type chunkJob struct {
	index  int
	text   string
	target string
	opts   SearchOptions
}

type chunkResult struct {
	index int
	count int
	err   error
}

func CountOccurrences(r io.Reader, target string, opts SearchOptions) (int, error) {
	if target == "" {
		return 0, nil
	}

	scanner := bufio.NewScanner(r)
	count := 0

	for scanner.Scan() {
		count += countInLine(scanner.Text(), target, opts)
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	return count, nil
}

// CountOccurrencesChunked counts target occurrences using line-boundary chunks and
// a worker pool. chunkSize must be greater than 0.
func CountOccurrencesChunked(r io.Reader, target string, opts SearchOptions, chunkSize int) (int, error) {
	if target == "" {
		return 0, nil
	}
	if chunkSize <= 0 {
		return 0, fmt.Errorf("chunk size must be greater than 0")
	}

	workerCount := runtime.NumCPU()
	if workerCount < 1 {
		workerCount = 1
	}

	jobs := make(chan chunkJob, workerCount)
	results := make(chan chunkResult, workerCount)
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				count, err := countInChunk(job.text, job.target, job.opts)
				results <- chunkResult{index: job.index, count: count, err: err}
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
		jobs <- chunkJob{
			index:  chunkIndex,
			text:   chunkBuilder.String(),
			target: target,
			opts:   opts,
		}
		chunkIndex++
		chunkByteCount = 0
		chunkBuilder.Reset()
	}

	for {
		line, err := reader.ReadString('\n')

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

		if err != nil {
			if err == io.EOF {
				break
			}
			close(jobs)
			wg.Wait()
			close(results)
			return 0, err
		}
	}

	flushChunk()
	totalJobs := chunkIndex
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	return collectCountsInOrder(results, totalJobs)
}

func collectCountsInOrder(results <-chan chunkResult, totalJobs int) (int, error) {
	if totalJobs == 0 {
		return 0, nil
	}

	nextIndex := 0
	total := 0
	pending := make(map[int]chunkResult, totalJobs)

	for res := range results {
		if res.err != nil {
			return 0, res.err
		}

		pending[res.index] = res
		for {
			next, ok := pending[nextIndex]
			if !ok {
				break
			}
			total += next.count
			delete(pending, nextIndex)
			nextIndex++
		}
	}

	if nextIndex != totalJobs {
		return 0, fmt.Errorf("missing chunk results: expected %d, got %d", totalJobs, nextIndex)
	}

	return total, nil
}

func countInChunk(chunk, target string, opts SearchOptions) (int, error) {
	scanner := bufio.NewScanner(strings.NewReader(chunk))
	count := 0

	for scanner.Scan() {
		count += countInLine(scanner.Text(), target, opts)
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	return count, nil
}

func countInLine(line, target string, opts SearchOptions) int {
	searchTarget := target
	if opts.IgnoreCase {
		line = strings.ToLower(line)
		searchTarget = strings.ToLower(target)
	}

	return strings.Count(line, searchTarget)
}
