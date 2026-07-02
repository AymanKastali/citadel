package membership

import (
	"fmt"

	"github.com/AymanKastali/citadel/internal/identity/domain"
)

// Value-object / construction failures — raised by the value object constructors
// and the Create guard clauses.

func NewEmptyIDError() *domain.DomainError {
	return &domain.DomainError{Message: "membership id must not be empty"}
}

func NewIDTooLongError(length int) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("membership id must not exceed %d characters, got %d", maxIDLength, length),
	}
}

func NewEmptyAccountIDError() *domain.DomainError {
	return &domain.DomainError{Message: "membership account id must not be empty"}
}

func NewEmptyOrganizationIDError() *domain.DomainError {
	return &domain.DomainError{Message: "membership organization id must not be empty"}
}

// Role-set guards — raised by AssignRole / RevokeRole.

func NewRoleAlreadyAssignedError() *domain.DomainError {
	return &domain.DomainError{Message: "role is already assigned to this membership"}
}

func NewRoleNotAssignedError() *domain.DomainError {
	return &domain.DomainError{Message: "role is not assigned to this membership"}
}

// Transition failures — raised by the commands' guard clauses.

func NewAlreadySuspendedError() *domain.DomainError {
	return &domain.DomainError{Message: "membership is already suspended"}
}

func NewAlreadyActiveError() *domain.DomainError {
	return &domain.DomainError{Message: "membership is already active"}
}
