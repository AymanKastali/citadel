package session

// Kind discriminates what mechanism a session backs. Like Status it is an
// enumerated type set by the entity (chosen at Open from the caller's params,
// itself produced from config/wiring — not parsed from a wire string here), so it
// has no New… constructor. Persistence <-> enum mapping is an infrastructure
// concern.
type Kind int

const (
	// Refresh — a refresh-token session: the client holds the secret and presents
	// it to rotate/refresh. The only Kind used in the MVP.
	Refresh Kind = iota
	// ServerSide — reserved. A future stateful/opaque server-side session mode. It
	// is defined so the discriminator and its persistence mapping stay stable, but
	// the domain never instantiates it in the MVP, and Rotate rejects it.
	ServerSide
)

// String renders the kind for logging and persistence mapping. It changes no
// behavior — it only names the enum value.
func (kind Kind) String() string {
	switch kind {
	case Refresh:
		return "refresh"
	case ServerSide:
		return "server_side"
	default:
		return "unknown"
	}
}
