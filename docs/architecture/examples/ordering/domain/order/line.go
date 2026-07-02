package order

// maxUnitPriceInCents is the upper bound of a valid line price: a ceiling that
// catches overflow and fat-finger input long before it reaches persistence.
const maxUnitPriceInCents = 100_000_000 // 1,000,000.00 in a two-decimal currency

// Line is one product on an order, captured as a value object: the product it
// refers to, how many, and the unit price at the time it was added.
type Line struct {
	productID        string
	quantity         Quantity
	unitPriceInCents int
}

// NewLineParams groups the values a Line is built from.
type NewLineParams struct {
	ProductID        string
	Quantity         Quantity
	UnitPriceInCents int
}

func NewLine(params NewLineParams) (Line, error) {
	productIsMissing := params.ProductID == ""
	if productIsMissing {
		return Line{}, NewLineWithoutProductError()
	}
	productIDIsTooLong := len(params.ProductID) > maxIDLength
	if productIDIsTooLong {
		return Line{}, NewLineProductIDTooLongError(len(params.ProductID))
	}
	quantityIsMissing := params.Quantity.IsZero()
	if quantityIsMissing {
		return Line{}, NewLineWithoutQuantityError()
	}
	priceIsZeroOrNegative := params.UnitPriceInCents <= 0
	if priceIsZeroOrNegative {
		return Line{}, NewInvalidLinePriceError(params.UnitPriceInCents)
	}
	priceExceedsMaximum := params.UnitPriceInCents > maxUnitPriceInCents
	if priceExceedsMaximum {
		return Line{}, NewLinePriceTooLargeError(params.UnitPriceInCents)
	}
	return Line{
		productID:        params.ProductID,
		quantity:         params.Quantity,
		unitPriceInCents: params.UnitPriceInCents,
	}, nil
}

func (line Line) ProductID() string { return line.productID }

func (line Line) Quantity() Quantity { return line.quantity }

func (line Line) UnitPriceInCents() int { return line.unitPriceInCents }
