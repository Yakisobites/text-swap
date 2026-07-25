package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

type progressFinishFunc func() error

func newFileProgressReader(cmd *cobra.Command, file *os.File, description string) (io.Reader, progressFinishFunc, error) {
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
		progressbar.OptionSetWriter(cmd.ErrOrStderr()),
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWidth(24),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionThrottle(0),
		progressbar.OptionOnCompletion(func() {
			_, _ = fmt.Fprintln(cmd.ErrOrStderr())
		}),
	)

	reader := io.TeeReader(file, bar)
	finish := func() error {
		return bar.Finish()
	}

	return reader, finish, nil
}
