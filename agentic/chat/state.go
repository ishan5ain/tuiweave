package chat

// State is the lifecycle state of a transcript cell.
type State string

const (
	StatePending   State = "pending"
	StateStreaming State = "streaming"
	StateRunning   State = "running"
	StateComplete  State = "complete"
	StateFailed    State = "failed"
	StateCancelled State = "cancelled"
)

// CellIdentity is an optional stable identity for a transcript cell. Existing
// CellFunc adapters remain valid without identity or lifecycle metadata.
type CellIdentity interface {
	CellID() string
}

// CellKind is an optional stable semantic kind for inspection consumers.
type CellKind interface {
	CellKind() string
}

// CellLifecycle is an optional lifecycle report for inspection consumers.
type CellLifecycle interface {
	Lifecycle() State
}
