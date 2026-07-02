package product

// maxPriceInCents is the upper bound of a valid price: a ceiling that catches
// overflow and fat-finger input long before it reaches persistence.
const maxPriceInCents = 100_000_000 // 1,000,000.00 in a two-decimal currency

// Price is a product's price in minor currency units (for example, cents).
// It is a value object: immutable, and always within its valid range.
type Price struct {
	amountInCents int
}

func NewPrice(amountInCents int) (Price, error) {
	amountIsZeroOrNegative := amountInCents <= 0
	if amountIsZeroOrNegative {
		return Price{}, NewInvalidPriceError(amountInCents)
	}
	amountExceedsMaximum := amountInCents > maxPriceInCents
	if amountExceedsMaximum {
		return Price{}, NewPriceTooLargeError(amountInCents)
	}
	return Price{amountInCents: amountInCents}, nil
}

func (price Price) AmountInCents() int { return price.amountInCents }

func (price Price) Equal(other Price) bool { return price.amountInCents == other.amountInCents }

func (price Price) IsZero() bool { return price == Price{} }
