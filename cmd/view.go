package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

type viewMode int

const (
	modeSearch viewMode = iota
	modeDiff
)

// Styles using Lipgloss
var (
	searchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1E1E2E")).
			Background(lipgloss.Color("#F9E2AF")).
			Bold(true)

	replaceStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1E1E2E")).
			Background(lipgloss.Color("#A6E3A1")).
			Bold(true)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#CDD6F4")).
			Background(lipgloss.Color("#313244")).
			Padding(0, 1)

	paneBorder = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#585B70"))
)

type model struct {
	mode        viewMode
	filePath    string
	searchTerm  string
	replaceTerm string
	rawContent  string

	vpLeft  viewport.Model
	vpRight viewport.Model

	width  int
	height int
	ready  bool
}

func newModel(mode viewMode, filePath, search, replace, content string) model {
	return model{
		mode:        mode,
		filePath:    filePath,
		searchTerm:  search,
		replaceTerm: replace,
		rawContent:  content,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}

		if m.mode == modeDiff {
			// Synchronize scrolling between left and right viewports
			var cmdLeft, cmdRight tea.Cmd
			m.vpLeft, cmdLeft = m.vpLeft.Update(msg)
			m.vpRight, cmdRight = m.vpRight.Update(msg)

			// Lock-step YOffset sync to maintain alignment
			m.vpRight.YOffset = m.vpLeft.YOffset
			return m, tea.Batch(cmdLeft, cmdRight)
		}

		var cmd tea.Cmd
		m.vpLeft, cmd = m.vpLeft.Update(msg)
		return m, cmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		headerHeight := 2
		availableHeight := m.height - headerHeight

		if availableHeight < 1 {
			availableHeight = 1
		}

		if !m.ready {
			m.vpLeft = viewport.New(m.width, availableHeight)
			m.vpRight = viewport.New(m.width, availableHeight)
			m.ready = true
		}

		if m.mode == modeSearch {
			m.vpLeft.Width = m.width
			m.vpLeft.Height = availableHeight
			m.vpLeft.SetContent(m.applyHighlight(m.rawContent, m.searchTerm, searchStyle))
		} else {
			// Calculate side-by-side pane dimensions with borders
			borderHorizontalOverhead := 2
			paneWidth := (m.width / 2) - borderHorizontalOverhead
			if paneWidth < 10 {
				paneWidth = 10
			}

			paneHeight := availableHeight - 2 // Account for top/bottom borders
			if paneHeight < 1 {
				paneHeight = 1
			}

			m.vpLeft.Width = paneWidth
			m.vpLeft.Height = paneHeight
			m.vpRight.Width = paneWidth
			m.vpRight.Height = paneHeight

			leftText := m.applyHighlight(m.rawContent, m.searchTerm, searchStyle)
			replacedText := strings.ReplaceAll(m.rawContent, m.searchTerm, m.replaceTerm)
			rightText := m.applyHighlight(replacedText, m.replaceTerm, replaceStyle)

			m.vpLeft.SetContent(leftText)
			m.vpRight.SetContent(rightText)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) applyHighlight(content, target string, style lipgloss.Style) string {
	if target == "" {
		return content
	}
	return strings.ReplaceAll(content, target, style.Render(target))
}

func (m model) View() string {
	if !m.ready {
		return "Initializing viewer..."
	}

	if m.mode == modeSearch {
		header := headerStyle.Render(fmt.Sprintf("File: %s | Search: '%s'", m.filePath, m.searchTerm))
		return lipgloss.JoinVertical(lipgloss.Left, header, m.vpLeft.View())
	}

	// Diff view mode layout
	leftTitle := headerStyle.Render(fmt.Sprintf("Original: '%s'", m.searchTerm))
	rightTitle := headerStyle.Render(fmt.Sprintf("Replaced: '%s'", m.replaceTerm))

	leftPane := paneBorder.Width(m.vpLeft.Width).Render(m.vpLeft.View())
	rightPane := paneBorder.Width(m.vpRight.Width).Render(m.vpRight.View())

	leftCol := lipgloss.JoinVertical(lipgloss.Left, leftTitle, leftPane)
	rightCol := lipgloss.JoinVertical(lipgloss.Left, rightTitle, rightPane)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightCol)
}

// Flag variables
var (
	searchFlag  string
	replaceFlag string
)

// viewCmd represents the view command
var viewCmd = &cobra.Command{
	Use:   "view <file>",
	Short: "Interactive file viewer with search highlighting and side-by-side diff",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		contentBytes, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		mode := modeSearch
		if replaceFlag != "" {
			mode = modeDiff
		}

		p := tea.NewProgram(
			newModel(mode, filePath, searchFlag, replaceFlag, string(contentBytes)),
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
		)

		if _, err := p.Run(); err != nil {
			return fmt.Errorf("execution error: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(viewCmd)

	// Command flags definition
	viewCmd.Flags().StringVarP(&searchFlag, "search", "s", "", "Search term for highlight mode")
	viewCmd.Flags().StringVarP(&replaceFlag, "replace", "r", "", "Replacement term for diff mode")
}
