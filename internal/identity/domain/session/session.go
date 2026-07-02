package session

import (
	"time"

	"github.com/AymanKastali/citadel/internal/identity/domain"
	"github.com/AymanKastali/citadel/internal/identity/domain/account"
)

// Session is an entity: the persistent server-side record of one login on one
// device — the concept formerly called the "refresh token", renamed to name the
// thing itself (the token is merely the secret that proves possession of the
// session). It embeds the base domain.Entity for its id and recorded events, and
// references its owning account by typed account.ID only — never an embedded
// Account entity.
//
// A session is long-lived and mutated in place: Rotate swaps the stored secret
// hash on the same row for the whole duration of the login, so one Session is one
// login on one device — there is exactly one live row per device, not a chain of
// one-shot rows. State changes only through the methods below; callers read via
// getters, never by assigning fields.
type Session struct {
	domain.Entity[ID]
	accountID  account.ID
	kind       Kind
	secretHash SecretHash
	status     Status
	expiresAt  ExpiresAt
}

// OpenParams groups the already-valid value objects a session is opened from, so
// Open takes one data argument instead of a positional list.
type OpenParams struct {
	ID         ID
	AccountID  account.ID
	Kind       Kind
	SecretHash SecretHash
	ExpiresAt  ExpiresAt
}

// Open builds a new session, rejecting any missing (zero-value) field in field
// order (Fail Fast). Kind is not guarded — its zero value, Refresh, is the valid
// MVP kind. The session starts Active and records SessionOpenedEvent. No policy is
// involved — the caller supplies the Kind and the config-derived ExpiresAt.
func Open(params OpenParams) (*Session, error) {
	idIsMissing := params.ID.IsZero()
	if idIsMissing {
		return nil, NewEmptyIDError()
	}
	accountIsMissing := params.AccountID.IsZero()
	if accountIsMissing {
		return nil, NewEmptyAccountIDError()
	}
	secretHashIsMissing := params.SecretHash.IsZero()
	if secretHashIsMissing {
		return nil, NewEmptySecretHashError()
	}
	expiresAtIsMissing := params.ExpiresAt.IsZero()
	if expiresAtIsMissing {
		return nil, NewZeroExpiresAtError()
	}
	session := &Session{
		Entity:     domain.NewEntity(params.ID),
		accountID:  params.AccountID,
		kind:       params.Kind,
		secretHash: params.SecretHash,
		status:     Active, // sessions always start usable
		expiresAt:  params.ExpiresAt,
	}
	session.Record(NewSessionOpenedEvent(session.ID(), session.accountID))
	return session, nil
}

// ReconstituteParams groups the full persisted state of a session — a superset of
// OpenParams that also carries Status.
type ReconstituteParams struct {
	ID         ID
	AccountID  account.ID
	Kind       Kind
	SecretHash SecretHash
	Status     Status
	ExpiresAt  ExpiresAt
}

// Reconstitute rebuilds a session from stored state (repository adapter only). It
// just loads the persisted fields into a fresh entity — no validation, no event,
// no policy.
func Reconstitute(params ReconstituteParams) *Session {
	return &Session{
		Entity:     domain.NewEntity(params.ID),
		accountID:  params.AccountID,
		kind:       params.Kind,
		secretHash: params.SecretHash,
		status:     params.Status,
		expiresAt:  params.ExpiresAt,
	}
}

// Rotate swaps the secret hash in place — the heart of the mutate-in-place design.
// On each refresh the application hashes a newly issued secret and hands the hash
// here, which replaces the stored one on the same session; the old secret's hash
// is gone, so presenting the old secret afterwards no longer matches (the basis
// for reuse detection). Guards run before any mutation (Fail Fast): only a Refresh
// session may rotate, and it must be usable — Active and not past expiry. Status
// and expiry are left unchanged (lifetime extension, if any, is an application
// decision not modeled here). now is passed in (the domain reads no clock), so the
// expiry guard is checked live, independent of any lazily-persisted status marker.
func (session *Session) Rotate(newSecretHash SecretHash, now time.Time) error {
	notRefresh := session.kind != Refresh
	if notRefresh {
		return NewCannotRotateNonRefreshError()
	}
	notActive := session.status != Active
	if notActive {
		return NewSessionNotActiveError()
	}
	expired := session.IsExpired(now)
	if expired {
		return NewExpiredError()
	}
	session.secretHash = newSecretHash
	session.Record(NewSessionRotatedEvent(session.ID(), session.accountID))
	return nil
}

// Revoke terminates this one session — used for logout on a device and for reuse
// detection (revoking the compromised device's session). Idempotency is a guard,
// not a silent no-op: re-revoking is a business error. Setting Revoked is terminal.
// "Log out everywhere" is not modeled here; it is the application driving the
// repository's RevokeAllByAccount, which loads each session and calls Revoke().
func (session *Session) Revoke() error {
	alreadyRevoked := session.status == Revoked
	if alreadyRevoked {
		return NewAlreadyRevokedError()
	}
	session.status = Revoked
	session.Record(NewSessionRevokedEvent(session.ID(), session.accountID))
	return nil
}

func (session *Session) AccountID() account.ID { return session.accountID }

func (session *Session) Kind() Kind { return session.kind }

func (session *Session) Status() Status { return session.status }

// SecretHash exposes only the stored hash — the raw secret is never in the domain,
// so nothing can leak it. It exists for the application to compare a
// hashed-presented secret against the stored hash during rotation/verification.
func (session *Session) SecretHash() SecretHash { return session.secretHash }

// IsExpired reports whether now is at or after the expiry instant — the live,
// clock-based source of truth for liveness, regardless of the stored Status.
func (session *Session) IsExpired(now time.Time) bool {
	return !now.Before(session.expiresAt.Value())
}

// IsActive reports whether the session is usable right now: Active status AND not
// yet expired. It is the single question the application asks before honoring a
// session, so a not-yet-swept Active row whose ExpiresAt has passed is correctly
// reported as not active.
func (session *Session) IsActive(now time.Time) bool {
	return session.status == Active && !session.IsExpired(now)
}
