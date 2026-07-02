package role

import "strings"

// maxIDLength is the upper bound of a valid id: comfortably fits any identifier
// scheme we use (UUIDs, ULIDs) while rejecting anything unreasonably long.
const maxIDLength = 64

// ID is a role's identity as a value object: immutable and self-validating.
type ID struct {
	value string
}

// NewID is the only way to build a valid ID.
func NewID(raw string) (ID, error) {
	trimmed := strings.TrimSpace(raw)
	idIsMissing := trimmed == ""
	if idIsMissing {
		return ID{}, NewEmptyIDError()
	}
	idIsTooLong := len(trimmed) > maxIDLength
	if idIsTooLong {
		return ID{}, NewIDTooLongError(len(trimmed))
	}
	return ID{value: trimmed}, nil
}

func (id ID) Value() string { return id.value }

func (id ID) Equal(other ID) bool { return id.value == other.value }

func (id ID) IsZero() bool { return id == ID{} }
