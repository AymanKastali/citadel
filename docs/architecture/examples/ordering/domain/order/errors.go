package order

import (
	"fmt"

	"github.com/AymanKastali/citadel/internal/ordering/domain"
)

func NewEmptyIDError() *domain.DomainError {
	return &domain.DomainError{Message: "order id must not be empty"}
}

func NewIDTooLongError(length int) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("order id must not exceed %d characters, got %d", maxIDLength, length),
	}
}

func NewInvalidQuantityError(value int) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("order quantity must be positive, got %d", value),
	}
}

func NewQuantityTooLargeError(value int) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("order quantity must not exceed %d, got %d", maxQuantity, value),
	}
}

func NewLineWithoutProductError() *domain.DomainError {
	return &domain.DomainError{Message: "order line must refer to a product"}
}

func NewLineProductIDTooLongError(length int) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("order line product id must not exceed %d characters, got %d", maxIDLength, length),
	}
}

func NewLineWithoutQuantityError() *domain.DomainError {
	return &domain.DomainError{Message: "order line must have a quantity"}
}

func NewInvalidLinePriceError(amountInCents int) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("order line price must be positive, got %d", amountInCents),
	}
}

func NewLinePriceTooLargeError(amountInCents int) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("order line price must not exceed %d, got %d", maxUnitPriceInCents, amountInCents),
	}
}

func NewAlreadyShippedError() *domain.DomainError {
	return &domain.DomainError{Message: "order has already shipped and can no longer change"}
}

func NewEmptyOrderError() *domain.DomainError {
	return &domain.DomainError{Message: "order must have at least one line before it can ship"}
}
