package cmd

import (
	"fmt"
	"io"
	"os"

	"text-swap/internal/chunk"
	"text-swap/internal/config"
	"text-swap/internal/progress"
	"text-swap/internal/textproc"

	"github.com/spf13/cobra"
)

type searchOptions struct {
	filePath     string
	searchTarget string
	configPath   string
	ignoreCase   bool
	chunkSize    int
}

func newSearchCmd() *cobra.Command {
	opts := &searchOptions{}

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search for a word in the specified file and display the number of occurrences.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.run(cmd)
		},
	}

	cmd.Flags().StringVarP(&opts.filePath, "file", "f", "", "A file path to read")
	_ = cmd.MarkFlagRequired("file")

	cmd.Flags().StringVarP(&opts.searchTarget, "target", "t", "", "A word for search")
	cmd.Flags().StringVarP(&opts.configPath, "config", "c", "", "Path to a YAML/JSON config file containing search rules")

	cmd.MarkFlagsOneRequired("target", "config")
	cmd.MarkFlagsMutuallyExclusive("target", "config")

	cmd.Flags().BoolVarP(&opts.ignoreCase, "ignore-case", "i", false, "Case-insensitive search")
	cmd.MarkFlagsMutuallyExclusive("config", "ignore-case")

	cmd.Flags().IntVar(&opts.chunkSize, "chunk-size", 0, "Chunk size in bytes for line-boundary parallel search (0 = auto by file size)")

	return cmd
}

func (o *searchOptions) run(cmd *cobra.Command) error {
	file, err := os.Open(o.filePath)
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	if cmd.Flags().Changed("config") {
		if o.configPath == "" {
			return fmt.Errorf("--config was provided but is empty")
		}
		return o.runWithConfig(cmd)
	}

	return o.runSingleTarget(cmd, file)
}

// runWithConfig handles the search process when a config file is provided.
func (o *searchOptions) runWithConfig(cmd *cobra.Command) error {
	data, err := os.ReadFile(o.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	rules, err := config.LoadRules(data)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	for _, rule := range rules {
		inFile, err := os.Open(o.filePath)
		if err != nil {
			return fmt.Errorf("cannot open file: %w", err)
		}

		reader, finishProgress, err := progress.NewFileProgressReader(cmd.ErrOrStderr(), inFile, fmt.Sprintf("search [%s]", rule.Target))
		if err != nil {
			_ = inFile.Close()
			return err
		}

		opts := textproc.SearchOptions{IgnoreCase: rule.IgnoreCase}
		count, err := o.countOccurrences(inFile, reader, rule.Target, opts)
		if finishErr := finishProgress(); finishErr != nil && err == nil {
			err = finishErr
		}
		_ = inFile.Close()
		if err != nil {
			return fmt.Errorf("error occurred while searching for [%s]: %w", rule.Target, err)
		}

		// Removed unnecessary trailing spaces before \n
		cmd.Printf("Target Word: %s\n", rule.Target)
		cmd.Printf("Count of [%s]: %d\n", rule.Target, count)
	}

	return nil
}

// runSingleTarget handles the search process for a single target string.
func (o *searchOptions) runSingleTarget(cmd *cobra.Command, file *os.File) error {
	opts := textproc.SearchOptions{
		IgnoreCase: o.ignoreCase,
	}

	reader, finishProgress, err := progress.NewFileProgressReader(cmd.ErrOrStderr(), file, fmt.Sprintf("search [%s]", o.searchTarget))
	if err != nil {
		return err
	}

	count, err := o.countOccurrences(file, reader, o.searchTarget, opts)
	if finishErr := finishProgress(); finishErr != nil && err == nil {
		err = finishErr
	}
	if err != nil {
		return fmt.Errorf("error occurred while searching: %w", err)
	}

	cmd.Printf("Target Word: %s\n", o.searchTarget)
	cmd.Printf("Count of [%s]: %d\n", o.searchTarget, count)

	return nil
}

func (o *searchOptions) countOccurrences(file *os.File, r io.Reader, target string, opts textproc.SearchOptions) (int, error) {
	chunkSize, err := chunk.PrepareChunkSize(file, o.chunkSize)
	if err != nil {
		return 0, fmt.Errorf("cannot seek input file: %w", err)
	}

	if chunkSize > 0 {
		return textproc.CountOccurrencesChunked(r, target, opts, chunkSize)
	}

	return textproc.CountOccurrences(r, target, opts)
}

func init() {
	rootCmd.AddCommand(newSearchCmd())
}
