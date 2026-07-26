package view

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Mode int

const (
	ModeSearch Mode = iota
	ModeDiff
)

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

type Model struct {
	mode        Mode
	filePath    string
	searchTerm  string
	replaceTerm string
	rawContent  string

	vpLeft  viewport.Model
	vpRight viewport.Model

	width   int
	height  int
	xOffset int
	ready   bool
}

func NewModel(mode Mode, filePath, search, replace, content string) Model {
	return Model{
		mode:        mode,
		filePath:    filePath,
		searchTerm:  search,
		replaceTerm: replace,
		rawContent:  content,
	}
}

func (m Model) GetMode() Mode {
	return m.mode
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "left", "h":
			maxOffset := computeMaxXOffset(m.mode, m.rawContent, m.searchTerm, m.replaceTerm, m.vpLeft.Width)
			m.xOffset = computeNextXOffset(m.xOffset, -1, maxOffset)
			m = m.refreshContent()
			return m, nil
		case "right", "l":
			maxOffset := computeMaxXOffset(m.mode, m.rawContent, m.searchTerm, m.replaceTerm, m.vpLeft.Width)
			m.xOffset = computeNextXOffset(m.xOffset, 1, maxOffset)
			m = m.refreshContent()
			return m, nil
		}

		if m.mode == ModeDiff {
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
		availableHeight := m.height - 1
		if availableHeight < 1 {
			availableHeight = 1
		}

		if !m.ready {
			m.vpLeft = viewport.New(m.width, availableHeight)
			m.vpRight = viewport.New(m.width, availableHeight)
			m.ready = true
		}

		if m.mode == ModeSearch {
			m.vpLeft.Width = m.width
			m.vpLeft.Height = availableHeight
		} else {
			colWidth := m.width / 2
			vpWidth := colWidth - 2
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

		maxOffset := computeMaxXOffset(m.mode, m.rawContent, m.searchTerm, m.replaceTerm, m.vpLeft.Width)
		m.xOffset = computeNextXOffset(m.xOffset, 0, maxOffset)
		m = m.refreshContent()
	}

	return m, nil
}

func (m Model) refreshContent() Model {
	searchHighlight := func(target string) string {
		return searchStyle.Render(target)
	}
	replaceHighlight := func(target string) string {
		return replaceStyle.Render(target)
	}

	if m.mode == ModeSearch {
		left := sliceAndHighlightContent(m.rawContent, m.searchTerm, m.xOffset, m.vpLeft.Width, searchHighlight)
		m.vpLeft.SetContent(left)
		return m
	}

	replaced := buildReplacedContent(m.rawContent, m.searchTerm, m.replaceTerm)
	left := sliceAndHighlightContent(m.rawContent, m.searchTerm, m.xOffset, m.vpLeft.Width, searchHighlight)
	right := sliceAndHighlightContent(replaced, m.replaceTerm, m.xOffset, m.vpRight.Width, replaceHighlight)
	m.vpLeft.SetContent(left)
	m.vpRight.SetContent(right)
	return m
}

func (m Model) View() string {
	if !m.ready {
		return "Initializing viewer..."
	}

	if m.mode == ModeSearch {
		header := headerStyle.Width(m.width).Render(fmt.Sprintf("File: %s | Search: '%s'", m.filePath, m.searchTerm))
		return lipgloss.JoinVertical(lipgloss.Left, header, m.vpLeft.View())
	}

	colWidth := m.width / 2
	titleStyle := headerStyle.Width(colWidth).MaxHeight(1)
	leftTitle := titleStyle.Render(fmt.Sprintf("Original: '%s'", m.searchTerm))
	rightTitle := titleStyle.Render(fmt.Sprintf("Replaced: '%s'", m.replaceTerm))

	leftPane := paneBorder.Width(colWidth - 2).Render(m.vpLeft.View())
	rightPane := paneBorder.Width(colWidth - 2).Render(m.vpRight.View())

	leftCol := lipgloss.JoinVertical(lipgloss.Left, leftTitle, leftPane)
	rightCol := lipgloss.JoinVertical(lipgloss.Left, rightTitle, rightPane)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightCol)
}
