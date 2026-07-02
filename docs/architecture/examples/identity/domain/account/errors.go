package account

import (
	"fmt"

	"github.com/AymanKastali/citadel/internal/identity/domain"
)

func NewEmptyIDError() *domain.DomainError {
	return &domain.DomainError{Message: "account id must not be empty"}
}

func NewIDTooLongError(length int) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("account id must not exceed %d characters, got %d", maxIDLength, length),
	}
}

func NewEmptyEmailError() *domain.DomainError {
	return &domain.DomainError{Message: "account email must not be empty"}
}

func NewEmailTooLongError(length int) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("account email must not exceed %d characters, got %d", maxEmailLength, length),
	}
}

func NewMalformedEmailError(value string) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("account email is malformed: %q", value),
	}
}
