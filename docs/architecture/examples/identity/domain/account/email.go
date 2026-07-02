package account

import "strings"

// maxEmailLength is the upper bound of a valid address, matching the practical
// limit set by RFC 5321 — long enough for any real address, short enough to
// reject junk.
const maxEmailLength = 254

// Email is an account's email address as a value object: immutable and
// self-validating. It is normalized to lower case so equality is stable.
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
	emailIsMalformed := !strings.ContainsRune(trimmed, '@')
	if emailIsMalformed {
		return Email{}, NewMalformedEmailError(trimmed)
	}
	return Email{value: strings.ToLower(trimmed)}, nil
}

func (email Email) Value() string { return email.value }

func (email Email) Equal(other Email) bool { return email.value == other.value }

func (email Email) IsZero() bool { return email == Email{} }
