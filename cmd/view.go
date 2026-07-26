package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"text-swap/internal/config"
	internalview "text-swap/internal/view"
)

type viewOptions struct {
	searchTerm  string
	replaceTerm string
	configPath  string
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
	cmd.Flags().StringVarP(&opts.configPath, "config", "c", "", "Path to a YAML/JSON config file containing search/replace rules")

	cmd.MarkFlagsMutuallyExclusive("config", "search")
	cmd.MarkFlagsMutuallyExclusive("config", "replace")

	return cmd
}

func (o *viewOptions) run(_ *cobra.Command, args []string) error {
	filePath := args[0]
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	rules, err := o.loadRules()
	if err != nil {
		return err
	}

	mode := o.detectMode(rules)

	m := internalview.NewModelWithRules(mode, filePath, rules, string(contentBytes))
	if err := o.runProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion()); err != nil {
		return fmt.Errorf("execution error: %w", err)
	}

	return nil
}

func (o *viewOptions) loadRules() ([]internalview.Rule, error) {
	if o.configPath == "" {
		return []internalview.Rule{
			{
				Target:      o.searchTerm,
				Replacement: o.replaceTerm,
			},
		}, nil
	}

	data, err := os.ReadFile(o.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	loaded, err := config.LoadRules(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	rules := make([]internalview.Rule, 0, len(loaded))
	for _, rule := range loaded {
		rules = append(rules, internalview.Rule{
			Target:      rule.Target,
			Replacement: rule.Replacement,
			IgnoreCase:  rule.IgnoreCase,
		})
	}

	return rules, nil
}

func (o *viewOptions) detectMode(rules []internalview.Rule) internalview.Mode {
	if o.configPath == "" {
		if o.replaceTerm != "" {
			return internalview.ModeDiff
		}
		return internalview.ModeSearch
	}

	for _, rule := range rules {
		if rule.Replacement != "" {
			return internalview.ModeDiff
		}
	}

	return internalview.ModeSearch
}

func defaultViewProgramRunner(m tea.Model, opts ...tea.ProgramOption) error {
	p := tea.NewProgram(m, opts...)
	_, err := p.Run()
	return err
}

func init() {
	rootCmd.AddCommand(newViewCmd())
}
