package role

import (
	"strings"
	"unicode/utf8"
)

// maxNameLength bounds the role's human-readable name, counted in runes so
// multi-byte names are measured fairly.
const maxNameLength = 100

// Name is a role's human-readable name as a value object (e.g. "Administrator",
// "Billing Manager"): immutable and self-validating.
type Name struct {
	value string
}

func NewName(raw string) (Name, error) {
	trimmed := strings.TrimSpace(raw)
	nameIsMissing := trimmed == ""
	if nameIsMissing {
		return Name{}, NewEmptyNameError()
	}
	nameIsTooLong := utf8.RuneCountInString(trimmed) > maxNameLength
	if nameIsTooLong {
		return Name{}, NewNameTooLongError(utf8.RuneCountInString(trimmed))
	}
	return Name{value: trimmed}, nil
}

func (name Name) Value() string { return name.value }

func (name Name) Equal(other Name) bool { return name.value == other.value }

func (name Name) IsZero() bool { return name == Name{} }
