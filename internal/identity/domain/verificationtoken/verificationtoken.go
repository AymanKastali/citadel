package verificationtoken

import (
	"time"

	"github.com/AymanKastali/citadel/internal/identity/domain"
	"github.com/AymanKastali/citadel/internal/identity/domain/account"
)

// VerificationToken is an entity: a one-shot, short-lived token issued against an
// account to prove control of a channel. A single generic entity covers both flows
// (email verification and password reset), told apart by its Purpose. It embeds the
// base domain.Entity for its id and recorded events, and references its owning
// account by typed account.ID only — never an embedded Account entity.
//
// consumedAt records the single-use marker: nil means unconsumed, a non-nil pointer
// records when it was consumed. A *time.Time is chosen over a bare bool so the
// timestamp also captures when (useful for auditing) and maps cleanly to a nullable
// column. State changes only through the methods below; callers read via getters,
// never by assigning fields.
type VerificationToken struct {
	domain.Entity[ID]
	accountID  account.ID
	purpose    Purpose
	hash       TokenHash
	expiresAt  ExpiresAt
	consumedAt *time.Time // nil = unconsumed; non-nil records when it was consumed
}

// IssueParams groups the already-valid value objects a token is issued from, so
// Issue takes one data argument instead of a positional list.
type IssueParams struct {
	ID        ID
	AccountID account.ID
	Purpose   Purpose
	Hash      TokenHash
	ExpiresAt ExpiresAt
}

// Issue builds a new token, rejecting any missing (zero-value) field in field
// order (Fail Fast). The value objects are already valid (their own constructors
// reject empty/too-long/malformed), so these guards defend the entity's own
// invariant that none of its fields is the zero value. The token starts unconsumed
// and records TokenIssuedEvent. There is no policy — Purpose, ExpiresAt, and the
// hash are decided by the caller (from config-driven TTL and the flow it is
// running) and passed in as ready value objects.
func Issue(params IssueParams) (*VerificationToken, error) {
	idIsMissing := params.ID.IsZero()
	if idIsMissing {
		return nil, NewEmptyIDError()
	}
	accountIsMissing := params.AccountID.IsZero()
	if accountIsMissing {
		return nil, NewEmptyAccountIDError()
	}
	purposeIsMissing := params.Purpose.IsZero()
	if purposeIsMissing {
		return nil, NewEmptyPurposeError()
	}
	hashIsMissing := params.Hash.IsZero()
	if hashIsMissing {
		return nil, NewEmptyTokenHashError()
	}
	expiresAtIsMissing := params.ExpiresAt.IsZero()
	if expiresAtIsMissing {
		return nil, NewZeroExpiresAtError()
	}
	token := &VerificationToken{
		Entity:     domain.NewEntity(params.ID),
		accountID:  params.AccountID,
		purpose:    params.Purpose,
		hash:       params.Hash,
		expiresAt:  params.ExpiresAt,
		consumedAt: nil, // starts unconsumed
	}
	token.Record(NewTokenIssuedEvent(token.ID(), token.accountID, token.purpose))
	return token, nil
}

// ReconstituteParams groups the full persisted state of a token — a superset of
// IssueParams that also carries the ConsumedAt marker.
type ReconstituteParams struct {
	ID         ID
	AccountID  account.ID
	Purpose    Purpose
	Hash       TokenHash
	ExpiresAt  ExpiresAt
	ConsumedAt *time.Time
}

// Reconstitute rebuilds a token from stored state (repository adapter only). It
// just loads the persisted fields into a fresh entity — no validation, no event,
// no policy. The stored consumed_at is nullable, so ConsumedAt is loaded through
// as-is: a spent token rebuilds as spent.
func Reconstitute(params ReconstituteParams) *VerificationToken {
	return &VerificationToken{
		Entity:     domain.NewEntity(params.ID),
		accountID:  params.AccountID,
		purpose:    params.Purpose,
		hash:       params.Hash,
		expiresAt:  params.ExpiresAt,
		consumedAt: params.ConsumedAt,
	}
}

// Consume spends the token exactly once. Guards run before any mutation (Fail
// Fast): an already-consumed token cannot be spent twice, and an expired token is
// unusable even if never consumed. Only if both pass does it record the
// consumption instant and TokenConsumedEvent. The secret comparison that
// authorizes reaching this call is an infrastructure/application concern — by the
// time Consume runs, the matching token has already been located by its hash.
func (token *VerificationToken) Consume(now time.Time) error {
	alreadyConsumed := token.consumedAt != nil
	if alreadyConsumed {
		return NewAlreadyConsumedError()
	}
	expired := token.IsExpired(now)
	if expired {
		return NewExpiredError()
	}
	consumedAt := now
	token.consumedAt = &consumedAt
	token.Record(NewTokenConsumedEvent(token.ID(), token.accountID, token.purpose))
	return nil
}

func (token *VerificationToken) AccountID() account.ID { return token.accountID }

func (token *VerificationToken) Purpose() Purpose { return token.purpose }

// IsExpired reports whether now is at or after the expiry instant.
func (token *VerificationToken) IsExpired(now time.Time) bool {
	return !now.Before(token.expiresAt.Value())
}

// IsConsumed reports whether the token has already been spent.
func (token *VerificationToken) IsConsumed() bool { return token.consumedAt != nil }
