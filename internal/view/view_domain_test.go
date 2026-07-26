package view

import "testing"

func TestComputeNextXOffset(t *testing.T) {
	tests := []struct {
		name      string
		current   int
		direction int
		maxOffset int
		want      int
	}{
		{name: "move right in range", current: 2, direction: 1, maxOffset: 5, want: 3},
		{name: "move left in range", current: 2, direction: -1, maxOffset: 5, want: 1},
		{name: "clamp at zero", current: 0, direction: -1, maxOffset: 5, want: 0},
		{name: "clamp at max", current: 5, direction: 1, maxOffset: 5, want: 5},
		{name: "direction zero keeps value", current: 3, direction: 0, maxOffset: 5, want: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeNextXOffset(tt.current, tt.direction, tt.maxOffset)
			if got != tt.want {
				t.Fatalf("computeNextXOffset() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestComputeMaxXOffset(t *testing.T) {
	tests := []struct {
		name         string
		mode         Mode
		raw          string
		rules        []Rule
		viewport     int
		wantMaxXOffs int
	}{
		{name: "search mode no scroll", mode: ModeSearch, raw: "hello", viewport: 10, wantMaxXOffs: 0},
		{name: "search mode needs scroll", mode: ModeSearch, raw: "hello world", viewport: 5, wantMaxXOffs: 6},
		{name: "diff mode uses longer replaced line", mode: ModeDiff, raw: "a", rules: []Rule{{Target: "a", Replacement: "alphabet"}}, viewport: 3, wantMaxXOffs: 5},
		{name: "zero viewport returns zero", mode: ModeSearch, raw: "abcdef", viewport: 0, wantMaxXOffs: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeMaxXOffset(tt.mode, tt.raw, tt.rules, tt.viewport)
			if got != tt.wantMaxXOffs {
				t.Fatalf("computeMaxXOffset() = %d, want %d", got, tt.wantMaxXOffs)
			}
		})
	}
}

func TestSliceAndHighlightContent(t *testing.T) {
	highlight := func(s string) string { return "[" + s + "]" }

	tests := []struct {
		name    string
		content string
		rules   []Rule
		xOffset int
		limit   int
		want    string
	}{
		{name: "slice ascii", content: "abcdef", xOffset: 2, limit: 3, want: "cde"},
		{name: "slice rune safe", content: "あいうえお", xOffset: 1, limit: 2, want: "いう"},
		{name: "offset beyond line", content: "abc", xOffset: 5, limit: 2, want: ""},
		{name: "highlight target", content: "foo bar foo", rules: []Rule{{Target: "foo"}}, xOffset: 0, limit: 20, want: "[foo] bar [foo]"},
		{name: "multi line", content: "abcdef\nuvwxyz", xOffset: 1, limit: 3, want: "bcd\nvwx"},
		{name: "highlight multiple rules", content: "foo bar", rules: []Rule{{Target: "foo"}, {Target: "bar"}}, xOffset: 0, limit: 20, want: "[foo] [bar]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sliceAndHighlightContent(tt.content, tt.rules, tt.xOffset, tt.limit, func(_ int, s string) string {
				return highlight(s)
			})
			if got != tt.want {
				t.Fatalf("sliceAndHighlightContent() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildReplacedContent_MultipleRules(t *testing.T) {
	rules := []Rule{
		{Target: "foo", Replacement: "bar"},
		{Target: "bar", Replacement: "baz"},
	}

	got := buildReplacedContent("foo", rules)
	if got != "baz" {
		t.Fatalf("buildReplacedContent() = %q, want %q", got, "baz")
	}
}
