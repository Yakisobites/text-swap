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

var (
	filePath     string
	searchTarget string
	configPath   string
	ignoreCase   bool
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search for a word in the specified file and display the number of occurrences.",
	RunE: func(cmd *cobra.Command, args []string) error {
		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("cannot open a file: %w", err)
		}
		defer func() {
			_ = file.Close()
		}()

		input, err := io.ReadAll(file)
		if err != nil {
			return fmt.Errorf("cannot read input file: %w", err)
		}

		if configPath != "" {
			data, err := os.ReadFile(configPath)
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
					return fmt.Errorf("error occurred while searching: %w", err)
				}

				cmd.Printf("Target Word: %s \n", rule.Target)
				cmd.Printf("Count of[%s]: %d \n", rule.Target, count)
			}

			return nil
		}

		opts := textproc.SearchOptions{
			IgnoreCase: ignoreCase,
		}

		count, err := textproc.CountOccurrences(bytes.NewReader(input), searchTarget, opts)
		if err != nil {
			return fmt.Errorf("error occurred while searching: %w", err)
		}

		cmd.Printf("Target Word: %s \n", searchTarget)
		cmd.Printf("Count of[%s]: %d \n", searchTarget, count)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().StringVarP(&filePath, "file", "f", "", "A file path to read")
	_ = searchCmd.MarkFlagRequired("file")

	searchCmd.Flags().StringVarP(&searchTarget, "target", "t", "", "A word for search")
	searchCmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to a YAML/JSON config file containing search rules")

	searchCmd.MarkFlagsOneRequired("target", "config")
	searchCmd.MarkFlagsMutuallyExclusive("target", "config")

	searchCmd.Flags().BoolVarP(&ignoreCase, "ignore-case", "i", false, "Case-insensitive search")
}
