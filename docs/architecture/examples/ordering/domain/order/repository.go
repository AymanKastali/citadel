package order

import "context"

// Repository is the order's repository port. The domain declares the
// interface; the concrete implementation lives in infrastructure.
type Repository interface {
	Create(ctx context.Context, order *Order) (*Order, error)
	Get(ctx context.Context, id ID) (*Order, error)
	Exists(ctx context.Context, id ID) (bool, error)
	Update(ctx context.Context, order *Order) error
	Delete(ctx context.Context, id ID) error
}
