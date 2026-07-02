package membership

// Status is where a membership sits in its lifecycle. It is set by the entity
// itself (at creation or by a transition command), never parsed from raw external
// input, so it has no constructor. Persistence <-> enum mapping is an
// infrastructure concern.
type Status int

const (
	// Active — the account is a live member of the organization. The starting
	// status: there is no invitation flow in the MVP, so a new membership is
	// active immediately.
	Active Status = iota
	// Suspended — temporarily blocked by an operator; can be reactivated. There
	// is no terminal status — removal is a repository Delete, not a lifecycle
	// state.
	Suspended
)

// String renders the status for logging and persistence mapping. It changes no
// behavior — it only names the enum value.
func (status Status) String() string {
	switch status {
	case Active:
		return "active"
	case Suspended:
		return "suspended"
	default:
		return "unknown"
	}
}
