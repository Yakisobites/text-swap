package cmd

import (
	"fmt"
	"os"

	"text-swap/internal/textproc"

	"github.com/spf13/cobra"
)

var (
	filePath     string
	searchTarget string
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

		opts := textproc.SearchOptions{
			IgnoreCase: ignoreCase,
		}

		count, err := textproc.CountOccurrences(file, searchTarget, opts)
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
	_ = searchCmd.MarkFlagRequired("target")

	searchCmd.Flags().BoolVarP(&ignoreCase, "ignore-case", "i", false, "Case-insensitive search")
}
