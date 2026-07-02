package permission

import (
	"fmt"

	"github.com/AymanKastali/citadel/internal/identity/domain"
)

// Value-object / construction failures — raised by the value object constructors
// and the Grant guard clauses. Lookup/persistence outcomes (not found, already
// exists) are not here; they belong at the repository/application boundary.

func NewEmptyIDError() *domain.DomainError {
	return &domain.DomainError{Message: "permission id must not be empty"}
}

func NewIDTooLongError(length int) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("permission id must not exceed %d characters, got %d", maxIDLength, length),
	}
}

func NewEmptyRoleIDError() *domain.DomainError {
	return &domain.DomainError{Message: "permission role id must not be empty"}
}

func NewEmptyScopeError() *domain.DomainError {
	return &domain.DomainError{Message: "permission scope must not be empty"}
}

func NewScopeTooLongError(length int) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("permission scope must be at most %d characters, got %d", maxScopeLength, length),
	}
}

func NewMalformedScopeError(value string) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("permission scope %q is malformed: expected colon-separated lowercase alphanumeric segments (e.g. \"users:read\")", value),
	}
}
