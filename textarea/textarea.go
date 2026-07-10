// Package textarea provides a multi-line text input for chat-style prompts:
// soft-wrapped display, a prompt gutter, and content-driven height.
//
// The model is logical lines ([][]rune); display is those lines soft-wrapped
// at the available width (character-level, deterministic). Enter inserts a
// newline — the app owns the send key. ContentHeight reports how many visual
// rows the content needs, so apps can grow the input:
//
//	h := min(4, m.input.ContentHeight())
//	layout.Vertical(layout.Fill(1), layout.Len(h)).Apply(...)
//
// Tier-1 editing only: arrows (up/down move by visual row), home/end,
// ctrl+a/e, backspace/delete (joining lines at boundaries), ctrl+u/k/w.
// Undo, kill ring, selections, and IME are future work.
package textarea

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishansain/gotui"
)

// Model is a textarea component. Create one with New.
type Model struct {
	width, height int
	lines         [][]rune
	row, col      int // cursor in logical coordinates (col in runes)
	yoff          int // first visible visual row
	focused       bool

	// Prompt is rendered before the first visual row; continuation rows are
	// indented to match. Default "> ".
	Prompt string
	// Placeholder is shown, faint, while the value is empty.
	Placeholder string

	promptStyle      lipgloss.Style
	textStyle        lipgloss.Style
	placeholderStyle lipgloss.Style
	cursorStyle      lipgloss.Style
}

const (
	ActionFocus = "focus"
	ActionBlur  = "blur"
	ActionClear = "clear"
)

// New returns an empty textarea styled from the theme's roles.
func New(theme gotui.Theme) Model {
	return Model{
		lines:            [][]rune{{}},
		Prompt:           "> ",
		promptStyle:      lipgloss.NewStyle().Foreground(theme.Accent),
		textStyle:        lipgloss.NewStyle().Foreground(theme.Text),
		placeholderStyle: lipgloss.NewStyle().Foreground(theme.TextFaint),
		cursorStyle:      lipgloss.NewStyle().Foreground(theme.Text).Reverse(true),
	}
}

// SetSize sets the box the textarea renders in.
func (m *Model) SetSize(width, height int) {
	m.width, m.height = width, height
	m.ensureCursorVisible()
}

// Focus makes the textarea accept keys and show its cursor.
func (m *Model) Focus() { m.focused = true }

// Blur stops the textarea from accepting keys and hides the cursor.
func (m *Model) Blur() { m.focused = false }

// Focused reports whether the textarea accepts keys.
func (m Model) Focused() bool { return m.focused }

// Value returns the text with lines joined by \n.
func (m Model) Value() string {
	parts := make([]string, len(m.lines))
	for i, l := range m.lines {
		parts[i] = string(l)
	}
	return strings.Join(parts, "\n")
}

// SetValue replaces the text and moves the cursor to its end.
func (m *Model) SetValue(s string) {
	raw := strings.Split(s, "\n")
	m.lines = make([][]rune, len(raw))
	for i, l := range raw {
		m.lines[i] = []rune(l)
	}
	m.row = len(m.lines) - 1
	m.col = len(m.lines[m.row])
	m.ensureCursorVisible()
}

// Reset clears the textarea.
func (m *Model) Reset() {
	m.lines = [][]rune{{}}
	m.row, m.col, m.yoff = 0, 0, 0
}

// Empty reports whether the textarea holds no text.
func (m Model) Empty() bool {
	return len(m.lines) == 1 && len(m.lines[0]) == 0
}

// InsertString inserts text at the cursor; \n starts new lines. This is also
// how apps insert a newline on a custom key (e.g. alt+enter).
func (m *Model) InsertString(s string) {
	first := true
	for line := range strings.SplitSeq(s, "\n") {
		if !first {
			m.splitLine()
		}
		m.insertRunes([]rune(line))
		first = false
	}
	m.ensureCursorVisible()
}

// wrapWidth is the usable text width after the prompt gutter.
func (m Model) wrapWidth() int {
	return m.width - len([]rune(m.Prompt))
}

// vrow is one visual (soft-wrapped) display row.
type vrow struct {
	line     int // logical line index
	startCol int // rune offset of this chunk within the line
	text     []rune
}

// visualRows wraps the logical lines at the current width. A logical line
// whose length is an exact positive multiple of the wrap width gets a
// trailing empty row when the cursor sits at its end, so the cursor always
// has a cell to occupy.
func (m Model) visualRows() []vrow {
	w := m.wrapWidth()
	if w <= 0 {
		return nil
	}
	var rows []vrow
	for li, line := range m.lines {
		n := len(line)
		for start := 0; ; start += w {
			end := min(start+w, n)
			rows = append(rows, vrow{line: li, startCol: start, text: line[start:end]})
			if end >= n {
				// Cursor at the exact end of a width-multiple line lives on
				// an extra empty row.
				if n > 0 && n%w == 0 && m.row == li && m.col == n {
					rows = append(rows, vrow{line: li, startCol: n, text: nil})
				}
				break
			}
		}
	}
	return rows
}

// cursorVisual locates the cursor within visualRows.
func (m Model) cursorVisual(rows []vrow) (idx, vcol int) {
	for i, r := range rows {
		if r.line != m.row {
			continue
		}
		if m.col >= r.startCol && m.col <= r.startCol+len(r.text) {
			// Prefer the next row when the cursor sits exactly on a wrap
			// boundary (start of the following chunk).
			if m.col == r.startCol+len(r.text) && i+1 < len(rows) && rows[i+1].line == m.row && rows[i+1].startCol == m.col {
				continue
			}
			return i, m.col - r.startCol
		}
	}
	return max(0, len(rows)-1), 0
}

// ContentHeight returns how many visual rows the content needs at the
// current width — use it to grow the input's layout slot.
func (m Model) ContentHeight() int {
	return max(1, len(m.visualRows()))
}

func (m *Model) ensureCursorVisible() {
	if m.height <= 0 || m.width <= 0 {
		return
	}
	rows := m.visualRows()
	idx, _ := m.cursorVisual(rows)
	if idx < m.yoff {
		m.yoff = idx
	}
	if idx >= m.yoff+m.height {
		m.yoff = idx - m.height + 1
	}
	m.yoff = max(0, min(m.yoff, max(0, len(rows)-m.height)))
}

func (m *Model) insertRunes(rs []rune) {
	if len(rs) == 0 {
		return
	}
	line := m.lines[m.row]
	m.lines[m.row] = append(line[:m.col], append(append([]rune{}, rs...), line[m.col:]...)...)
	m.col += len(rs)
}

// splitLine breaks the current line at the cursor (enter).
func (m *Model) splitLine() {
	line := m.lines[m.row]
	before := append([]rune{}, line[:m.col]...)
	after := append([]rune{}, line[m.col:]...)
	m.lines[m.row] = before
	m.lines = append(m.lines[:m.row+1], append([][]rune{after}, m.lines[m.row+1:]...)...)
	m.row++
	m.col = 0
}

func (m *Model) backspace() {
	if m.col > 0 {
		line := m.lines[m.row]
		m.lines[m.row] = append(line[:m.col-1], line[m.col:]...)
		m.col--
		return
	}
	if m.row > 0 {
		prev := m.lines[m.row-1]
		m.col = len(prev)
		m.lines[m.row-1] = append(prev, m.lines[m.row]...)
		m.lines = append(m.lines[:m.row], m.lines[m.row+1:]...)
		m.row--
	}
}

func (m *Model) deleteForward() {
	line := m.lines[m.row]
	if m.col < len(line) {
		m.lines[m.row] = append(line[:m.col], line[m.col+1:]...)
		return
	}
	if m.row < len(m.lines)-1 {
		m.lines[m.row] = append(line, m.lines[m.row+1]...)
		m.lines = append(m.lines[:m.row+1], m.lines[m.row+2:]...)
	}
}

func (m *Model) moveLeft() {
	if m.col > 0 {
		m.col--
	} else if m.row > 0 {
		m.row--
		m.col = len(m.lines[m.row])
	}
}

func (m *Model) moveRight() {
	if m.col < len(m.lines[m.row]) {
		m.col++
	} else if m.row < len(m.lines)-1 {
		m.row++
		m.col = 0
	}
}

// moveVertical moves the cursor by one visual row, clamping the column.
func (m *Model) moveVertical(delta int) {
	rows := m.visualRows()
	if len(rows) == 0 {
		return
	}
	idx, vcol := m.cursorVisual(rows)
	target := idx + delta
	if target < 0 || target >= len(rows) {
		return
	}
	r := rows[target]
	m.row = r.line
	m.col = r.startCol + min(vcol, len(r.text))
}

func (m *Model) killToLineEnd() {
	line := m.lines[m.row]
	if m.col < len(line) {
		m.lines[m.row] = line[:m.col]
		return
	}
	m.deleteForward() // at end of line, ctrl+k joins (emacs behavior)
}

func (m *Model) deleteWordBack() {
	line := m.lines[m.row]
	if m.col == 0 {
		m.backspace()
		return
	}
	i := m.col
	for i > 0 && line[i-1] == ' ' {
		i--
	}
	for i > 0 && line[i-1] != ' ' {
		i--
	}
	m.lines[m.row] = append(line[:i], line[m.col:]...)
	m.col = i
}

// Update handles editing keys while focused. Enter inserts a newline; the
// app decides what sends.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if next, handled := m.applyAction(msg); handled {
		return next, nil
	}
	if !m.focused {
		return m, nil
	}
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}

	if key.Text != "" && key.Mod == 0 {
		m.InsertString(key.Text) // handles pasted newlines too
		return m, nil
	}

	switch key.String() {
	case "enter":
		m.splitLine()
	case "backspace":
		m.backspace()
	case "delete":
		m.deleteForward()
	case "left":
		m.moveLeft()
	case "right":
		m.moveRight()
	case "up":
		m.moveVertical(-1)
	case "down":
		m.moveVertical(1)
	case "home", "ctrl+a":
		m.col = 0
	case "end", "ctrl+e":
		m.col = len(m.lines[m.row])
	case "ctrl+u":
		m.lines[m.row] = append([]rune{}, m.lines[m.row][m.col:]...)
		m.col = 0
	case "ctrl+k":
		m.killToLineEnd()
	case "ctrl+w":
		m.deleteWordBack()
	default:
		return m, nil
	}
	m.ensureCursorVisible()
	return m, nil
}

// View renders the visible window of visual rows with the prompt gutter and,
// when focused, the cursor.
func (m Model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	w := m.wrapWidth()
	if w <= 0 {
		return m.promptStyle.Render(m.Prompt)
	}

	if m.Empty() && m.Placeholder != "" {
		return m.renderPlaceholder(w)
	}

	rows := m.visualRows()
	cursorIdx, vcol := m.cursorVisual(rows)

	out := make([]string, 0, m.height)
	for i := range m.height {
		vi := m.yoff + i
		if vi >= len(rows) {
			out = append(out, strings.Repeat(" ", m.width))
			continue
		}
		out = append(out, m.renderRow(rows[vi], vi, cursorIdx, vcol))
	}
	return strings.Join(out, "\n")
}

func (m Model) renderRow(r vrow, vi, cursorIdx, vcol int) string {
	gutter := m.promptStyle.Render(m.Prompt)
	if vi != 0 {
		gutter = strings.Repeat(" ", len([]rune(m.Prompt)))
	}

	text := r.text
	var b strings.Builder
	b.WriteString(gutter)
	if m.focused && vi == cursorIdx {
		b.WriteString(m.textStyle.Render(string(text[:vcol])))
		if vcol < len(text) {
			b.WriteString(m.cursorStyle.Render(string(text[vcol])))
			b.WriteString(m.textStyle.Render(string(text[vcol+1:])))
		} else {
			b.WriteString(m.cursorStyle.Render(" "))
		}
	} else {
		b.WriteString(m.textStyle.Render(string(text)))
	}

	used := len([]rune(m.Prompt)) + len(text)
	if m.focused && vi == cursorIdx && vcol >= len(text) {
		used++
	}
	if pad := m.width - used; pad > 0 {
		b.WriteString(strings.Repeat(" ", pad))
	}
	return b.String()
}

func (m Model) renderPlaceholder(w int) string {
	ph := []rune(m.Placeholder)
	if len(ph) > w {
		ph = ph[:w]
	}
	prompt := m.promptStyle.Render(m.Prompt)
	if !m.focused {
		return prompt + m.placeholderStyle.Render(string(ph))
	}
	if len(ph) == 0 {
		return prompt + m.cursorStyle.Render(" ")
	}
	return prompt + m.cursorStyle.Render(string(ph[0])) + m.placeholderStyle.Render(string(ph[1:]))
}
