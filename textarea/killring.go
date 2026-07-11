package textarea

const maxKillRing = 20

// CanYank reports whether Yank has a most-recent killed value to insert.
func (m Model) CanYank() bool { return len(m.killRing) > 0 }

// Yank inserts the most recently killed value at the cursor, replacing an
// active selection. Yank is an edit and can be undone independently.
func (m *Model) Yank() {
	if !m.CanYank() {
		return
	}
	text := m.killRing[0]
	m.applyEdit(func() { m.replaceSelection(text) })
}

func (m *Model) addKill(text string) {
	if text == "" {
		return
	}
	m.killRing = append([]string{text}, m.killRing...)
	if len(m.killRing) > maxKillRing {
		m.killRing = m.killRing[:maxKillRing]
	}
}

func (m *Model) killSelection() {
	if !m.HasSelection() {
		return
	}
	text := m.SelectedText()
	m.deleteSelection()
	m.addKill(text)
}
