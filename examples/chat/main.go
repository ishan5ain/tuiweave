// Command chat is the Phase 3 exit-criterion app: a mock-backed agentic
// chat session exercising every gotui/agentic component — streaming
// markdown, a tool call gated by a permission prompt, an inline diff, and a
// live usage bar.
//
//	go run ./examples/chat
//
// Type a message and press enter to run the scripted session. tab toggles
// focus between transcript and input, ctrl+c quits.
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishansain/gotui"
	"github.com/ishansain/gotui/agentic/chat"
	"github.com/ishansain/gotui/agentic/diffview"
	"github.com/ishansain/gotui/agentic/markdown"
	"github.com/ishansain/gotui/agentic/permission"
	"github.com/ishansain/gotui/agentic/toolcall"
	"github.com/ishansain/gotui/agentic/usagebar"
	"github.com/ishansain/gotui/focus"
	"github.com/ishansain/gotui/layout"
	"github.com/ishansain/gotui/overlay"
	"github.com/ishansain/gotui/textarea"
)

const responsePart1 = `Good question! The **layout** package splits terminal space with
constraints:

- ` + "`Len(n)`" + ` — fixed size
- ` + "`Fill(w)`" + ` — grows like flex

Let me verify the suite is green first:`

const responsePart2 = `All green. Here's the pattern in practice:

` + "```go\nlayout.Vertical(\n\tlayout.Fill(1), // transcript\n\tlayout.Len(1),  // input\n).Apply(area, &m.chat, &m.input)\n```" + `

That's the whole wiring — rectangles in, sized components out.`

const sampleDiff = `--- a/app.go
+++ b/app.go
@@ -1,3 +1,4 @@
 func (m app) layout(w, h int) {
-	m.chat.SetSize(w, h-1)
+	layout.Vertical(layout.Fill(1), layout.Len(1)).
+		Apply(layout.NewRect(0, 0, w, h), &m.chat, &m.input)
 }`

// step is the mock session's state machine.
type step int

const (
	stepIdle step = iota
	stepStreamingIntro
	stepAwaitPermission
	stepToolRunning
	stepStreamingRest
)

type tickMsg struct{}

func tick() tea.Cmd {
	return tea.Tick(60*time.Millisecond, func(time.Time) tea.Msg { return tickMsg{} })
}

type model struct {
	theme         gotui.Theme
	width, height int

	transcript chat.Model
	input      textarea.Model
	usage      usagebar.Model
	perm       permission.Model
	md         markdown.Renderer
	fm         focus.Manager

	step      step
	turn      int
	deltas    []string
	cur       *chat.Assistant
	tool      *toolcall.Block
	toolTicks int
	stats     usagebar.Stats
	showPerm  bool
}

func newModel() model {
	theme := gotui.Dark()
	m := model{
		theme:      theme,
		transcript: chat.New(theme),
		input:      textarea.New(theme),
		usage:      usagebar.New(theme),
		perm:       permission.New(theme),
		md:         markdown.NewRenderer(theme),
		fm:         focus.NewManager(2),
		stats:      usagebar.Stats{Model: "pi-large", TokensIn: 1200},
	}
	m.input.Placeholder = "ask the mock agent anything…"
	m.perm.ID = "bash"
	m.perm.Title = `Run "go test ./..."?`
	m.perm.Body = "The agent wants to run a shell command."
	intro := chat.NewText(theme, "mock session — responses are scripted")
	intro.SetID("system-1")
	m.transcript.Append(intro)
	m.fm.Set(1) // input first
	m.applyFocus()
	m.usage.SetStats(m.stats)
	return m
}

func (m *model) applyFocus() {
	m.fm.Apply(&m.transcript, &m.input)
}

// chunk splits s into ~4-word streaming deltas.
func chunk(s string) []string {
	words := strings.SplitAfter(s, " ")
	var out []string
	for i := 0; i < len(words); i += 4 {
		out = append(out, strings.Join(words[i:min(i+4, len(words))], ""))
	}
	return out
}

func (m *model) startSession(question string) tea.Cmd {
	m.turn++
	user := chat.NewUser(m.theme, question)
	user.SetID(fmt.Sprintf("user-%d", m.turn))
	m.transcript.Append(user)
	m.cur = chat.NewAssistant(m.theme, m.md)
	m.cur.SetID(fmt.Sprintf("assistant-%d", m.turn))
	m.transcript.Append(m.cur)
	m.deltas = chunk(responsePart1)
	m.step = stepStreamingIntro
	m.input.Reset()
	return tick()
}

// advance drives the state machine one tick.
func (m *model) advance() tea.Cmd {
	switch m.step {
	case stepStreamingIntro, stepStreamingRest:
		if len(m.deltas) > 0 {
			m.cur.Append(m.deltas[0])
			m.deltas = m.deltas[1:]
			m.bumpStats(3)
			m.transcript.Invalidate()
			return tick()
		}
		if m.step == stepStreamingIntro {
			m.step = stepAwaitPermission
			m.showPerm = true
			return nil // wait for the user's answer
		}
		m.cur.SetLifecycle(chat.StateComplete)
		m.step = stepIdle
		return nil

	case stepToolRunning:
		m.toolTicks++
		if m.toolTicks < 3 {
			return tick()
		}
		m.tool.SetStatus(toolcall.StatusSuccess)
		m.tool.AppendOutput("ok  \tgithub.com/ishansain/gotui/layout\t0.3s\nok  \tgithub.com/ishansain/gotui/snaptest\t0.2s")
		m.tool.Expanded = true
		theme, diff := m.theme, sampleDiff
		// part 2 streams into a fresh assistant cell so it appears after the
		// tool call and diff, not inside the pre-tool message.
		m.cur = chat.NewAssistant(theme, m.md)
		m.cur.SetID(fmt.Sprintf("assistant-%d-followup", m.turn))
		m.transcript.Append(
			chat.NewText(theme, "applied layout fix:"),
			chat.CellFunc(func(w int) string { return diffview.Sprint(theme, diff, w) }),
			m.cur,
		)
		m.deltas = chunk(responsePart2)
		m.step = stepStreamingRest
		return tick()
	}
	return nil
}

func (m *model) bumpStats(tokens int) {
	m.stats.TokensOut += tokens
	m.stats.Cost += float64(tokens) * 0.00002
	m.stats.ContextUsed = min(1, m.stats.ContextUsed+float64(tokens)*0.0004)
	m.usage.SetStats(m.stats)
}

// layout re-splits the window; the input slot grows with its content
// (up to 4 rows), chat-style.
func (m *model) layout() {
	if m.width <= 0 || m.height <= 0 {
		return
	}
	inputH := min(4, m.input.ContentHeight())
	var transcript, input, usage layout.Rect
	layout.Vertical(
		layout.Fill(1),
		layout.Len(inputH),
		layout.Len(1),
	).Split(layout.NewRect(0, 0, m.width, m.height)).
		Assign(&transcript, &input, &usage)
	m.transcript.SetSize(transcript.Dx(), transcript.Dy())
	m.input.SetSize(input.Dx(), input.Dy())
	m.usage.SetSize(usage.Dx(), usage.Dy())
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.layout()
		m.perm.SetSize(min(44, msg.Width-4), 12)

	case tickMsg:
		cmds = append(cmds, m.advance())

	case permission.ResultMsg:
		m.showPerm = false
		if msg.Option == "Deny" {
			if m.cur != nil {
				m.cur.SetLifecycle(chat.StateCancelled)
			}
			m.transcript.Append(chat.NewText(m.theme, "tool call denied — stopping here"))
			m.step = stepIdle
		} else {
			if m.cur != nil {
				m.cur.SetLifecycle(chat.StateComplete)
			}
			m.tool = toolcall.New(m.theme, "Bash", "go test ./...")
			m.tool.SetID(fmt.Sprintf("tool-%d", m.turn))
			m.tool.SetStatus(toolcall.StatusRunning)
			m.transcript.Append(m.tool)
			m.step = stepToolRunning
			m.toolTicks = 0
			cmds = append(cmds, tick())
		}

	case tea.KeyPressMsg:
		if m.showPerm {
			m.perm, cmd = m.perm.Update(msg)
			return m, cmd
		}
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.fm.Next()
			m.applyFocus()
		case "enter":
			// enter sends; alt+enter inserts a newline (below). The textarea
			// never sees a bare enter.
			if m.fm.Index() == 1 && m.step == stepIdle && strings.TrimSpace(m.input.Value()) != "" {
				cmds = append(cmds, m.startSession(strings.TrimSpace(m.input.Value())))
				m.layout()
			}
		case "alt+enter":
			if m.fm.Index() == 1 {
				m.input.InsertString("\n")
				m.layout()
			}
		default:
			m.transcript, cmd = m.transcript.Update(msg)
			cmds = append(cmds, cmd)
			m.input, cmd = m.input.Update(msg)
			cmds = append(cmds, cmd)
			m.layout() // input may have grown or shrunk
		}

	case tea.MouseWheelMsg:
		m.transcript, cmd = m.transcript.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// render composes the full frame; separate from View so tests can golden it.
func (m model) render() string {
	base := lipgloss.JoinVertical(lipgloss.Left,
		m.transcript.View(),
		m.input.View(),
		m.usage.View(),
	)
	if m.showPerm {
		base = overlay.Center(base, m.perm.View())
	}
	return base
}

func (m model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
