package product

import (
	"fmt"

	"github.com/AymanKastali/citadel/internal/ordering/domain"
)

func NewEmptyIDError() *domain.DomainError {
	return &domain.DomainError{Message: "product id must not be empty"}
}

func NewIDTooLongError(length int) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("product id must not exceed %d characters, got %d", maxIDLength, length),
	}
}

func NewEmptyNameError() *domain.DomainError {
	return &domain.DomainError{Message: "product name must not be empty"}
}

func NewNameTooLongError(length int) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("product name must not exceed %d characters, got %d", maxNameLength, length),
	}
}

func NewInvalidPriceError(amountInCents int) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("product price must be positive, got %d", amountInCents),
	}
}

func NewPriceTooLargeError(amountInCents int) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("product price must not exceed %d, got %d", maxPriceInCents, amountInCents),
	}
}
