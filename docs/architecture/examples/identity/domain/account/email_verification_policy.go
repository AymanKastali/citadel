package account

// EmailVerificationPolicy decides the status a newly registered account starts
// in. It is a domain policy — a Strategy: a pure interface, injected into
// Register alongside its params, with a concrete strategy chosen at the
// composition root from configuration.
//
// The interface lives in the account package because it serves a single entity
// (unlike a cross-entity domain service, which lives at the domain root). Its
// method is a query — it decides and returns; the entity acts on the answer and
// keeps ownership of its own state.
type EmailVerificationPolicy interface {
	// InitialStatus decides the status a freshly registered account starts in.
	InitialStatus() Status
}

// RequiredEmailVerification starts a new account pending until it confirms its
// email. Selected at the composition root when EMAIL_VERIFICATION_REQUIRED=true.
type RequiredEmailVerification struct{}

func (RequiredEmailVerification) InitialStatus() Status { return PendingVerification }

// OptionalEmailVerification activates a new account immediately. Selected at the
// composition root when EMAIL_VERIFICATION_REQUIRED=false.
type OptionalEmailVerification struct{}

func (OptionalEmailVerification) InitialStatus() Status { return Active }
