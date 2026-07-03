// Package textinput provides a single-line text input with a prompt,
// placeholder, cursor, and horizontal scrolling.
package textinput

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishansain/gotui"
)

// Model is a text input component. Create one with New.
type Model struct {
	width, height int
	value         []rune
	pos           int // cursor position in runes
	focused       bool

	// Prompt is rendered before the editable area. Default "> ".
	Prompt string
	// Placeholder is shown, faint, while the value is empty.
	Placeholder string

	promptStyle      lipgloss.Style
	textStyle        lipgloss.Style
	placeholderStyle lipgloss.Style
	cursorStyle      lipgloss.Style
}

// New returns an empty text input styled from the theme's roles.
func New(theme gotui.Theme) Model {
	return Model{
		Prompt:           "> ",
		promptStyle:      lipgloss.NewStyle().Foreground(theme.Accent),
		textStyle:        lipgloss.NewStyle().Foreground(theme.Text),
		placeholderStyle: lipgloss.NewStyle().Foreground(theme.TextFaint),
		cursorStyle:      lipgloss.NewStyle().Foreground(theme.Text).Reverse(true),
	}
}

// SetSize sets the box the input renders in; the input is one line tall.
func (m *Model) SetSize(width, height int) {
	m.width, m.height = width, height
}

// Focus makes the input accept keys and show its cursor.
func (m *Model) Focus() { m.focused = true }

// Blur stops the input from accepting keys and hides the cursor.
func (m *Model) Blur() { m.focused = false }

// Focused reports whether the input accepts keys.
func (m Model) Focused() bool { return m.focused }

// Value returns the current text.
func (m Model) Value() string { return string(m.value) }

// SetValue replaces the text and moves the cursor to its end.
func (m *Model) SetValue(s string) {
	m.value = []rune(s)
	m.pos = len(m.value)
}

// Reset clears the input.
func (m *Model) Reset() {
	m.value = nil
	m.pos = 0
}

// Update handles editing keys while focused: printable text inserts at the
// cursor; backspace/delete remove; left/right/home/end (and ctrl+a/ctrl+e)
// move; ctrl+u clears before the cursor, ctrl+k after, ctrl+w deletes the
// previous word.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focused {
		return m, nil
	}
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}

	if key.Text != "" && key.Mod == 0 {
		m.insert([]rune(key.Text))
		return m, nil
	}

	switch key.String() {
	case "backspace":
		if m.pos > 0 {
			m.value = append(m.value[:m.pos-1], m.value[m.pos:]...)
			m.pos--
		}
	case "delete":
		if m.pos < len(m.value) {
			m.value = append(m.value[:m.pos], m.value[m.pos+1:]...)
		}
	case "left":
		m.pos = max(0, m.pos-1)
	case "right":
		m.pos = min(len(m.value), m.pos+1)
	case "home", "ctrl+a":
		m.pos = 0
	case "end", "ctrl+e":
		m.pos = len(m.value)
	case "ctrl+u":
		m.value = append([]rune{}, m.value[m.pos:]...)
		m.pos = 0
	case "ctrl+k":
		m.value = m.value[:m.pos]
	case "ctrl+w":
		m.deleteWordBack()
	}
	return m, nil
}

func (m *Model) insert(runes []rune) {
	m.value = append(m.value[:m.pos], append(runes, m.value[m.pos:]...)...)
	m.pos += len(runes)
}

func (m *Model) deleteWordBack() {
	i := m.pos
	for i > 0 && m.value[i-1] == ' ' {
		i--
	}
	for i > 0 && m.value[i-1] != ' ' {
		i--
	}
	m.value = append(m.value[:i], m.value[m.pos:]...)
	m.pos = i
}

// View renders the prompt, visible text window, and cursor.
func (m Model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	prompt := m.promptStyle.Render(m.Prompt)
	avail := m.width - len([]rune(m.Prompt))
	if avail <= 0 {
		return prompt
	}

	if len(m.value) == 0 {
		return prompt + m.renderPlaceholder(avail)
	}

	// Keep the cursor inside the visible window. The cursor may sit one past
	// the last rune, so the window holds avail-1 runes when focused. The
	// window is derived from the cursor each render: content start until the
	// cursor passes the right edge, then cursor pinned at the right edge.
	visible := avail
	if m.focused {
		visible--
	}
	off := max(0, m.pos-visible)
	off = min(off, max(0, len(m.value)-visible))

	end := min(len(m.value), off+visible)
	var b strings.Builder
	b.WriteString(prompt)
	if m.focused {
		before := m.value[off:m.pos]
		b.WriteString(m.textStyle.Render(string(before)))
		if m.pos < len(m.value) {
			b.WriteString(m.cursorStyle.Render(string(m.value[m.pos])))
			if m.pos+1 <= end {
				b.WriteString(m.textStyle.Render(string(m.value[m.pos+1 : end])))
			}
		} else {
			b.WriteString(m.cursorStyle.Render(" "))
		}
	} else {
		b.WriteString(m.textStyle.Render(string(m.value[off:end])))
	}
	return b.String()
}

func (m Model) renderPlaceholder(avail int) string {
	ph := []rune(m.Placeholder)
	if len(ph) > avail {
		ph = ph[:avail]
	}
	if !m.focused {
		return m.placeholderStyle.Render(string(ph))
	}
	if len(ph) == 0 {
		return m.cursorStyle.Render(" ")
	}
	return m.cursorStyle.Render(string(ph[0])) + m.placeholderStyle.Render(string(ph[1:]))
}
