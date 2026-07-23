package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"text-swap/internal/config"
	"text-swap/internal/textproc"
)

type replaceOptions struct {
	filePath    string
	outPath     string
	target      string
	configPath  string
	replacement string
	ignoreCase  bool
}

func newReplaceCmd() *cobra.Command {
	opts := &replaceOptions{}

	cmd := &cobra.Command{
		Use:   "replace",
		Short: "Replace target string in the specified file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.run(cmd)
		},
	}

	cmd.Flags().StringVarP(&opts.filePath, "file", "f", "", "A file path to read")
	_ = cmd.MarkFlagRequired("file")

	cmd.Flags().StringVarP(&opts.target, "target", "t", "", "A word to be replaced")
	cmd.Flags().StringVarP(&opts.configPath, "config", "c", "", "Path to a YAML/JSON config file containing replacement rules")
	cmd.Flags().StringVarP(&opts.replacement, "replacement", "r", "", "A new word to replace with")
	cmd.Flags().StringVarP(&opts.outPath, "out", "o", "", "A file path to write output (default: stdout)")
	cmd.Flags().BoolVarP(&opts.ignoreCase, "ignore-case", "i", false, "Case-insensitive replacement")

	// Ensure that either --target or --config is provided, but not both
	cmd.MarkFlagsOneRequired("target", "config")
	cmd.MarkFlagsMutuallyExclusive("target", "config")

	// Prevent users from providing single-target options when using a config file
	cmd.MarkFlagsMutuallyExclusive("config", "replacement")
	cmd.MarkFlagsMutuallyExclusive("config", "ignore-case")

	return cmd
}

func (o *replaceOptions) run(cmd *cobra.Command) error {
	inFile, err := os.Open(o.filePath)
	if err != nil {
		return fmt.Errorf("cannot open input file: %w", err)
	}
	defer func() {
		_ = inFile.Close()
	}()

	rules, err := o.loadRules(cmd)
	if err != nil {
		return err
	}

	outFile, closeFn, err := o.setupOutput(cmd)
	if err != nil {
		return err
	}
	defer func() {
		_ = closeFn()
	}()

	totalCount, err := textproc.ReplaceAll(inFile, outFile, rules)
	if err != nil {
		return fmt.Errorf("error occurred while replacing: %w", err)
	}

	if o.outPath != "" {
		cmd.Printf("Replaced %d occurrences in %s\n", totalCount, o.outPath)
	}

	// Capture any error that occurs when actually closing the file
	if err := closeFn(); err != nil {
		return fmt.Errorf("cannot close output file: %w", err)
	}
	closeFn = func() error { return nil }

	return nil
}

// loadRules determines the source of the replacement rules and loads them.
func (o *replaceOptions) loadRules(cmd *cobra.Command) ([]config.Rule, error) {
	if cmd.Flags().Changed("config") {
		if o.configPath == "" {
			return nil, fmt.Errorf("--config was provided but is empty")
		}
		data, err := os.ReadFile(o.configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		rules, err := config.LoadRules(data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
		return rules, nil
	}

	if cmd.Flags().Changed("target") {
		return []config.Rule{
			{
				Target:      o.target,
				Replacement: o.replacement,
				IgnoreCase:  o.ignoreCase,
			},
		}, nil
	}

	// This should logically not be reached due to Cobra's MarkFlagsOneRequired,
	// but it's safe to return an error just in case.
	return nil, fmt.Errorf("either --target or --config must be provided")
}

// setupOutput determines the destination io.Writer and provides a cleanup function.
func (o *replaceOptions) setupOutput(cmd *cobra.Command) (io.Writer, func() error, error) {
	if o.outPath == "" {
		// Output to stdout if no outPath is specified
		return cmd.OutOrStdout(), func() error { return nil }, nil
	}

	if o.outPath == o.filePath {
		return nil, nil, fmt.Errorf("output path must be different from input file")
	}

	f, err := os.Create(o.outPath)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create output file: %w", err)
	}

	return f, f.Close, nil
}

func init() {
	rootCmd.AddCommand(newReplaceCmd())
}
