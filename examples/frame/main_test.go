package main

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/ishansain/gotui/snaptest"
)

func TestFrameExampleGolden(t *testing.T) {
	m := newModel()
	next, _ := m.Update(tea.WindowSizeMsg{Width: 68, Height: 13})
	m = next.(model)
	snaptest.Snap(t, m.render())
}
