package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"text-swap/internal/textproc"
)

type replaceOptions struct {
	filePath    string
	outPath     string
	target      string
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
	_ = cmd.MarkFlagRequired("target")

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

	var outFile *os.File
	if o.outPath != "" {
		if o.outPath == o.filePath {
			return fmt.Errorf("output path must be different from input file")
		}
		f, err := os.Create(o.outPath)
		if err != nil {
			return fmt.Errorf("cannot create output file: %w", err)
		}
		defer func() {
			_ = f.Close()
		}()
		outFile = f
	} else {
		outFile = os.Stdout
	}

	procOpts := textproc.ReplaceOptions{
		IgnoreCase: o.ignoreCase,
	}

	count, err := textproc.ReplaceWords(inFile, outFile, o.target, o.replacement, procOpts)
	if err != nil {
		return fmt.Errorf("error occurred while replacing: %w", err)
	}

	if o.outPath != "" {
		cmd.Printf("Replaced [%s] -> [%s] (%d occurrences) in %s\n", o.target, o.replacement, count, o.outPath)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(newReplaceCmd())
}
