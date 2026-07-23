package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"text-swap/internal/config"
	"text-swap/internal/textproc"

	"github.com/spf13/cobra"
)

type searchOptions struct {
	filePath     string
	searchTarget string
	configPath   string
	ignoreCase   bool
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

	// Load file content entirely into memory.
	// Note: Be cautious of OOM when handling extremely large files.
	input, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("cannot read input file: %w", err)
	}

	if cmd.Flags().Changed("config") {
		if o.configPath == "" {
			return fmt.Errorf("--config was provided but is empty")
		}
		return o.runWithConfig(cmd, input)
	}

	return o.runSingleTarget(cmd, input)
}

// runWithConfig handles the search process when a config file is provided.
func (o *searchOptions) runWithConfig(cmd *cobra.Command, input []byte) error {
	data, err := os.ReadFile(o.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	rules, err := config.LoadRules(data)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	for _, rule := range rules {
		opts := textproc.SearchOptions{IgnoreCase: rule.IgnoreCase}
		count, err := textproc.CountOccurrences(bytes.NewReader(input), rule.Target, opts)
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
func (o *searchOptions) runSingleTarget(cmd *cobra.Command, input []byte) error {
	opts := textproc.SearchOptions{
		IgnoreCase: o.ignoreCase,
	}

	count, err := textproc.CountOccurrences(bytes.NewReader(input), o.searchTarget, opts)
	if err != nil {
		return fmt.Errorf("error occurred while searching: %w", err)
	}

	cmd.Printf("Target Word: %s\n", o.searchTarget)
	cmd.Printf("Count of [%s]: %d\n", o.searchTarget, count)

	return nil
}

func init() {
	rootCmd.AddCommand(newSearchCmd())
}
