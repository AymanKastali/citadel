package product

const productRepricedEventName = "product.repriced"

// ProductRepriced is recorded when a product's price changes. It is an immutable
// domain event and carries the product's id, not the product itself.
type ProductRepriced struct {
	productID ID
}

func NewProductRepriced(productID ID) ProductRepriced { return ProductRepriced{productID: productID} }

func (event ProductRepriced) ProductID() ID { return event.productID }

func (event ProductRepriced) EventName() string { return productRepricedEventName }
