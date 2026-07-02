package account

import (
	"fmt"

	"github.com/AymanKastali/citadel/internal/identity/domain"
)

// Value-object failures (piece a) — raised by the value object constructors.

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

func NewEmptyPasswordHashError() *domain.DomainError {
	return &domain.DomainError{Message: "account password hash must not be empty"}
}

func NewPasswordHashTooLongError(length int) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("account password hash must not exceed %d characters, got %d", maxPasswordHashLength, length),
	}
}

// Transition failures (piece c) — raised by the commands' guard clauses.

func NewAlreadyVerifiedError() *domain.DomainError {
	return &domain.DomainError{Message: "account email is already verified"}
}

func NewCannotVerifyError() *domain.DomainError {
	return &domain.DomainError{Message: "account cannot be verified from its current status"}
}

func NewCannotChangePasswordError() *domain.DomainError {
	return &domain.DomainError{Message: "account password cannot be changed on a deactivated account"}
}

func NewAlreadySuspendedError() *domain.DomainError {
	return &domain.DomainError{Message: "account is already suspended"}
}

func NewCannotSuspendError() *domain.DomainError {
	return &domain.DomainError{Message: "only an active account can be suspended"}
}

func NewNotSuspendedError() *domain.DomainError {
	return &domain.DomainError{Message: "account is not suspended, so it cannot be reactivated"}
}

func NewAlreadyDeactivatedError() *domain.DomainError {
	return &domain.DomainError{Message: "account is already deactivated"}
}

func NewCannotLoginError() *domain.DomainError {
	return &domain.DomainError{Message: "account cannot log in from its current status"}
}
