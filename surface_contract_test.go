package tuiweave_test

import (
	"strings"
	"testing"

	lipgloss "charm.land/lipgloss/v2"

	"github.com/ishan5ain/tuiweave"
	"github.com/ishan5ain/tuiweave/action"
	"github.com/ishan5ain/tuiweave/agentic/permission"
	"github.com/ishan5ain/tuiweave/autocomplete"
	"github.com/ishan5ain/tuiweave/button"
	"github.com/ishan5ain/tuiweave/dialog"
	"github.com/ishan5ain/tuiweave/frame"
	"github.com/ishan5ain/tuiweave/palette"
	"github.com/ishan5ain/tuiweave/progress"
	"github.com/ishan5ain/tuiweave/snaptest"
	"github.com/ishan5ain/tuiweave/statusbar"
	"github.com/ishan5ain/tuiweave/tabs"
	"github.com/ishan5ain/tuiweave/toggle"
	"github.com/ishan5ain/tuiweave/toolbar"
)

func TestInheritedRaisedSurfaceGolden(t *testing.T) {
	theme := tuiweave.Dark()
	theme.SurfaceRaised = lipgloss.NoColor{}

	nav := tabs.New(theme)
	nav.SetTabs(
		tabs.Tab{ID: "services", Label: "Services"},
		tabs.Tab{ID: "jobs", Label: "Jobs"},
	)
	nav.SetSize(28, 1)
	nav.Focus()

	bar := statusbar.New(theme)
	bar.SetLeft(
		statusbar.Segment{Text: "tuiweave", Kind: statusbar.KindAccent},
		statusbar.Segment{Text: " ready", Kind: statusbar.KindNormal},
	)
	bar.SetRight(statusbar.Segment{Text: "healthy", Kind: statusbar.KindSuccess})
	bar.SetSize(28, 1)

	completions := autocomplete.New(theme)
	completions.SetItems(
		autocomplete.Item{ID: "status", Value: "git status", Label: "git status", Description: "inspect"},
		autocomplete.Item{ID: "test", Value: "go test ./...", Label: "go test ./...", Description: "verify"},
	)
	completions.SetSize(28, 2)
	completions.Focus()

	actions := []action.Item{
		{ID: "open", Label: "Open", Description: "open workspace"},
		{ID: "refresh", Label: "Refresh", Description: "refresh state"},
	}

	tools := toolbar.New(theme)
	tools.SetItems(actions...)
	tools.SetSize(28, 1)
	tools.Focus()

	setting := toggle.New(theme)
	setting.SetLabel("Auto-refresh")
	setting.SetChecked(true)
	setting.SetSize(28, 1)

	task := progress.New(theme)
	task.SetLabel("Indexing")
	task.SetPercent(0.72)
	task.SetStatus(progress.StatusSuccess)
	task.SetSize(28, 1)

	commands := palette.New(theme)
	commands.SetItems(actions...)
	commands.SetSize(28, 3)
	commands.Focus()

	open := button.New(theme)
	open.SetLabel("Open workspace")
	open.SetSize(28, 1)

	confirm := dialog.New(theme)
	confirm.Title = "Delete workspace?"
	confirm.Body = "This cannot be undone."
	confirm.SetSize(32, 8)

	approval := permission.New(theme)
	approval.Title = "Run tests?"
	approval.Body = "The command reads this repository."
	approval.SetProvenance(permission.Provenance{Tool: "shell", Operation: "execute"})
	approval.SetSize(36, 12)

	view := strings.Join([]string{
		"panel",
		frame.Panel(theme, "Application-owned body", 28, frame.PanelOptions{Title: "Services", Padding: 1}),
		"badge",
		frame.Badge(theme, "healthy", frame.BadgeSuccess),
		"tabs",
		nav.View(),
		"statusbar",
		bar.View(),
		"autocomplete",
		completions.View(),
		"toolbar",
		tools.View(),
		"toggle",
		setting.View(),
		"progress",
		task.View(),
		"palette",
		commands.View(),
		"button",
		open.View(),
		"dialog",
		confirm.View(),
		"permission",
		approval.View(),
	}, "\n")

	snaptest.Snap(t, view)
	snaptest.SnapCells(t, view, snaptest.WithRoles(theme))
}
