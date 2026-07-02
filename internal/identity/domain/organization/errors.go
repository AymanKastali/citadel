package organization

import (
	"fmt"

	"github.com/AymanKastali/citadel/internal/identity/domain"
)

// Value-object failures — raised by the value object constructors and the Create
// guard clauses.

func NewEmptyIDError() *domain.DomainError {
	return &domain.DomainError{Message: "organization id must not be empty"}
}

func NewIDTooLongError(length int) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("organization id must not exceed %d characters, got %d", maxIDLength, length),
	}
}

func NewEmptyNameError() *domain.DomainError {
	return &domain.DomainError{Message: "organization name must not be empty"}
}

func NewNameTooLongError(length int) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("organization name must not exceed %d characters, got %d", maxNameLength, length),
	}
}

func NewEmptySlugError() *domain.DomainError {
	return &domain.DomainError{Message: "organization slug must not be empty"}
}

func NewSlugTooLongError(length int) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("organization slug must not exceed %d characters, got %d", maxSlugLength, length),
	}
}

func NewMalformedSlugError(value string) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("organization slug is malformed: %q (want lowercase alphanumeric labels joined by single hyphens)", value),
	}
}

// Transition failures — raised by the commands' guard clauses.

func NewAlreadySuspendedError() *domain.DomainError {
	return &domain.DomainError{Message: "organization is already suspended"}
}

func NewAlreadyActiveError() *domain.DomainError {
	return &domain.DomainError{Message: "organization is already active"}
}
