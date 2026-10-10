package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/inspect"
	"github.com/ishan5ain/tuiweave/snaptest"
)

// Exercise the actual runtime as well as the direct Update/scenario tests:
// startup, resize, navigation, the app's quit command, and final rendering.
func TestConsumerProgramLifecycleWithoutInput(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var output bytes.Buffer
	initialSize := make(chan struct{}, 1)
	p := tea.NewProgram(newModel(),
		tea.WithContext(ctx),
		tea.WithInput(nil),
		tea.WithOutput(&output),
		tea.WithEnvironment([]string{"TERM=xterm-256color", "TERM_PROGRAM=ghostty"}),
		tea.WithWindowSize(48, 8),
		tea.WithFilter(func(_ tea.Model, msg tea.Msg) tea.Msg {
			if size, ok := msg.(tea.WindowSizeMsg); ok && size.Width == 48 && size.Height == 8 {
				select {
				case initialSize <- struct{}{}:
				default:
				}
			}
			return msg
		}),
		tea.WithoutSignals(),
		tea.WithoutCatchPanics(),
	)
	type result struct {
		model tea.Model
		err   error
	}
	done := make(chan result, 1)
	go func() {
		m, err := p.Run()
		done <- result{m, err}
	}()
	// The initial resize is sent asynchronously by Bubble Tea. Wait for it
	// before injecting another resize so it cannot overwrite the test's size.
	select {
	case <-initialSize:
	case <-ctx.Done():
		t.Fatalf("waiting for initial size: %v", ctx.Err())
	}
	p.Send(tea.WindowSizeMsg{Width: 32, Height: 6})
	p.Send(tea.KeyPressMsg{Code: tea.KeyDown})
	p.Send(tea.KeyPressMsg{Code: 'q', Text: "q"})
	r := <-done
	if r.err != nil {
		t.Fatalf("program lifecycle: %v", r.err)
	}
	m := r.model.(model)
	if m.width != 32 || m.height != 6 || m.items.SelectedItem() != "focus" {
		t.Fatalf("final state: size=%dx%d selected=%q", m.width, m.height, m.items.SelectedItem())
	}
	if !strings.Contains(ansi.Strip(output.String()), "Components") {
		t.Fatal("program did not render the consumer frame")
	}
	for _, query := range []string{ansi.RequestModeSynchronizedOutput, ansi.RequestModeUnicodeCore} {
		if strings.Contains(output.String(), query) {
			t.Fatalf("program with input disabled emitted terminal query %q", query)
		}
	}
}

func TestConsumerPublicAPIView(t *testing.T) {
	m := sized(t)
	snaptest.Snap(t, m.render())
	snaptest.SnapCells(t, m.render(), snaptest.WithRoles(tuiweave.Dark()))
}

func TestConsumerPublicAPIInteraction(t *testing.T) {
	m := sized(t)
	next, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	if cmd != nil {
		t.Fatal("navigation emitted an unexpected command")
	}
	m = next.(model)
	if got := m.items.SelectedItem(); got != "focus" {
		t.Fatalf("selected item = %q, want focus", got)
	}
}

func TestConsumerInspectionGolden(t *testing.T) {
	m := sized(t)
	data, err := inspect.Marshal(m.Inspect())
	if err != nil {
		t.Fatal(err)
	}
	snaptest.Snap(t, string(data))
}

func TestConsumerQualifiedSemanticActionRouting(t *testing.T) {
	m := sized(t)
	before := m.items.Selected()
	next, cmd := m.Update(inspect.Invoke("items.previous"))
	if cmd != nil {
		t.Fatal("disabled previous action emitted a command")
	}
	m = next.(model)
	if m.items.Selected() != before {
		t.Fatalf("disabled previous action changed selection to %d", m.items.Selected())
	}

	next, cmd = m.Update(inspect.Invoke("items.next"))
	if cmd != nil {
		t.Fatal("list next action emitted a command")
	}
	m = next.(model)
	if m.items.SelectedItem() != "focus" {
		t.Fatalf("qualified next selected %q", m.items.SelectedItem())
	}

	data, err := inspect.Marshal(m.Inspect())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"id": "items.next"`) {
		t.Fatalf("updated inspection missing qualified next action:\n%s", data)
	}
}

func sized(t *testing.T) model {
	t.Helper()
	m := newModel()
	next, cmd := m.Update(tea.WindowSizeMsg{Width: 48, Height: 8})
	if cmd != nil {
		t.Fatal("window sizing emitted an unexpected command")
	}
	return next.(model)
}
