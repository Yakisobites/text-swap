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

	cmd.Flags().StringVarP(&opts.configPath, "config", "c", "", "A config file path contains words to be replaced")

	// Ensure that either --target or --config is provided, but not both
	cmd.MarkFlagsOneRequired("target", "config")
	cmd.MarkFlagsMutuallyExclusive("target", "config")

	cmd.Flags().StringVarP(&opts.replacement, "replacement", "r", "", "A new word to replace with")

	cmd.Flags().StringVarP(&opts.outPath, "out", "o", "", "A file path to write output (default: stdout)")

	cmd.Flags().BoolVarP(&opts.ignoreCase, "ignore-case", "i", false, "Case-insensitive replacement")

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

	var rules []config.Rule
	if o.configPath != "" {
		data, err := os.ReadFile(o.configPath)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}

		rules, err = config.LoadRules(data)
		if err != nil {
			return fmt.Errorf("failed to parse config file: %w", err)
		}
	} else if o.target != "" {
		rules = []config.Rule{
			{
				Target:      o.target,
				Replacement: o.replacement,
				IgnoreCase:  o.ignoreCase,
			},
		}
	}

	var outFile io.Writer
	var closeFn func() error

	if o.outPath != "" {
		if o.outPath == o.filePath {
			return fmt.Errorf("output path must be different from input file")
		}
		f, err := os.Create(o.outPath)
		if err != nil {
			return fmt.Errorf("cannot create output file: %w", err)
		}
		outFile = f
		closeFn = f.Close
	} else {
		outFile = cmd.OutOrStdout()
		closeFn = func() error { return nil }
	}

	defer func() {
		_ = closeFn()
	}()

	// Pass all rules at once to textproc to avoid reading from an exhausted io.Reader
	totalCount, err := textproc.ReplaceAll(inFile, outFile, rules)
	if err != nil {
		return fmt.Errorf("error occurred while replacing: %w", err)
	}

	if o.outPath != "" {
		cmd.Printf("Replaced %d occurrences in %s\n", totalCount, o.outPath)
	}

	if err := closeFn(); err != nil {
		return fmt.Errorf("cannot close output file: %w", err)
	}
	closeFn = func() error { return nil }

	return nil
}

func init() {
	rootCmd.AddCommand(newReplaceCmd())
}
