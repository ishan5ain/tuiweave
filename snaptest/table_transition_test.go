package snaptest

import (
	"bytes"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/table"
)

func TestTableRegionalIndicatorWcWidthTransition(t *testing.T) {
	tb := table.New(tuiweave.Dark())
	tb.SetSize(24, 4)
	tb.SetColumns(table.Column{Title: "Country", Width: 16})
	tb.SetRows(
		[]string{"🇨🇦 Canada"},
		[]string{"🇯🇵 Japan"},
	)
	tb.Focus()

	var output bytes.Buffer
	renderer := uv.NewTerminalRenderer(&output, []string{
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
	})
	renderer.SetFullscreen(true)
	renderer.Resize(24, 4)

	states := []struct {
		name string
		msg  tea.Msg
	}{
		{name: "initial"},
		{name: "selected japan", msg: tea.KeyPressMsg{Code: tea.KeyDown}},
		{name: "selected canada again", msg: tea.KeyPressMsg{Code: tea.KeyUp}},
	}
	for _, state := range states {
		if state.msg != nil {
			tb, _ = tb.Update(state.msg)
		}

		t.Run(state.name, func(t *testing.T) {
			buf := renderToGridMethod(tb.View(), ansi.WcWidth)
			assertRegionalIndicatorCell(t, buf, 2, "🇨🇦", "C")
			assertRegionalIndicatorCell(t, buf, 3, "🇯🇵", "J")

			output.Reset()
			renderer.Render(buf.RenderBuffer)
			if err := renderer.Flush(); err != nil {
				t.Fatalf("flush renderer: %v", err)
			}
			frame := output.String()
			for _, flag := range []string{"🇨🇦", "🇯🇵"} {
				if !strings.Contains(frame, flag) {
					t.Fatalf("renderer update lost %q: %q", flag, frame)
				}
			}
		})
	}
}

func assertRegionalIndicatorCell(t *testing.T, buf uv.ScreenBuffer, row int, flag, firstLetter string) {
	t.Helper()

	cell := buf.CellAt(0, row)
	if cell == nil {
		t.Fatalf("row %d has no flag cell", row)
	}
	if cell.Content != flag || cell.Width != 2 {
		t.Fatalf("row %d flag cell = {%q width=%d}, want {%q width=2}", row, cell.Content, cell.Width, flag)
	}
	continuation := buf.CellAt(1, row)
	if continuation == nil || continuation.Width != 0 {
		t.Fatalf("row %d flag continuation = %#v, want width-zero placeholder", row, continuation)
	}
	letter := buf.CellAt(3, row)
	if letter == nil || letter.Content != firstLetter {
		t.Fatalf("row %d first country letter = %#v, want %q at column 3", row, letter, firstLetter)
	}
}
