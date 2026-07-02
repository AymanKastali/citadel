package product

const productRepricedEventName = "product.repriced"

// ProductRepricedEvent is recorded when a product's price changes. It is an immutable
// domain event and carries the product's id, not the product itself.
type ProductRepricedEvent struct {
	productID ID
}

func NewProductRepricedEvent(productID ID) ProductRepricedEvent {
	return ProductRepricedEvent{productID: productID}
}

func (event ProductRepricedEvent) ProductID() ID { return event.productID }

func (event ProductRepricedEvent) EventName() string { return productRepricedEventName }
