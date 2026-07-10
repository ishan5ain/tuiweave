package progress

import (
	"strconv"

	"github.com/ishansain/gotui/inspect"
)

// Inspect reports the stable, app-independent state of the indicator.
func (m Model) Inspect() inspect.Node {
	return inspect.Node{
		Kind:   "progress",
		Bounds: inspect.Bounds{Width: m.width, Height: m.height},
		Label:  m.label,
		Status: statusName(m.status),
		Attributes: map[string]string{
			"percent":      strconv.FormatFloat(m.percent, 'f', -1, 64),
			"show_percent": strconv.FormatBool(m.showPercent),
		},
	}
}

func statusName(status Status) string {
	switch status {
	case StatusSuccess:
		return "success"
	case StatusWarning:
		return "warning"
	case StatusDanger:
		return "danger"
	case StatusInfo:
		return "info"
	default:
		return "normal"
	}
}
