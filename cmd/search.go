package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	filePath   string
	word       string
	ignoreCase bool
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search and display the specified file's content",
	RunE: func(cmd *cobra.Command, args []string) error {
		if filePath == "" {
			return fmt.Errorf("set your filepath(--file or -f)")
		}
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("searching your file is failed: %w", err)
		}
		cmd.Printf("Target Word: %s \n", word)
		cmd.Println("--- Contents of your file ---")
		cmd.Print(string(content))

		count := 0
		if ignoreCase {
			count = strings.Count(strings.ToLower(string(content)), strings.ToLower(word))
		} else {
			count = strings.Count(string(content), word)
		}
		cmd.Printf("[%s]の出現回数: %d回\n", word, count)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().StringVarP(&filePath, "file", "f", "your filepath", "setting your filePath")
	_ = searchCmd.MarkFlagRequired("file")
	searchCmd.Flags().StringVarP(&word, "target", "t", "target word", "")
	_ = searchCmd.MarkFlagRequired("target")
	searchCmd.Flags().BoolVarP(&ignoreCase, "ignore-case", "i", false, "Search while ignoring case")
}
