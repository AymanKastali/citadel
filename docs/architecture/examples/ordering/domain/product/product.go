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

// ReconstituteParams groups the full persisted state of a product. A product has no
// lifecycle beyond its name and price, so this mirrors NewProductParams.
type ReconstituteParams struct {
	ID    ID
	Name  Name
	Price Price
}

// Reconstitute rebuilds a product from stored state (repository adapter only). It just
// loads the persisted fields into a fresh entity — no validation, no event, no policy.
func Reconstitute(params ReconstituteParams) *Product {
	return &Product{
		Entity: domain.NewEntity(params.ID),
		name:   params.Name,
		price:  params.Price,
	}
}

func (product *Product) Name() Name { return product.name }

func (product *Product) Price() Price { return product.price }

// Reprice changes the price and records the fact. It takes a Price value object,
// so there is no raw input left to validate here.
func (product *Product) Reprice(price Price) {
	product.price = price
	product.Record(NewProductRepricedEvent(product.ID()))
}

func (product *Product) Rename(name Name) { product.name = name }
