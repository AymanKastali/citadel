package account

import (
	"time"

	"github.com/AymanKastali/citadel/internal/identity/domain"
)

// Account is an entity: mutable, and the single owner of the rules that govern
// an account. State changes only through its methods, never by direct field
// assignment from outside the package. It embeds domain.Entity for its id and
// recorded events. It references no other entity — an account's organization and
// roles live on membership, not here.
type Account struct {
	domain.Entity[ID]
	email        Email
	passwordHash PasswordHash
	status       Status
	lastLoginAt  *time.Time
}

// RegisterParams groups the already-valid value objects an account is
// registered from, so Register takes one data argument instead of a positional
// list. The email-verification policy is NOT a field here — a policy is a
// behavioral dependency, not data, so it is passed alongside this struct.
type RegisterParams struct {
	ID           ID
	Email        Email
	PasswordHash PasswordHash
}

// Register builds a new account, rejecting any missing (zero-value) field. The
// email-verification policy is passed alongside the params struct — never
// inside it — so the caller (wired at the composition root from config) chooses
// the strategy without the entity or the params knowing which one. The policy
// decides the starting status (a query); the account records what happened.
// lastLoginAt is left nil: a freshly registered account has never logged in.
func Register(params RegisterParams, verification EmailVerificationPolicy) (*Account, error) {
	idIsMissing := params.ID.IsZero()
	if idIsMissing {
		return nil, NewEmptyIDError()
	}
	emailIsMissing := params.Email.IsZero()
	if emailIsMissing {
		return nil, NewEmptyEmailError()
	}
	passwordHashIsMissing := params.PasswordHash.IsZero()
	if passwordHashIsMissing {
		return nil, NewEmptyPasswordHashError()
	}
	account := &Account{
		Entity:       domain.NewEntity(params.ID),
		email:        params.Email,
		passwordHash: params.PasswordHash,
		status:       verification.InitialStatus(), // the strategy decides; the entity accommodates
	}
	account.Record(NewAccountRegisteredEvent(account.ID(), account.status))
	return account, nil
}

// ReconstituteParams groups the full persisted state of an account, as read back
// from storage. Unlike RegisterParams it also carries the status and last-login
// time, because rebuilding an account restores where it already sits in its
// lifecycle instead of deriving a starting status from a policy.
type ReconstituteParams struct {
	ID           ID
	Email        Email
	PasswordHash PasswordHash
	Status       Status
	LastLoginAt  *time.Time
}

// Reconstitute rebuilds an account from stored state (repository adapter only). It
// just loads the persisted fields into a fresh entity — no validation, no event, no
// policy — so it takes only a params struct and returns the entity. LastLoginAt is
// carried through as-is (nil for an account that has never logged in).
func Reconstitute(params ReconstituteParams) *Account {
	return &Account{
		Entity:       domain.NewEntity(params.ID),
		email:        params.Email,
		passwordHash: params.PasswordHash,
		status:       params.Status,
		lastLoginAt:  params.LastLoginAt,
	}
}

// VerifyEmail confirms the account's email, moving PendingVerification → Active.
// An already-Active account is a distinct, self-describing failure from one that
// can never be verified (suspended or terminal), so the two guards raise
// different errors.
func (account *Account) VerifyEmail() error {
	alreadyVerified := account.status == Active
	if alreadyVerified {
		return NewAlreadyVerifiedError()
	}
	cannotVerify := account.status != PendingVerification
	if cannotVerify {
		return NewCannotVerifyError()
	}
	account.status = Active
	account.Record(NewEmailVerifiedEvent(account.ID()))
	return nil
}

// ChangePassword replaces the stored hash. It takes an already-hashed
// PasswordHash value object (the plaintext never enters the domain). It is
// rejected on a Deactivated account — a terminal account is closed, so mutating
// its credentials is meaningless; every other status may change its password.
// Status is unchanged; only the hash and the recorded fact move.
func (account *Account) ChangePassword(newHash PasswordHash) error {
	cannotChangePassword := account.status == Deactivated
	if cannotChangePassword {
		return NewCannotChangePasswordError()
	}
	account.passwordHash = newHash
	account.Record(NewPasswordChangedEvent(account.ID()))
	return nil
}

// Suspend blocks an account by operator action, moving Active → Suspended.
// Suspending an already-Suspended account and suspending one that cannot be
// suspended at all (PendingVerification or the terminal Deactivated) are
// separate failures.
func (account *Account) Suspend() error {
	alreadySuspended := account.status == Suspended
	if alreadySuspended {
		return NewAlreadySuspendedError()
	}
	cannotSuspend := account.status != Active
	if cannotSuspend {
		return NewCannotSuspendError()
	}
	account.status = Suspended
	account.Record(NewAccountSuspendedEvent(account.ID()))
	return nil
}

// Reactivate lifts a suspension, moving Suspended → Active. It is the inverse of
// Suspend and the only way back to Active from Suspended; any other starting
// status is a single failure.
func (account *Account) Reactivate() error {
	notSuspended := account.status != Suspended
	if notSuspended {
		return NewNotSuspendedError()
	}
	account.status = Active
	account.Record(NewAccountReactivatedEvent(account.ID()))
	return nil
}

// Deactivate closes the account, moving any non-terminal status → Deactivated
// (terminal). The only guard is against re-deactivating an already-Deactivated
// account, which keeps the transition idempotent-by-rejection rather than
// silently re-recording the fact.
func (account *Account) Deactivate() error {
	alreadyDeactivated := account.status == Deactivated
	if alreadyDeactivated {
		return NewAlreadyDeactivatedError()
	}
	account.status = Deactivated
	account.Record(NewAccountDeactivatedEvent(account.ID()))
	return nil
}

// Login records a successful authentication (status unchanged). It does not check
// the password — credential verification is the application's PasswordHasher port.
// By the time this is called, the app's Authenticate use case has already verified
// the password; Login enforces the one rule the account itself owns — only an
// Active account may log in — then stamps lastLoginAt and records the fact. The
// three non-Active statuses collapse to a single NewCannotLoginError; the
// application may branch on Status() for messaging. The domain never reads the
// clock itself; the application passes the current time in.
func (account *Account) Login(now time.Time) error {
	cannotLogin := account.status != Active
	if cannotLogin {
		return NewCannotLoginError()
	}
	account.lastLoginAt = &now
	account.Record(NewAccountLoggedInEvent(account.ID(), now))
	return nil
}

func (account *Account) Email() Email { return account.email }

func (account *Account) PasswordHash() PasswordHash { return account.passwordHash }

func (account *Account) Status() Status { return account.status }

// LastLoginAt reports the last successful authentication. The bool is false when the
// account has never logged in (lastLoginAt is nil); returning a value + ok keeps the
// caller from dereferencing a pointer into the entity's internal state.
func (account *Account) LastLoginAt() (time.Time, bool) {
	neverLoggedIn := account.lastLoginAt == nil
	if neverLoggedIn {
		return time.Time{}, false
	}
	return *account.lastLoginAt, true
}
