package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	internalview "text-swap/internal/view"
)

type viewOptions struct {
	searchTerm  string
	replaceTerm string
	runProgram  func(m tea.Model, opts ...tea.ProgramOption) error
}

func newViewCmd() *cobra.Command {
	opts := &viewOptions{}
	opts.runProgram = defaultViewProgramRunner

	cmd := &cobra.Command{
		Use:   "view <file>",
		Short: "Interactive file viewer with search highlighting and side-by-side diff",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.run(cmd, args)
		},
	}

	cmd.Flags().StringVarP(&opts.searchTerm, "search", "s", "", "Search term for highlight mode")
	cmd.Flags().StringVarP(&opts.replaceTerm, "replace", "r", "", "Replacement term for diff mode")

	return cmd
}

func (o *viewOptions) run(cmd *cobra.Command, args []string) error {
	filePath := args[0]
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	mode := internalview.ModeSearch
	if o.replaceTerm != "" {
		mode = internalview.ModeDiff
	}

	m := internalview.NewModel(mode, filePath, o.searchTerm, o.replaceTerm, string(contentBytes))
	if err := o.runProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion()); err != nil {
		return fmt.Errorf("execution error: %w", err)
	}

	return nil
}

func defaultViewProgramRunner(m tea.Model, opts ...tea.ProgramOption) error {
	p := tea.NewProgram(m, opts...)
	_, err := p.Run()
	return err
}

func init() {
	rootCmd.AddCommand(newViewCmd())
}
