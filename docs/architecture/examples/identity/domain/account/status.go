package account

// Status is where an account sits in its lifecycle. Only the two
// registration-time statuses (PendingVerification, Active) are ever chosen by
// the email-verification policy; the rest are reached by later transitions the
// account owns (verifying, suspending, deactivating).
type Status int

const (
	// PendingVerification — the account exists but must confirm its email
	// before it is fully usable. The starting status when verification is required.
	PendingVerification Status = iota
	// Active — the account is usable. The starting status when verification is
	// not required, and where a PendingVerification account lands once confirmed.
	Active
	// Suspended — temporarily blocked by an operator; can be reactivated.
	Suspended
	// Deactivated — closed by the user or an operator; terminal.
	Deactivated
)
