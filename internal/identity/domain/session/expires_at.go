package session

import "time"

// ExpiresAt is the moment a session stops being valid, as a value object:
// immutable and self-validating. It wraps a stdlib time.Time (stdlib time is
// allowed in the domain; third-party packages are not). The domain does no I/O and
// reads no clock — "now" is passed into the methods that need it, never fetched
// here.
type ExpiresAt struct {
	value time.Time
}

// NewExpiresAt is the only way to build a valid ExpiresAt. It rejects the zero
// time; it applies no other bound — the lifetime is chosen by the application from
// config, and whether the instant is in the past is a runtime question answered by
// the entity's IsExpired(now), not a construction invariant.
func NewExpiresAt(raw time.Time) (ExpiresAt, error) {
	expiresAtIsMissing := raw.IsZero()
	if expiresAtIsMissing {
		return ExpiresAt{}, NewZeroExpiresAtError()
	}
	return ExpiresAt{value: raw}, nil
}

func (expiresAt ExpiresAt) Value() time.Time { return expiresAt.value }

// Equal compares via time.Time.Equal (never ==, which also compares the monotonic
// clock reading and location).
func (expiresAt ExpiresAt) Equal(other ExpiresAt) bool {
	return expiresAt.value.Equal(other.value)
}

func (expiresAt ExpiresAt) IsZero() bool { return expiresAt.value.IsZero() }
