package view

import (
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

func TestViewModel_LeftRightXOffset(t *testing.T) {
	m := NewModel(ModeSearch, "sample.txt", "", "", "0123456789")
	m.vpLeft = viewport.New(5, 3)
	m.ready = true
	m = m.refreshContent()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	next := updated.(Model)
	if next.xOffset != 1 {
		t.Fatalf("xOffset after right = %d, want 1", next.xOffset)
	}

	updated, _ = next.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	next = updated.(Model)
	if next.xOffset != 0 {
		t.Fatalf("xOffset after left = %d, want 0", next.xOffset)
	}
}

func TestViewModel_DiffScrollSync(t *testing.T) {
	m := NewModel(ModeDiff, "sample.txt", "", "", "a\nb\nc\nd\ne\nf\ng")
	m.vpLeft = viewport.New(10, 2)
	m.vpRight = viewport.New(10, 2)
	m.ready = true
	m = m.refreshContent()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	next := updated.(Model)

	if next.vpLeft.YOffset != next.vpRight.YOffset {
		t.Fatalf("diff YOffset not synced: left=%d right=%d", next.vpLeft.YOffset, next.vpRight.YOffset)
	}
}

func TestViewModel_WindowResizeSetsReadyAndDimensions(t *testing.T) {
	m := NewModel(ModeDiff, "sample.txt", "foo", "bar", "foo\nfoo")

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	next := updated.(Model)

	if !next.ready {
		t.Fatal("model should be ready after WindowSizeMsg")
	}
	if next.vpLeft.Width <= 0 || next.vpLeft.Height <= 0 {
		t.Fatalf("left viewport dimensions are invalid: width=%d height=%d", next.vpLeft.Width, next.vpLeft.Height)
	}
	if next.vpRight.Width <= 0 || next.vpRight.Height <= 0 {
		t.Fatalf("right viewport dimensions are invalid: width=%d height=%d", next.vpRight.Width, next.vpRight.Height)
	}
}

func TestNewModelWithRules_ClonesRules(t *testing.T) {
	rules := []Rule{{Target: "foo", Replacement: "bar"}}
	m := NewModelWithRules(ModeDiff, "sample.txt", rules, "foo")

	rules[0].Target = "changed"
	got := m.GetRules()
	if len(got) != 1 || got[0].Target != "foo" {
		t.Fatalf("rules were not cloned: %#v", got)
	}
}

func TestSearchAndReplaceSummary(t *testing.T) {
	rules := []Rule{
		{Target: "foo", Replacement: "bar"},
		{Target: "hello", Replacement: "world"},
	}

	if got := searchSummary(rules); got != "[foo, hello]" {
		t.Fatalf("searchSummary() = %q", got)
	}

	if got := replaceSummary(rules); got != "[foo->bar, hello->world]" {
		t.Fatalf("replaceSummary() = %q", got)
	}
}
