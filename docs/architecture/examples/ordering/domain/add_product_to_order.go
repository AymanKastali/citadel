package domain

import (
	"github.com/AymanKastali/citadel/internal/ordering/domain/order"
	"github.com/AymanKastali/citadel/internal/ordering/domain/product"
)

// AddProductToOrder puts a product onto an order at the product's current
// price. It spans two entities and belongs to neither, so it lives at the
// domain root as a domain service: stateless, pure, and free of persistence.
//
// It coordinates the two entities and translates between them; each entity
// still enforces its own rules (the order rejects the line if it has shipped).
func AddProductToOrder(targetOrder *order.Order, orderedProduct *product.Product, quantity order.Quantity) error {
	line, err := order.NewLine(order.NewLineParams{
		ProductID:        orderedProduct.ID().Value(),
		Quantity:         quantity,
		UnitPriceInCents: orderedProduct.Price().AmountInCents(),
	})
	if err != nil {
		return err
	}
	return targetOrder.AddLine(line)
}
