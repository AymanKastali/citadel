package order

const orderShippedEventName = "order.shipped"

// OrderShipped is recorded when an order ships. It is an immutable domain event
// and carries the order's id, not the order itself — events reference other
// entities by id, exactly like the entities do.
type OrderShipped struct {
	orderID ID
}

func NewOrderShipped(orderID ID) OrderShipped { return OrderShipped{orderID: orderID} }

func (event OrderShipped) OrderID() ID { return event.orderID }

func (event OrderShipped) EventName() string { return orderShippedEventName }
