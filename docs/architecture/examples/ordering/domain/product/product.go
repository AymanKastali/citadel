package product

import "github.com/AymanKastali/citadel/internal/ordering/domain"

// Product is an entity: mutable, and the single owner of the rules that govern
// a product. State changes only through its methods, never by direct field
// assignment from outside the package. It embeds domain.Entity for its id and
// recorded events.
type Product struct {
	domain.Entity[ID]
	name  Name
	price Price
}

// NewProductParams groups the already-valid value objects a Product is built
// from, so the constructor takes one argument instead of a positional list.
type NewProductParams struct {
	ID    ID
	Name  Name
	Price Price
}

// NewProduct builds a Product, rejecting any missing (zero-value) field.
func NewProduct(params NewProductParams) (*Product, error) {
	idIsMissing := params.ID.IsZero()
	if idIsMissing {
		return nil, NewEmptyIDError()
	}
	nameIsMissing := params.Name.IsZero()
	if nameIsMissing {
		return nil, NewEmptyNameError()
	}
	priceIsMissing := params.Price.IsZero()
	if priceIsMissing {
		return nil, NewInvalidPriceError(0)
	}
	return &Product{
		Entity: domain.NewEntity(params.ID),
		name:   params.Name,
		price:  params.Price,
	}, nil
}

func (product *Product) Name() Name { return product.name }

func (product *Product) Price() Price { return product.price }

// Reprice changes the price and records the fact. It takes a Price value object,
// so there is no raw input left to validate here.
func (product *Product) Reprice(price Price) {
	product.price = price
	product.Record(NewProductRepriced(product.ID()))
}

func (product *Product) Rename(name Name) { product.name = name }
