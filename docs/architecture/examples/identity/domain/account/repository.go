package account

import "context"

// Repository is the account's repository port. The domain declares the
// interface; the concrete implementation lives in infrastructure.
type Repository interface {
	Create(ctx context.Context, account *Account) (*Account, error)
	Get(ctx context.Context, id ID) (*Account, error)
	Exists(ctx context.Context, id ID) (bool, error)
	Update(ctx context.Context, account *Account) error
	Delete(ctx context.Context, id ID) error
}
