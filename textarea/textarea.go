// Package textarea provides a multi-line text input for chat-style prompts:
// soft-wrapped display, a prompt gutter, and content-driven height.
//
// The model is logical lines ([][]rune); display is those lines soft-wrapped
// at the available cell width (grapheme-aware, deterministic). Enter inserts a
// newline — the app owns the send key. ContentHeight reports how many visual
// rows the content needs, so apps can grow the input:
//
//	h := min(4, m.input.ContentHeight())
//	layout.Vertical(layout.Fill(1), layout.Len(h)).Apply(...)
//
// Tier-2 editing adds logical-rune selections, select-all/replacement, bounded
// undo/redo, word-wise movement, richer word deletion, and a small model-owned
// kill ring. IME remains a separate follow-up work item.
package textarea

import (
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishan5ain/tuiweave"
)

// Model is a textarea component. Create one with New.
type Model struct {
	width, height     int
	lines             [][]rune
	row, col          int // cursor in logical coordinates (col in runes)
	yoff              int // first visible visual row
	focused           bool
	anchor            position
	hasAnchor         bool
	undo              []editState
	redo              []editState
	killRing          []string
	canCoalesceKill   bool
	lastKillDirection killDirection
	yankActive        bool
	yankIndex         int
	yankStart         position
	yankEnd           position

	// Prompt is rendered before the first visual row; continuation rows are
	// indented to match. Default "> ".
	Prompt string
	// Placeholder is shown, faint, while the value is empty.
	Placeholder string

	promptStyle      lipgloss.Style
	textStyle        lipgloss.Style
	selectionStyle   lipgloss.Style
	placeholderStyle lipgloss.Style
	cursorStyle      lipgloss.Style
}

// Position identifies a logical rune position in the textarea.
//
// Row and Column are zero-based. Column counts runes in a logical line, not
// terminal cells; applications should use CursorPosition and ReplaceRange
// when integrating completion or other app-owned editing behavior.
type Position struct {
	Row    int
	Column int
}

const (
	ActionFocus          = "focus"
	ActionBlur           = "blur"
	ActionClear          = "clear"
	ActionUndo           = "undo"
	ActionRedo           = "redo"
	ActionYank           = "yank"
	ActionSelectAll      = "select_all"
	ActionClearSelection = "clear_selection"
)

// New returns an empty textarea styled from the theme's roles.
func New(theme tuiweave.Theme) Model {
	return Model{
		lines:            [][]rune{{}},
		Prompt:           "> ",
		promptStyle:      lipgloss.NewStyle().Foreground(theme.Accent),
		textStyle:        lipgloss.NewStyle().Foreground(theme.Text),
		selectionStyle:   lipgloss.NewStyle().Foreground(theme.SelectionFg).Background(theme.SelectionBg),
		placeholderStyle: lipgloss.NewStyle().Foreground(theme.TextFaint),
		cursorStyle:      lipgloss.NewStyle().Foreground(theme.Text).Reverse(true),
	}
}

// SetTheme rebuilds all styles from the theme's roles, preserving
// non-style state such as the content and cursor position.
func (m *Model) SetTheme(theme tuiweave.Theme) {
	m.promptStyle = lipgloss.NewStyle().Foreground(theme.Accent)
	m.textStyle = lipgloss.NewStyle().Foreground(theme.Text)
	m.selectionStyle = lipgloss.NewStyle().Foreground(theme.SelectionFg).Background(theme.SelectionBg)
	m.placeholderStyle = lipgloss.NewStyle().Foreground(theme.TextFaint)
	m.cursorStyle = lipgloss.NewStyle().Foreground(theme.Text).Reverse(true)
}

// SetSize sets the box the textarea renders in.
func (m *Model) SetSize(width, height int) {
	m.width, m.height = width, height
	m.ensureCursorVisible()
}

// Focus makes the textarea accept keys and show its cursor.
func (m *Model) Focus() {
	m.focused = true
	m.ensureCursorVisible()
}

// Blur stops the textarea from accepting keys and hides the cursor.
func (m *Model) Blur() {
	m.focused = false
	m.ensureCursorVisible()
}

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
	m.setValue(s)
	m.hasAnchor = false
	m.resetHistory()
	m.resetKillState()
	m.ensureCursorVisible()
}

func (m *Model) setValue(s string) {
	raw := strings.Split(s, "\n")
	m.lines = make([][]rune, len(raw))
	for i, l := range raw {
		m.lines[i] = []rune(l)
	}
	m.row = len(m.lines) - 1
	m.col = len(m.lines[m.row])
}

// Reset clears the textarea.
func (m *Model) Reset() {
	m.lines = [][]rune{{}}
	m.row, m.col, m.yoff = 0, 0, 0
	m.hasAnchor = false
	m.resetHistory()
	m.resetKillState()
}

// Empty reports whether the textarea holds no text.
func (m Model) Empty() bool {
	return len(m.lines) == 1 && len(m.lines[0]) == 0
}

// InsertString inserts text at the cursor; \n starts new lines. This is also
// how apps insert a newline on a custom key (e.g. alt+enter).
func (m *Model) InsertString(s string) {
	if s == "" {
		return
	}
	m.resetTransientEditing()
	m.applyEdit(func() { m.replaceSelection(s) })
}

func (m *Model) insertStringRaw(s string) {
	first := true
	for line := range strings.SplitSeq(s, "\n") {
		if !first {
			m.splitLine()
		}
		m.insertRunes([]rune(line))
		first = false
	}
}

// visiblePrompt is the prompt clipped to the assigned width. A prompt
// grapheme that does not fit is omitted, which keeps narrow boxes bounded and
// leaves any remaining cell available for content.
func (m Model) visiblePrompt() string {
	return ansi.Truncate(m.Prompt, m.width, "")
}

// wrapWidth is the usable text width after the prompt gutter.
func (m Model) wrapWidth() int {
	return m.width - ansi.StringWidth(m.visiblePrompt())
}

// vrow is one visual (soft-wrapped) display row.
type vrow struct {
	line     int // logical line index
	startCol int // rune offset of this chunk within the line
	text     []rune
	width    int // terminal-cell width of text
}

// visualRows wraps the logical lines at the current cell width. Grapheme
// clusters stay together, so wide and combining characters are never split
// across display rows. A logical line whose final row exactly fills the
// width gets a trailing empty row when the cursor sits at its end.
func (m Model) visualRows() []vrow {
	w := m.wrapWidth()
	if w <= 0 {
		return nil
	}
	var rows []vrow
	for li, line := range m.lines {
		clusters := clustersOf(line)
		if len(clusters) == 0 {
			rows = append(rows, vrow{line: li})
			continue
		}
		start, rowWidth := 0, 0
		for _, cluster := range clusters {
			clusterWidth := cluster.width
			if clusterWidth > w {
				clusterWidth = w
			}
			if rowWidth > 0 && rowWidth+clusterWidth > w {
				rows = append(rows, vrow{
					line:     li,
					startCol: start,
					text:     line[start:cluster.start],
					width:    rowWidth,
				})
				start, rowWidth = cluster.start, 0
			}
			rowWidth += clusterWidth
		}
		rows = append(rows, vrow{
			line:     li,
			startCol: start,
			text:     line[start:],
			width:    rowWidth,
		})
		if m.focused && rowWidth == w && m.row == li && m.col == len(line) {
			rows = append(rows, vrow{line: li, startCol: len(line)})
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
			return i, min(r.width, cellWidth(r.text[:m.col-r.startCol]))
		}
	}
	return max(0, len(rows)-1), 0
}

// ContentHeight returns how many visual rows the content needs at the
// current width — use it to grow the input's layout slot.
func (m Model) ContentHeight() int {
	return max(1, len(m.visualRows()))
}

// Cursor returns the logical cursor position (row, col in runes).
func (m Model) Cursor() (row, col int) {
	return m.row, m.col
}

// SetCursor sets the logical cursor position. Used primarily in tests.
func (m *Model) SetCursor(row, col int) {
	if row < 0 {
		row = 0
	}
	if row >= len(m.lines) {
		row = len(m.lines) - 1
	}
	if col < 0 {
		col = 0
	}
	if col > len(m.lines[row]) {
		col = len(m.lines[row])
	}
	m.row = row
	m.col = col
	m.ensureCursorVisible()
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
		start := previousClusterStart(line, m.col)
		end := nextClusterEnd(line, start)
		m.lines[m.row] = append(append([]rune{}, line[:start]...), line[end:]...)
		m.col = start
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
		start := clusterStartAt(line, m.col)
		end := nextClusterEnd(line, start)
		m.lines[m.row] = append(append([]rune{}, line[:start]...), line[end:]...)
		m.col = start
		return
	}
	if m.row < len(m.lines)-1 {
		m.lines[m.row] = append(line, m.lines[m.row+1]...)
		m.lines = append(m.lines[:m.row+1], m.lines[m.row+2:]...)
	}
}

func (m *Model) moveLeft() {
	if m.col > 0 {
		m.col = previousClusterStart(m.lines[m.row], m.col)
	} else if m.row > 0 {
		m.row--
		m.col = len(m.lines[m.row])
	}
}

func (m *Model) moveRight() {
	if m.col < len(m.lines[m.row]) {
		m.col = nextClusterEnd(m.lines[m.row], m.col)
	} else if m.row < len(m.lines)-1 {
		m.row++
		m.col = 0
	}
}

func wordSpace(r rune) bool { return unicode.IsSpace(r) }

func (m Model) wordStartPosition() position {
	if m.col == 0 {
		if m.row == 0 {
			return position{}
		}
		return position{row: m.row - 1, col: len(m.lines[m.row-1])}
	}
	line := m.lines[m.row]
	i := m.col
	for i > 0 && wordSpace(line[i-1]) {
		i--
	}
	for i > 0 && !wordSpace(line[i-1]) {
		i--
	}
	return position{row: m.row, col: i}
}

func (m Model) wordEndPosition() position {
	line := m.lines[m.row]
	i := m.col
	for i < len(line) && !wordSpace(line[i]) {
		i++
	}
	for i < len(line) && wordSpace(line[i]) {
		i++
	}
	if i == len(line) && m.row < len(m.lines)-1 {
		return position{row: m.row + 1}
	}
	return position{row: m.row, col: i}
}

func (m *Model) moveWordLeft() {
	start := m.wordStartPosition()
	m.row, m.col = start.row, start.col
}

func (m *Model) moveWordRight() {
	end := m.wordEndPosition()
	m.row, m.col = end.row, end.col
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
	m.col = r.startCol + runeOffsetAtCell(r.text, vcol)
}

func (m *Model) moveWithSelection(extend bool, collapse int, move func()) {
	if !extend {
		if start, end, ok := m.selectionRange(); ok {
			if collapse < 0 {
				m.row, m.col = start.row, start.col
			} else if collapse > 0 {
				m.row, m.col = end.row, end.col
			}
			m.ClearSelection()
			if collapse != 0 {
				return
			}
		}
	}
	if extend {
		m.beginSelection()
	} else {
		m.ClearSelection()
	}
	move()
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

	commandMods := tea.ModCtrl | tea.ModAlt | tea.ModMeta | tea.ModSuper | tea.ModHyper
	if key.Text != "" && key.Mod&commandMods == 0 {
		m.InsertString(key.Text) // handles pasted newlines too
		return m, nil
	}

	keyName := key.String()
	isKill := keyName == "ctrl+u" || keyName == "ctrl+k" || keyName == "ctrl+w" ||
		keyName == "ctrl+delete" || keyName == "alt+backspace"
	if keyName != "alt+y" && !isKill {
		m.resetTransientEditing()
	}
	if isKill {
		m.yankActive = false
		m.yankIndex = 0
	}

	extend := key.Mod&tea.ModShift != 0
	switch keyName {
	case "enter":
		m.applyEdit(func() { m.replaceSelection("\n") })
	case "backspace":
		m.applyEdit(func() {
			if m.HasSelection() {
				m.deleteSelection()
				return
			}
			m.backspace()
		})
	case "delete":
		m.applyEdit(func() {
			if m.HasSelection() {
				m.deleteSelection()
				return
			}
			m.deleteForward()
		})
	case "ctrl+delete":
		m.applyEdit(func() {
			if m.HasSelection() {
				m.killSelection(killForward)
				return
			}
			start := m.cursorPosition()
			m.setSelection(start, m.wordEndPosition())
			m.killSelection(killForward)
		})
	case "left", "shift+left":
		m.moveWithSelection(extend, -1, m.moveLeft)
	case "right", "shift+right":
		m.moveWithSelection(extend, 1, m.moveRight)
	case "ctrl+left", "ctrl+shift+left":
		m.moveWithSelection(extend, -1, m.moveWordLeft)
	case "ctrl+right", "ctrl+shift+right":
		m.moveWithSelection(extend, 1, m.moveWordRight)
	case "up", "shift+up":
		m.moveWithSelection(extend, -1, func() { m.moveVertical(-1) })
	case "down", "shift+down":
		m.moveWithSelection(extend, 1, func() { m.moveVertical(1) })
	case "home", "shift+home":
		m.moveWithSelection(extend, -1, func() { m.col = 0 })
	case "end", "shift+end":
		m.moveWithSelection(extend, 1, func() { m.col = len(m.lines[m.row]) })
	case "ctrl+a", "ctrl+shift+a":
		m.SelectAll()
	case "ctrl+e", "ctrl+shift+e":
		m.moveWithSelection(extend, 1, func() { m.col = len(m.lines[m.row]) })
	case "ctrl+z":
		m.Undo()
	case "ctrl+y", "ctrl+shift+z":
		m.Redo()
	case "alt+y":
		m.Yank()
	case "ctrl+u":
		m.applyEdit(func() {
			if m.HasSelection() {
				m.killSelection(killBackward)
				return
			}
			start := position{row: m.row, col: 0}
			m.setSelection(start, m.cursorPosition())
			m.killSelection(killBackward)
		})
	case "ctrl+k":
		m.applyEdit(func() {
			if m.HasSelection() {
				m.killSelection(killForward)
				return
			}
			start := m.cursorPosition()
			end := start
			if start.col < len(m.lines[start.row]) {
				end.col = len(m.lines[start.row])
			} else if start.row < len(m.lines)-1 {
				end = position{row: start.row + 1, col: 0}
			}
			m.setSelection(start, end)
			m.killSelection(killForward)
		})
	case "ctrl+w":
		m.applyEdit(func() {
			if m.HasSelection() {
				m.killSelection(killBackward)
				return
			}
			start := m.wordStartPosition()
			m.setSelection(start, m.cursorPosition())
			m.killSelection(killBackward)
		})
	case "alt+backspace":
		m.applyEdit(func() {
			if m.HasSelection() {
				m.killSelection(killBackward)
				return
			}
			start := m.wordStartPosition()
			m.setSelection(start, m.cursorPosition())
			m.killSelection(killBackward)
		})
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
		return m.fitView(m.promptStyle.Render(m.visiblePrompt()))
	}

	if m.Empty() && m.Placeholder != "" {
		return m.fitView(m.renderPlaceholder(w))
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
	return m.fitView(strings.Join(out, "\n"))
}

func (m Model) fitView(view string) string {
	rows := strings.Split(view, "\n")
	if len(rows) > m.height {
		rows = rows[:m.height]
	}
	for i, row := range rows {
		width := ansi.StringWidth(row)
		if width > m.width {
			row = ansi.Truncate(row, m.width, "")
			width = ansi.StringWidth(row)
		}
		if width < m.width {
			row += strings.Repeat(" ", m.width-width)
		}
		rows[i] = row
	}
	for len(rows) < m.height {
		rows = append(rows, strings.Repeat(" ", m.width))
	}
	return strings.Join(rows, "\n")
}

func (m Model) renderRow(r vrow, vi, cursorIdx, vcol int) string {
	prompt := m.visiblePrompt()
	promptWidth := ansi.StringWidth(prompt)
	gutter := m.promptStyle.Render(prompt)
	if vi != 0 {
		gutter = strings.Repeat(" ", promptWidth)
	}

	var b strings.Builder
	b.WriteString(gutter)
	b.WriteString(m.renderText(r, vi, cursorIdx, vcol))

	used := promptWidth + r.width
	if m.focused && vi == cursorIdx && m.col == r.startCol+len(r.text) {
		used++
	}
	if pad := m.width - used; pad > 0 {
		b.WriteString(strings.Repeat(" ", pad))
	}
	return b.String()
}

func (m Model) renderText(r vrow, vi, cursorIdx, vcol int) string {
	const (
		normalRun = iota
		selectionRun
		cursorRun
	)
	styleFor := func(kind int) lipgloss.Style {
		switch kind {
		case selectionRun:
			return m.selectionStyle
		case cursorRun:
			return m.cursorStyle
		default:
			return m.textStyle
		}
	}

	var b strings.Builder
	runStart, runKind := 0, normalRun
	flush := func(end int) {
		if end > runStart {
			text := string(r.text[runStart:end])
			if width := cellWidth(r.text[runStart:end]); width > m.wrapWidth() {
				text = strings.Repeat(" ", m.wrapWidth())
			}
			b.WriteString(styleFor(runKind).Render(text))
		}
		runStart = end
	}
	cursorOffset := -1
	if m.focused && vi == cursorIdx {
		cursorOffset = m.col - r.startCol
	}
	for _, cluster := range clustersOf(r.text) {
		kind := normalRun
		for i := cluster.start; i < cluster.end; i++ {
			if m.selectionContains(r.line, r.startCol+i) {
				kind = selectionRun
				break
			}
		}
		if cursorOffset >= cluster.start && cursorOffset < cluster.end {
			kind = cursorRun
		}
		if cluster.start > runStart && kind != runKind {
			flush(cluster.start)
		}
		runKind = kind
	}
	flush(len(r.text))
	if cursorOffset == len(r.text) {
		b.WriteString(m.cursorStyle.Render(" "))
	}
	return b.String()
}

func (m Model) renderPlaceholder(w int) string {
	ph := ansi.Truncate(m.Placeholder, w, "")
	prompt := m.promptStyle.Render(m.visiblePrompt())
	if !m.focused {
		return prompt + m.placeholderStyle.Render(ph)
	}
	if ph == "" {
		return prompt + m.cursorStyle.Render(" ")
	}
	cluster, width := ansi.FirstGraphemeCluster(ph, ansi.GraphemeWidth)
	if width == 0 {
		return prompt + m.cursorStyle.Render(" ") + m.placeholderStyle.Render(ph)
	}
	remaining := ansi.Truncate(ph[len(cluster):], max(0, w-width), "")
	return prompt + m.cursorStyle.Render(cluster) + m.placeholderStyle.Render(remaining)
}
