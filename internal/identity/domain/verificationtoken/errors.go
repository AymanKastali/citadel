package verificationtoken

import (
	"fmt"

	"github.com/AymanKastali/citadel/internal/identity/domain"
)

// Business-rule / invariant violations only, each a factory returning a
// *domain.DomainError. Lookup/persistence outcomes (not found, already exists) are
// not here; they belong at the repository/application boundary. The package
// carries the concept, so there is no stutter in the names.

// --- value-object / construction invariants ---

func NewEmptyIDError() *domain.DomainError {
	return &domain.DomainError{Message: "verification token id must not be empty"}
}

func NewIDTooLongError(length int) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("verification token id must be at most %d characters, got %d", maxIDLength, length),
	}
}

func NewEmptyAccountIDError() *domain.DomainError {
	return &domain.DomainError{Message: "verification token account id must not be empty"}
}

func NewEmptyPurposeError() *domain.DomainError {
	return &domain.DomainError{Message: "verification token purpose must be set"}
}

func NewEmptyTokenHashError() *domain.DomainError {
	return &domain.DomainError{Message: "verification token hash must not be empty"}
}

func NewTokenHashTooLongError(length int) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("verification token hash must be at most %d characters, got %d", maxTokenHashLength, length),
	}
}

func NewZeroExpiresAtError() *domain.DomainError {
	return &domain.DomainError{Message: "verification token expiry must not be the zero time"}
}

// --- command / lifecycle invariants ---

func NewAlreadyConsumedError() *domain.DomainError {
	return &domain.DomainError{Message: "verification token has already been consumed"}
}

func NewExpiredError() *domain.DomainError {
	return &domain.DomainError{Message: "verification token has expired"}
}
