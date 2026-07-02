package verificationtoken

// Purpose is which flow a token belongs to. Like account.Status it is an
// enumerated type set by the entity at issue time, not parsed from raw external
// input — the caller passes an already-chosen Purpose into Issue, so it has no
// New… constructor. Persistence <-> enum mapping is an infrastructure concern.
//
// The zero value is invalid: purposeUnset (0) is reserved as an unset sentinel so
// a zero-value field is caught by Issue's guard, and the real values start from 1.
type Purpose int

const (
	purposeUnset      Purpose = iota // 0 — invalid sentinel; a zero-value Purpose means "unset"
	EmailVerification                // confirms the account owns its email address
	PasswordReset                    // authorizes an out-of-band password reset
)

// IsZero reports whether the purpose is the unset sentinel — a caller that forgot
// to choose one.
func (purpose Purpose) IsZero() bool { return purpose == purposeUnset }

// String renders the purpose for logging and persistence mapping. It changes no
// behavior — it only names the enum value.
func (purpose Purpose) String() string {
	switch purpose {
	case EmailVerification:
		return "email_verification"
	case PasswordReset:
		return "password_reset"
	default:
		return "unknown"
	}
}
