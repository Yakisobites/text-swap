package progress

import (
	"fmt"
	"io"
	"os"

	"github.com/schollz/progressbar/v3"
)

type FinishFunc func() error

func NewFileProgressReader(w io.Writer, file *os.File, description string) (io.Reader, FinishFunc, error) {
	info, err := file.Stat()
	if err != nil {
		return nil, nil, fmt.Errorf("cannot stat input file: %w", err)
	}

	total := info.Size()
	if total < 0 {
		total = 0
	}

	bar := progressbar.NewOptions64(
		total,
		progressbar.OptionSetWriter(w),
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWidth(24),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionThrottle(0),
		progressbar.OptionOnCompletion(func() {
			_, _ = fmt.Fprintln(w)
		}),
	)

	reader := io.TeeReader(file, bar)
	finish := func() error {
		return bar.Finish()
	}

	return reader, finish, nil
}
