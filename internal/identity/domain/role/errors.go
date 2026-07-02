package role

import (
	"fmt"

	"github.com/AymanKastali/citadel/internal/identity/domain"
)

// Value-object / construction failures. There is deliberately no "empty
// description" error — an absent description is a valid value, so the only
// description invariant is its max length.

func NewEmptyIDError() *domain.DomainError {
	return &domain.DomainError{Message: "role id must not be empty"}
}

func NewIDTooLongError(length int) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("role id must not exceed %d characters, got %d", maxIDLength, length),
	}
}

func NewEmptyNameError() *domain.DomainError {
	return &domain.DomainError{Message: "role name must not be empty"}
}

func NewNameTooLongError(length int) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("role name must be at most %d characters, got %d", maxNameLength, length),
	}
}

func NewDescriptionTooLongError(length int) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("role description must be at most %d characters, got %d", maxDescriptionLength, length),
	}
}

func NewEmptyOrganizationIDError() *domain.DomainError {
	return &domain.DomainError{Message: "role organization id must not be empty"}
}
