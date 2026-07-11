package textarea

type killDirection uint8

const (
	killBackward killDirection = iota
	killForward
)

const maxKillRing = 20

// CanYank reports whether Yank has killed text to insert.
func (m Model) CanYank() bool { return len(m.killRing) > 0 }

func (m *Model) resetTransientEditing() {
	m.canCoalesceKill = false
	m.yankActive = false
	m.yankIndex = 0
	m.yankStart = position{}
	m.yankEnd = position{}
}

func (m *Model) resetKillState() {
	m.killRing = nil
	m.resetTransientEditing()
}

// Yank inserts the most recently killed value at the cursor, replacing an
// active selection. Repeating Yank rotates through older kills without
// appending duplicate text. Each replacement is independently undoable.
func (m *Model) Yank() {
	m.canCoalesceKill = false
	if !m.CanYank() {
		m.yankActive = false
		m.yankIndex = 0
		return
	}

	index := 0
	start := m.cursorPosition()
	end := start
	rotating := m.yankActive
	if rotating {
		if len(m.killRing) < 2 {
			return
		}
		index = (m.yankIndex + 1) % len(m.killRing)
		start, end = m.yankStart, m.yankEnd
	} else if selectionStart, _, ok := m.selectionRange(); ok {
		start = selectionStart
	}
	text := m.killRing[index]
	m.applyEdit(func() {
		if rotating {
			m.setSelection(start, end)
		}
		m.replaceSelection(text)
	})
	m.yankActive = true
	m.yankIndex = index
	m.yankStart = start
	m.yankEnd = m.cursorPosition()
}

func (m *Model) addKill(text string, direction killDirection) {
	if text == "" {
		return
	}
	ring := append([]string(nil), m.killRing...)
	if m.canCoalesceKill && len(ring) > 0 && m.lastKillDirection == direction {
		if direction == killBackward {
			ring[0] = text + ring[0]
		} else {
			ring[0] += text
		}
	} else {
		ring = append([]string{text}, ring...)
	}
	m.killRing = ring
	if len(m.killRing) > maxKillRing {
		m.killRing = m.killRing[:maxKillRing]
	}
	m.canCoalesceKill = true
	m.lastKillDirection = direction
	m.yankActive = false
	m.yankIndex = 0
}

func (m *Model) killSelection(direction killDirection) {
	if !m.HasSelection() {
		m.canCoalesceKill = false
		return
	}
	text := m.SelectedText()
	m.deleteSelection()
	m.addKill(text, direction)
}
