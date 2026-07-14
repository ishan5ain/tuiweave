// Package textinput provides a single-line text input with a prompt,
// placeholder, cursor, and grapheme-aware horizontal scrolling.
package textinput

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishan5ain/tuiweave"
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

const (
	ActionFocus = "focus"
	ActionBlur  = "blur"
	ActionClear = "clear"
)

// New returns an empty text input styled from the theme's roles.
func New(theme tuiweave.Theme) Model {
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
		m.insert([]rune(key.Text))
		return m, nil
	}

	switch key.String() {
	case "backspace":
		if m.pos > 0 {
			start := previousClusterStart(m.value, m.pos)
			end := nextClusterEnd(m.value, start)
			m.value = append(append([]rune{}, m.value[:start]...), m.value[end:]...)
			m.pos = start
		}
	case "delete":
		if m.pos < len(m.value) {
			start := clusterStartAt(m.value, m.pos)
			end := nextClusterEnd(m.value, start)
			m.value = append(append([]rune{}, m.value[:start]...), m.value[end:]...)
			m.pos = start
		}
	case "left":
		m.pos = previousClusterStart(m.value, m.pos)
	case "right":
		m.pos = nextClusterEnd(m.value, m.pos)
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
	value := make([]rune, 0, len(m.value)+len(runes))
	value = append(value, m.value[:m.pos]...)
	value = append(value, runes...)
	value = append(value, m.value[m.pos:]...)
	m.value = value
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
	m.value = append(append([]rune{}, m.value[:i]...), m.value[m.pos:]...)
	m.pos = i
}

// View renders the prompt, visible text window, and cursor.
func (m Model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	prompt := ansi.Truncate(m.Prompt, m.width, "")
	promptWidth := ansi.StringWidth(prompt)
	avail := m.width - promptWidth
	if avail <= 0 {
		return m.promptStyle.Render(prompt)
	}
	prompt = m.promptStyle.Render(prompt)

	if len(m.value) == 0 {
		return prompt + m.renderPlaceholder(avail)
	}

	start, end, focusStart, focusEnd, focusVisible := windowForCursor(m.value, m.pos, avail, m.focused)
	var b strings.Builder
	b.WriteString(prompt)
	if m.focused {
		if focusVisible && focusStart < focusEnd {
			b.WriteString(m.textStyle.Render(string(m.value[start:focusStart])))
			b.WriteString(m.cursorStyle.Render(string(m.value[focusStart:focusEnd])))
			b.WriteString(m.textStyle.Render(string(m.value[focusEnd:end])))
		} else {
			b.WriteString(m.textStyle.Render(string(m.value[start:end])))
			if m.pos >= len(m.value) || !focusVisible {
				b.WriteString(m.cursorStyle.Render(" "))
			}
		}
	} else {
		b.WriteString(m.textStyle.Render(string(m.value[start:end])))
	}
	return b.String()
}

func (m Model) renderPlaceholder(avail int) string {
	ph := []rune(m.Placeholder)
	if !m.focused {
		return m.placeholderStyle.Render(ansi.Truncate(string(ph), avail, ""))
	}
	if len(ph) == 0 || avail <= 0 {
		return m.cursorStyle.Render(" ")
	}
	clusters := clustersOf(ph)
	if len(clusters) == 0 || clusters[0].width > avail {
		return m.cursorStyle.Render(" ")
	}
	first := clusters[0]
	remaining := ansi.Truncate(string(ph[first.end:]), avail-first.width, "")
	return m.cursorStyle.Render(string(ph[first.start:first.end])) + m.placeholderStyle.Render(remaining)
}
