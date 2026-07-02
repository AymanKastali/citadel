package session

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
	return &domain.DomainError{Message: "session id must not be empty"}
}

func NewIDTooLongError(length int) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("session id must be at most %d characters, got %d", maxIDLength, length),
	}
}

func NewEmptyAccountIDError() *domain.DomainError {
	return &domain.DomainError{Message: "session account id must not be empty"}
}

func NewEmptySecretHashError() *domain.DomainError {
	return &domain.DomainError{Message: "session secret hash must not be empty"}
}

func NewSecretHashTooLongError(length int) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("session secret hash must be at most %d characters, got %d", maxSecretHashLength, length),
	}
}

func NewZeroExpiresAtError() *domain.DomainError {
	return &domain.DomainError{Message: "session expiry must not be the zero time"}
}

// --- command / lifecycle invariants ---

func NewCannotRotateNonRefreshError() *domain.DomainError {
	return &domain.DomainError{Message: "only a refresh session may be rotated"}
}

func NewSessionNotActiveError() *domain.DomainError {
	return &domain.DomainError{Message: "session is not active"}
}

func NewExpiredError() *domain.DomainError {
	return &domain.DomainError{Message: "session has expired"}
}

func NewAlreadyRevokedError() *domain.DomainError {
	return &domain.DomainError{Message: "session is already revoked"}
}
