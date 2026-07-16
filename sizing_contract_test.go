package tuiweave_test

import (
	"strings"
	"testing"

	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/agentic/chat"
	"github.com/ishan5ain/tuiweave/agentic/diffview"
	"github.com/ishan5ain/tuiweave/agentic/usagebar"
	"github.com/ishan5ain/tuiweave/autocomplete"
	"github.com/ishan5ain/tuiweave/button"
	"github.com/ishan5ain/tuiweave/help"
	"github.com/ishan5ain/tuiweave/layout"
	"github.com/ishan5ain/tuiweave/list"
	"github.com/ishan5ain/tuiweave/menu"
	"github.com/ishan5ain/tuiweave/palette"
	"github.com/ishan5ain/tuiweave/progress"
	"github.com/ishan5ain/tuiweave/statusbar"
	"github.com/ishan5ain/tuiweave/table"
	"github.com/ishan5ain/tuiweave/tabs"
	"github.com/ishan5ain/tuiweave/textarea"
	"github.com/ishan5ain/tuiweave/textinput"
	"github.com/ishan5ain/tuiweave/toggle"
	"github.com/ishan5ain/tuiweave/toolbar"
	"github.com/ishan5ain/tuiweave/viewport"
)

type widthBoundedComponent interface {
	SetSize(width, height int)
	SizeMode() layout.SizeMode
	View() string
}

func TestOneLineComponentsReportWidthBounded(t *testing.T) {
	theme := tuiweave.Dark()

	b := button.New(theme)
	b.SetLabel("Open")
	tg := toggle.New(theme)
	tg.SetLabel("Refresh")
	tb := tabs.New(theme)
	tb.SetTabs(tabs.Tab{ID: "overview", Label: "Overview"})
	tl := toolbar.New(theme)
	tl.SetItems(toolbar.Item{ID: "refresh", Label: "Refresh"})
	p := progress.New(theme)
	p.SetLabel("Indexing")
	h := help.New(theme)
	h.SetBindings(help.Binding{Key: "q", Desc: "quit"})
	s := statusbar.New(theme)
	s.SetLeft(statusbar.Segment{Text: "ready", Kind: statusbar.KindSuccess})
	ti := textinput.New(theme)
	ti.Placeholder = "command"
	u := usagebar.New(theme)
	u.SetStats(usagebar.Stats{Model: "model"})

	components := []struct {
		name      string
		component widthBoundedComponent
	}{
		{name: "button", component: &b},
		{name: "toggle", component: &tg},
		{name: "tabs", component: &tb},
		{name: "toolbar", component: &tl},
		{name: "progress", component: &p},
		{name: "help", component: &h},
		{name: "statusbar", component: &s},
		{name: "textinput", component: &ti},
		{name: "usagebar", component: &u},
	}

	for _, tt := range components {
		t.Run(tt.name, func(t *testing.T) {
			if got := layout.SizeModeOf(tt.component); got != layout.SizeWidthBounded {
				t.Fatalf("size mode = %d, want SizeWidthBounded", got)
			}

			tt.component.SetSize(12, 4)
			view := tt.component.View()
			if view == "" {
				t.Fatal("positive constraints rendered an empty view")
			}
			if got := lipgloss.Height(view); got != 1 {
				t.Fatalf("natural height = %d, want 1", got)
			}
			if got := lipgloss.Width(view); got > 12 {
				t.Fatalf("rendered width = %d, exceeds constraint 12", got)
			}

			tt.component.SetSize(0, 4)
			if got := tt.component.View(); got != "" {
				t.Fatalf("zero-width view = %q, want empty", got)
			}
			tt.component.SetSize(12, 0)
			if got := tt.component.View(); got != "" {
				t.Fatalf("zero-height view = %q, want empty", got)
			}
		})
	}
}

func TestMenuRemainsBoundedAcrossStates(t *testing.T) {
	theme := tuiweave.Dark()
	for _, populated := range []bool{false, true} {
		for _, focused := range []bool{false, true} {
			for _, width := range []int{1, 2, 7, 20} {
				for _, height := range []int{1, 2, 5} {
					m := menu.New(theme)
					if populated {
						m.SetItems(
							menu.Item{ID: "open", Label: "Open workspace"},
							menu.Item{ID: "delete", Label: "Delete", Disabled: true},
						)
					}
					if focused {
						m.Focus()
					}
					m.SetSize(width, height)

					if got := layout.SizeModeOf(&m); got != layout.SizeBounded {
						t.Fatalf("menu size mode = %d, want SizeBounded", got)
					}
					rows := strings.Split(m.View(), "\n")
					if len(rows) != height {
						t.Fatalf("populated=%v focused=%v size=%dx%d rendered %d rows", populated, focused, width, height, len(rows))
					}
					for row, line := range rows {
						if got := lipgloss.Width(line); got != width {
							t.Fatalf("populated=%v focused=%v size=%dx%d row %d width = %d", populated, focused, width, height, row, got)
						}
					}
				}
			}
		}
	}

	m := menu.New(theme)
	for _, size := range [][2]int{{0, 3}, {3, 0}, {-1, 3}, {3, -1}} {
		m.SetSize(size[0], size[1])
		if got := m.View(); got != "" {
			t.Fatalf("size %dx%d view = %q, want empty", size[0], size[1], got)
		}
	}
}

type boundedComponent interface {
	SetSize(width, height int)
	View() string
}

type focusableComponent interface {
	Focus()
	Blur()
}

func TestBoundedComponentsFillAssignedBox(t *testing.T) {
	theme := tuiweave.Dark()

	l := list.New(theme)
	l.SetItems("alpha", "界 combining e\u0301", "omega")
	le := list.New(theme)
	m := menu.New(theme)
	m.SetItems(menu.Item{ID: "open", Label: "Open"})
	v := viewport.New(theme)
	v.SetContent("alpha\n界 combining e\u0301\nomega")
	ve := viewport.New(theme)
	tb := table.New(theme)
	tb.SetColumns(table.Column{Title: "ID", Width: 2}, table.Column{Title: "Name"})
	tb.SetRows([]string{"1", "alpha"}, []string{"2", "界"})
	tbe := table.New(theme)
	tbe.SetColumns(table.Column{Title: "ID", Width: 2}, table.Column{Title: "Name"})
	ta := textarea.New(theme)
	ta.Placeholder = "type here"
	tap := textarea.New(theme)
	tap.SetValue("alpha\n界")
	p := palette.New(theme)
	p.SetItems(palette.Item{ID: "open", Label: "Open workspace"})
	pe := palette.New(theme)
	c := chat.New(theme)
	c.Append(chat.CellFunc(func(width int) string { return "transcript" }))
	ce := chat.New(theme)
	d := diffview.New(theme)
	d.SetDiff("@@ -1 +1 @@\n-old\n+new")
	de := diffview.New(theme)

	components := []struct {
		name      string
		component boundedComponent
	}{
		{name: "list/populated", component: &l},
		{name: "list/empty", component: &le},
		{name: "menu", component: &m},
		{name: "viewport/populated", component: &v},
		{name: "viewport/empty", component: &ve},
		{name: "table/populated", component: &tb},
		{name: "table/empty", component: &tbe},
		{name: "textarea/placeholder", component: &ta},
		{name: "textarea/populated", component: &tap},
		{name: "palette/populated", component: &p},
		{name: "palette/empty", component: &pe},
		{name: "chat/populated", component: &c},
		{name: "chat/empty", component: &ce},
		{name: "diffview/populated", component: &d},
		{name: "diffview/empty", component: &de},
	}

	for _, tt := range components {
		t.Run(tt.name, func(t *testing.T) {
			if got := layout.SizeModeOf(tt.component); got != layout.SizeBounded {
				t.Fatalf("size mode = %d, want SizeBounded", got)
			}
			for _, focused := range []bool{false, true} {
				if component, ok := tt.component.(focusableComponent); ok {
					if focused {
						component.Focus()
					} else {
						component.Blur()
					}
				}
				for _, width := range []int{1, 2, 7, 20} {
					for _, height := range []int{1, 2, 4} {
						tt.component.SetSize(width, height)
						assertExactBox(t, tt.component.View(), width, height, focused)
					}
				}
			}

			for _, size := range [][2]int{{0, 3}, {3, 0}, {-1, 3}, {3, -1}} {
				tt.component.SetSize(size[0], size[1])
				if got := tt.component.View(); got != "" {
					t.Fatalf("size %dx%d view = %q, want empty", size[0], size[1], got)
				}
			}
		})
	}
}

func TestAutocompleteReportsConditionalWidthBoundedSize(t *testing.T) {
	m := autocomplete.New(tuiweave.Dark())
	if got := layout.SizeModeOf(&m); got != layout.SizeWidthBounded {
		t.Fatalf("size mode = %d, want SizeWidthBounded", got)
	}
	m.SetSize(12, 3)
	if got := m.View(); got != "" {
		t.Fatalf("no-match view = %q, want empty", got)
	}

	m.SetItems(autocomplete.Item{ID: "open", Value: "open", Label: "Open"})
	assertExactBox(t, m.View(), 12, 3, false)
	m.SetQuery("missing")
	if got := m.View(); got != "" {
		t.Fatalf("filtered no-match view = %q, want empty", got)
	}
}

func assertExactBox(t *testing.T, view string, width, height int, focused bool) {
	t.Helper()
	rows := strings.Split(view, "\n")
	if len(rows) != height {
		t.Fatalf("focused=%v size=%dx%d rendered %d rows: %q", focused, width, height, len(rows), view)
	}
	for row, line := range rows {
		if got := lipgloss.Width(line); got != width {
			t.Fatalf("focused=%v size=%dx%d row %d width = %d: %q", focused, width, height, row, got, line)
		}
	}
}
