package main

import (
	"testing"

	"github.com/ishansain/gotui/snaptest"
)

func TestFrameExampleGolden(t *testing.T) {
	m := newModel()
	m.width, m.height = 68, 13
	snaptest.Snap(t, m.render())
}
