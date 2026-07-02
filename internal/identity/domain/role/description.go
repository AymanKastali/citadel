package role

import (
	"strings"
	"unicode/utf8"
)

// maxDescriptionLength bounds the role's free-text description, counted in runes.
const maxDescriptionLength = 500

// Description is a role's free-text description as a value object: immutable and
// self-validating. Unlike every other value object in this context it is
// OPTIONAL — an empty (or whitespace-only) value is valid and yields the
// zero-value Description, not an error; the only invariant is the upper bound. So
// IsZero() here answers the business question "was a description provided?" — it
// is informational, not an invalidity signal.
type Description struct {
	value string
}

// NewDescription accepts an empty value — an absent description is valid. It
// enforces only the upper bound; empty trims to the zero-value Description.
func NewDescription(raw string) (Description, error) {
	trimmed := strings.TrimSpace(raw)
	descriptionIsTooLong := utf8.RuneCountInString(trimmed) > maxDescriptionLength
	if descriptionIsTooLong {
		return Description{}, NewDescriptionTooLongError(utf8.RuneCountInString(trimmed))
	}
	return Description{value: trimmed}, nil // trimmed may be "" — that is a valid Description
}

func (description Description) Value() string { return description.value }

func (description Description) Equal(other Description) bool {
	return description.value == other.value
}

func (description Description) IsZero() bool { return description == Description{} }
