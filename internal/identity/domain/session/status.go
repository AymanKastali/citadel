package session

// Status is the session's place in its lifecycle. Like account's Status it is an
// enumerated type set by the entity itself (at Open, or by a transition command);
// it is never parsed from raw external input, so it has no New… constructor.
// Persistence <-> enum mapping is an infrastructure concern.
//
// Expired is a lazily-persisted convenience marker, not the source of truth for
// liveness: a session may sit as Active in storage after its ExpiresAt has passed,
// and a background sweep (or the next load) may later set the stored status to
// Expired. The live, always-correct answer is the entity's IsExpired(now) — which
// is why IsActive(now) checks both the stored status and a live expiry check, and
// Rotate guards on IsExpired(now) rather than on status == Expired.
type Status int

const (
	// Active — usable: may be rotated and accepted. The starting status at Open.
	Active Status = iota
	// Revoked — invalidated (logout, reuse detection, or an operator action);
	// terminal. Cannot be rotated or re-activated.
	Revoked
	// Expired — the stored status once the session is known to be past its
	// ExpiresAt. A lazily-persisted marker, not the source of truth for liveness.
	Expired
)

// String renders the status for logging and persistence mapping. It changes no
// behavior — it only names the enum value.
func (status Status) String() string {
	switch status {
	case Active:
		return "active"
	case Revoked:
		return "revoked"
	case Expired:
		return "expired"
	default:
		return "unknown"
	}
}
