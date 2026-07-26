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

	width   int
	height  int
	xOffset int // Horizontal scroll offset
	ready   bool
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

		// Horizontal scroll handling
		case "left", "h":
			if m.xOffset > 0 {
				m.xOffset--
				m = m.updateContents()
			}
			return m, nil

		case "right", "l":
			// Prevent scrolling beyond the longest line
			if m.xOffset < m.getMaxXOffset() {
				m.xOffset++
				m = m.updateContents()
			}
			return m, nil
		}

		if m.mode == modeDiff {
			var cmdLeft, cmdRight tea.Cmd
			m.vpLeft, cmdLeft = m.vpLeft.Update(msg)
			m.vpRight, cmdRight = m.vpRight.Update(msg)

			m.vpRight.YOffset = m.vpLeft.YOffset
			return m, tea.Batch(cmdLeft, cmdRight)
		}

		var cmd tea.Cmd
		m.vpLeft, cmd = m.vpLeft.Update(msg)
		return m, cmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		headerHeight := 1
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
		} else {
			colWidth := m.width / 2
			borderOverhead := 2

			vpWidth := colWidth - borderOverhead
			if vpWidth < 10 {
				vpWidth = 10
			}

			paneHeight := availableHeight - 3
			if paneHeight < 1 {
				paneHeight = 1
			}

			m.vpLeft.Width = vpWidth
			m.vpLeft.Height = paneHeight
			m.vpRight.Width = vpWidth
			m.vpRight.Height = paneHeight
		}

		// Re-render viewport content on window resize
		m = m.updateContents()
	}

	return m, tea.Batch(cmds...)
}

// Helper to refresh viewport contents applying xOffset and highlighting
func (m model) updateContents() model {
	if m.mode == modeSearch {
		m.vpLeft.SetContent(m.processContent(m.rawContent, m.searchTerm, searchStyle, m.vpLeft.Width))
	} else {
		leftText := m.processContent(m.rawContent, m.searchTerm, searchStyle, m.vpLeft.Width)
		replacedText := strings.ReplaceAll(m.rawContent, m.searchTerm, m.replaceTerm)
		rightText := m.processContent(replacedText, m.replaceTerm, replaceStyle, m.vpRight.Width)

		m.vpLeft.SetContent(leftText)
		m.vpRight.SetContent(rightText)
	}
	return m
}

// Slice content horizontally with UTF-8 safety and apply highlights
func (m model) processContent(content, target string, style lipgloss.Style, limitWidth int) string {
	lines := strings.Split(content, "\n")
	processedLines := make([]string, len(lines))

	for i, line := range lines {
		runes := []rune(line)
		if m.xOffset >= len(runes) {
			processedLines[i] = ""
			continue
		}

		// Calculate visible slice range based on xOffset and pane width
		end := m.xOffset + limitWidth
		if limitWidth > 0 && end < len(runes) {
			runes = runes[m.xOffset:end]
		} else {
			runes = runes[m.xOffset:]
		}

		visibleLine := string(runes)
		if target != "" {
			visibleLine = strings.ReplaceAll(visibleLine, target, style.Render(target))
		}
		processedLines[i] = visibleLine
	}

	return strings.Join(processedLines, "\n")
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
		header := headerStyle.Width(m.width).Render(fmt.Sprintf("File: %s | Search: '%s'", m.filePath, m.searchTerm))
		return lipgloss.JoinVertical(lipgloss.Left, header, m.vpLeft.View())
	}

	// Diff view mode dimensions
	colWidth := m.width / 2

	// Restrain title width strictly to the column width
	titleStyle := headerStyle.Width(colWidth).MaxHeight(1)
	leftTitle := titleStyle.Render(fmt.Sprintf("Original: '%s'", m.searchTerm))
	rightTitle := titleStyle.Render(fmt.Sprintf("Replaced: '%s'", m.replaceTerm))

	// Enforce exact outer box width including borders
	leftPane := paneBorder.Width(colWidth - 2).Render(m.vpLeft.View())
	rightPane := paneBorder.Width(colWidth - 2).Render(m.vpRight.View())

	leftCol := lipgloss.JoinVertical(lipgloss.Left, leftTitle, leftPane)
	rightCol := lipgloss.JoinVertical(lipgloss.Left, rightTitle, rightPane)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightCol)
}

// Calculate maximum horizontal scroll offset based on the longest line
func (m model) getMaxXOffset() int {
	vpWidth := m.vpLeft.Width
	leftMax := maxLineLen(m.rawContent)

	maxLen := leftMax
	if m.mode == modeDiff {
		replacedText := strings.ReplaceAll(m.rawContent, m.searchTerm, m.replaceTerm)
		rightMax := maxLineLen(replacedText)
		if rightMax > maxLen {
			maxLen = rightMax
		}
	}

	maxOffset := maxLen - vpWidth
	if maxOffset < 0 {
		return 0
	}
	return maxOffset
}

// Calculate the maximum line length in runes
func maxLineLen(content string) int {
	maxLen := 0
	for _, line := range strings.Split(content, "\n") {
		length := len([]rune(line))
		if length > maxLen {
			maxLen = length
		}
	}
	return maxLen
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
