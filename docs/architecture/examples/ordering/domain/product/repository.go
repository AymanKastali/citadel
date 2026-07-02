package product

import "context"

// Repository is the product's repository port. The domain declares the
// interface; the concrete implementation lives in infrastructure.
type Repository interface {
	Create(ctx context.Context, product *Product) (*Product, error)
	Get(ctx context.Context, id ID) (*Product, error)
	Exists(ctx context.Context, id ID) (bool, error)
	Update(ctx context.Context, product *Product) error
	Delete(ctx context.Context, id ID) error
}
