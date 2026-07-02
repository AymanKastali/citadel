package order

// maxQuantity is the upper bound of a valid quantity: high enough for any real
// order line, low enough to reject overflow and fat-finger input.
const maxQuantity = 10_000

// Quantity is how many of a product an order line holds. It is a value object:
// immutable and always within its valid range.
type Quantity struct {
	value int
}

func NewQuantity(value int) (Quantity, error) {
	quantityIsZeroOrNegative := value <= 0
	if quantityIsZeroOrNegative {
		return Quantity{}, NewInvalidQuantityError(value)
	}
	quantityExceedsMaximum := value > maxQuantity
	if quantityExceedsMaximum {
		return Quantity{}, NewQuantityTooLargeError(value)
	}
	return Quantity{value: value}, nil
}

func (quantity Quantity) Value() int { return quantity.value }

func (quantity Quantity) Equal(other Quantity) bool { return quantity.value == other.value }

func (quantity Quantity) IsZero() bool { return quantity == Quantity{} }
