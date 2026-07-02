package order

import "github.com/AymanKastali/citadel/internal/ordering/domain"

// Status is where an order is in its lifecycle.
type Status int

const (
	Open Status = iota
	Shipped
)

// Order is an entity: mutable, and the owner of the rules that decide when its
// lines may change and when it may ship. It embeds domain.Entity for its id and
// recorded events.
type Order struct {
	domain.Entity[ID]
	status Status
	lines  []Line
}

// NewOrderParams groups the value objects an Order is built from.
type NewOrderParams struct {
	ID ID
}

// NewOrder builds an empty, open order.
func NewOrder(params NewOrderParams) (*Order, error) {
	idIsMissing := params.ID.IsZero()
	if idIsMissing {
		return nil, NewEmptyIDError()
	}
	return &Order{Entity: domain.NewEntity(params.ID), status: Open}, nil
}

// ReconstituteParams groups the full persisted state of an order. Unlike
// NewOrderParams it carries the status and the lines, because rebuilding an order
// restores where it already is — not the empty, open order NewOrder starts from.
type ReconstituteParams struct {
	ID     ID
	Status Status
	Lines  []Line
}

// Reconstitute rebuilds an order from stored state (repository adapter only). It just
// loads the persisted fields into a fresh entity — no validation, no event, no policy.
func Reconstitute(params ReconstituteParams) *Order {
	return &Order{
		Entity: domain.NewEntity(params.ID),
		status: params.Status,
		lines:  params.Lines,
	}
}

func (order *Order) Status() Status { return order.status }

// Lines returns a copy so callers cannot mutate the order's state behind its
// back.
func (order *Order) Lines() []Line {
	copied := make([]Line, len(order.lines))
	copy(copied, order.lines)
	return copied
}

// AddLine adds a line, enforcing that a shipped order can no longer change.
func (order *Order) AddLine(line Line) error {
	if order.hasShipped() {
		return NewAlreadyShippedError()
	}
	order.lines = append(order.lines, line)
	return nil
}

// Ship marks the order shipped, enforcing that it has something to ship and is
// not already shipped.
func (order *Order) Ship() error {
	if order.hasShipped() {
		return NewAlreadyShippedError()
	}
	if order.isEmpty() {
		return NewEmptyOrderError()
	}
	order.status = Shipped
	order.Record(NewOrderShippedEvent(order.ID()))
	return nil
}

func (order *Order) hasShipped() bool { return order.status == Shipped }

func (order *Order) isEmpty() bool { return len(order.lines) == 0 }
