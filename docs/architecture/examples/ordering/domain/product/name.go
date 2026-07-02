package product

import (
	"strings"
	"unicode/utf8"
)

// maxNameLength is the upper bound of a valid name: long enough for any real
// product name, short enough to reject junk and abuse.
const maxNameLength = 200

// Name is a product's name as a value object: immutable and self-validating.
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
