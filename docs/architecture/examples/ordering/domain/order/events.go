package order

const orderShippedEventName = "order.shipped"

// OrderShippedEvent is recorded when an order ships. It is an immutable domain event
// and carries the order's id, not the order itself — events reference other
// entities by id, exactly like the entities do.
type OrderShippedEvent struct {
	orderID ID
}

func NewOrderShippedEvent(orderID ID) OrderShippedEvent { return OrderShippedEvent{orderID: orderID} }

func (event OrderShippedEvent) OrderID() ID { return event.orderID }

func (event OrderShippedEvent) EventName() string { return orderShippedEventName }
