package account

import (
	"net/mail"
	"strings"
)

// maxEmailLength is the upper bound of a valid address, matching the practical
// limit set by RFC 5321 — long enough for any real address, short enough to
// reject junk.
const maxEmailLength = 254

// Email is an account's email address as a value object: immutable and
// self-validating. It is validated against the RFC 5322 address grammar (via the
// standard library's net/mail parser — no third-party dependency, so the domain
// stays pure) and normalized to lower case so equality is stable.
type Email struct {
	value string
}

func NewEmail(raw string) (Email, error) {
	trimmed := strings.TrimSpace(raw)
	emailIsMissing := trimmed == ""
	if emailIsMissing {
		return Email{}, NewEmptyEmailError()
	}
	emailIsTooLong := len(trimmed) > maxEmailLength
	if emailIsTooLong {
		return Email{}, NewEmailTooLongError(len(trimmed))
	}
	address, err := mail.ParseAddress(trimmed)
	if err != nil {
		return Email{}, NewMalformedEmailError(trimmed)
	}
	// ParseAddress also accepts display-name / angle-bracket / comment forms
	// (e.g. "Name <a@b.com>"). An account email must be a bare addr-spec, so
	// reject anything that carried a display name or did not round-trip to
	// exactly what was supplied.
	hasDisplayName := address.Name != ""
	if hasDisplayName {
		return Email{}, NewMalformedEmailError(trimmed)
	}
	notBareAddress := address.Address != trimmed
	if notBareAddress {
		return Email{}, NewMalformedEmailError(trimmed)
	}
	return Email{value: strings.ToLower(address.Address)}, nil
}

func (email Email) Value() string { return email.value }

func (email Email) Equal(other Email) bool { return email.value == other.value }

func (email Email) IsZero() bool { return email == Email{} }
