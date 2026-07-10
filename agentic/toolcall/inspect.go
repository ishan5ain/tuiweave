package toolcall

import (
	"strconv"

	"github.com/ishansain/gotui/inspect"
)

const (
	ActionCancel = "cancel"
	ActionRetry  = "retry"
)

// Actions reports lifecycle operations available to the application or an
// inspection consumer.
func (b *Block) Actions() []inspect.Action {
	return []inspect.Action{
		{ID: ActionCancel, Label: "Cancel tool call", Enabled: b.status == StatusPending || b.status == StatusRunning},
		{ID: ActionRetry, Label: "Retry tool call", Enabled: b.status == StatusError || b.status == StatusCancelled},
	}
}

// ApplyAction performs a local semantic action and reports whether the ID was
// recognized. The application still decides whether policy permits the action.
func (b *Block) ApplyAction(id string) bool {
	switch id {
	case ActionCancel:
		if b.status != StatusPending && b.status != StatusRunning {
			return false
		}
		b.Cancel()
	case ActionRetry:
		if b.status != StatusError && b.status != StatusCancelled {
			return false
		}
		b.Retry()
	default:
		return false
	}
	return true
}

// Inspect reports identity, lifecycle, and non-rendered tool-call metadata.
// Output content is summarized by line count rather than copied into the tree.
func (b *Block) Inspect() inspect.Node {
	return inspect.Node{
		ID:      b.id,
		Kind:    "toolcall",
		Label:   b.Name,
		Status:  string(b.Lifecycle()),
		Actions: b.Actions(),
		Attributes: map[string]string{
			"summary":      b.Summary,
			"attempt":      strconv.Itoa(b.attempt),
			"output_lines": strconv.Itoa(len(b.output)),
		},
	}
}
