package textarea

import "strings"

const maxHistory = 100

type position struct {
	row int
	col int
}

type editState struct {
	lines     [][]rune
	row, col  int
	anchor    position
	hasAnchor bool
}

func (m Model) cursorPosition() position {
	return position{row: m.row, col: m.col}
}

// CursorPosition reports the cursor's logical rune position.
func (m Model) CursorPosition() Position {
	return Position{Row: m.row, Column: m.col}
}

func (m Model) clampPosition(p Position) position {
	if len(m.lines) == 0 {
		return position{}
	}
	row := max(0, min(p.Row, len(m.lines)-1))
	col := max(0, min(p.Column, len(m.lines[row])))
	return position{row: row, col: col}
}

func beforePosition(left, right position) bool {
	return left.row < right.row || (left.row == right.row && left.col < right.col)
}

// ReplaceRange replaces the normalized logical-rune range [start, end) with
// value. Coordinates are clamped to the current logical buffer and reversed
// ranges are ordered automatically. The replacement is one undoable edit and
// leaves the cursor immediately after the inserted value.
func (m *Model) ReplaceRange(start, end Position, value string) {
	first := m.clampPosition(start)
	last := m.clampPosition(end)
	if beforePosition(last, first) {
		first, last = last, first
	}
	m.resetTransientEditing()
	m.applyEdit(func() {
		m.setSelection(first, last)
		m.replaceSelection(value)
	})
}

func (m Model) selectionRange() (start, end position, ok bool) {
	if !m.hasAnchor {
		return position{}, position{}, false
	}
	cursor := m.cursorPosition()
	if m.anchor == cursor {
		return position{}, position{}, false
	}
	if m.anchor.row < cursor.row || (m.anchor.row == cursor.row && m.anchor.col < cursor.col) {
		return m.anchor, cursor, true
	}
	return cursor, m.anchor, true
}

// HasSelection reports whether the textarea has a non-empty selection.
func (m Model) HasSelection() bool {
	_, _, ok := m.selectionRange()
	return ok
}

// SelectedText returns the selected logical text, including newlines between
// selected logical lines. It returns an empty string when there is no
// selection.
func (m Model) SelectedText() string {
	start, end, ok := m.selectionRange()
	if !ok {
		return ""
	}
	if start.row == end.row {
		return string(m.lines[start.row][start.col:end.col])
	}

	var b strings.Builder
	b.WriteString(string(m.lines[start.row][start.col:]))
	for row := start.row + 1; row < end.row; row++ {
		b.WriteByte('\n')
		b.WriteString(string(m.lines[row]))
	}
	b.WriteByte('\n')
	b.WriteString(string(m.lines[end.row][:end.col]))
	return b.String()
}

// SelectAll selects the complete logical value and places the cursor at its
// end. Selection changes do not create undo entries.
func (m *Model) SelectAll() {
	m.resetTransientEditing()
	if m.Empty() {
		m.ClearSelection()
		return
	}
	m.anchor = position{}
	m.hasAnchor = true
	m.row = len(m.lines) - 1
	m.col = len(m.lines[m.row])
	m.ensureCursorVisible()
}

// ClearSelection removes the current selection without changing the cursor.
func (m *Model) ClearSelection() {
	m.hasAnchor = false
	m.resetTransientEditing()
}

func (m Model) selectionContains(row, col int) bool {
	start, end, ok := m.selectionRange()
	if !ok || row < start.row || row > end.row {
		return false
	}
	if row == start.row && col < start.col {
		return false
	}
	if row == end.row && col >= end.col {
		return false
	}
	return true
}

func (m *Model) beginSelection() {
	if !m.hasAnchor {
		m.anchor = m.cursorPosition()
		m.hasAnchor = true
	}
}

func (m *Model) setSelection(start, end position) {
	if start == end {
		m.hasAnchor = false
		return
	}
	m.anchor = start
	m.hasAnchor = true
	m.row, m.col = end.row, end.col
}

func (m *Model) deleteSelection() {
	start, end, ok := m.selectionRange()
	if !ok {
		return
	}
	if start.row == end.row {
		line := m.lines[start.row]
		m.lines[start.row] = append(append([]rune{}, line[:start.col]...), line[end.col:]...)
	} else {
		prefix := append([]rune{}, m.lines[start.row][:start.col]...)
		suffix := append([]rune{}, m.lines[end.row][end.col:]...)
		merged := append(prefix, suffix...)
		lines := make([][]rune, 0, len(m.lines)-(end.row-start.row))
		lines = append(lines, m.lines[:start.row]...)
		lines = append(lines, merged)
		lines = append(lines, m.lines[end.row+1:]...)
		m.lines = lines
	}
	m.row, m.col = start.row, start.col
	m.hasAnchor = false
}

func (m *Model) replaceSelection(s string) {
	m.deleteSelection()
	m.hasAnchor = false
	m.insertStringRaw(s)
}

func cloneLines(lines [][]rune) [][]rune {
	copyLines := make([][]rune, len(lines))
	for i, line := range lines {
		copyLines[i] = append([]rune{}, line...)
	}
	return copyLines
}

func (m Model) snapshot() editState {
	return editState{
		lines:     cloneLines(m.lines),
		row:       m.row,
		col:       m.col,
		anchor:    m.anchor,
		hasAnchor: m.hasAnchor,
	}
}

func sameLines(left, right [][]rune) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if string(left[i]) != string(right[i]) {
			return false
		}
	}
	return true
}

func sameState(left, right editState) bool {
	return left.row == right.row &&
		left.col == right.col &&
		left.anchor == right.anchor &&
		left.hasAnchor == right.hasAnchor &&
		sameLines(left.lines, right.lines)
}

func (m *Model) restore(state editState) {
	m.lines = cloneLines(state.lines)
	m.row = state.row
	m.col = state.col
	m.anchor = state.anchor
	m.hasAnchor = state.hasAnchor
	m.ensureCursorVisible()
}

func (m *Model) applyEdit(edit func()) {
	before := m.snapshot()
	// Update returns a value model. Detach the working copy before any edit can
	// replace a line or mutate a shared line slice from the previous model.
	m.lines = cloneLines(m.lines)
	edit()
	after := m.snapshot()
	if sameState(before, after) {
		return
	}
	m.undo = append(m.undo, before)
	if len(m.undo) > maxHistory {
		m.undo = m.undo[len(m.undo)-maxHistory:]
	}
	m.redo = nil
	m.ensureCursorVisible()
}

func (m *Model) resetHistory() {
	m.undo = nil
	m.redo = nil
}

// CanUndo reports whether Undo will restore an earlier edit state.
func (m Model) CanUndo() bool { return len(m.undo) > 0 }

// CanRedo reports whether Redo will restore an undone edit state.
func (m Model) CanRedo() bool { return len(m.redo) > 0 }

// Undo restores the most recent edit state, if one exists.
func (m *Model) Undo() {
	m.resetTransientEditing()
	if len(m.undo) == 0 {
		return
	}
	current := m.snapshot()
	last := len(m.undo) - 1
	target := m.undo[last]
	m.undo = m.undo[:last]
	m.redo = append(m.redo, current)
	if len(m.redo) > maxHistory {
		m.redo = m.redo[len(m.redo)-maxHistory:]
	}
	m.restore(target)
}

// Redo reapplies the most recently undone edit state, if one exists.
func (m *Model) Redo() {
	m.resetTransientEditing()
	if len(m.redo) == 0 {
		return
	}
	current := m.snapshot()
	last := len(m.redo) - 1
	target := m.redo[last]
	m.redo = m.redo[:last]
	m.undo = append(m.undo, current)
	if len(m.undo) > maxHistory {
		m.undo = m.undo[len(m.undo)-maxHistory:]
	}
	m.restore(target)
}
