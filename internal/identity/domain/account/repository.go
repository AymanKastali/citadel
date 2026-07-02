package account

import "context"

// Repository is the account's repository port. The domain declares the
// interface; the concrete implementation lives in infrastructure. Its methods use
// persistence verbs (not domain verbs), take a context.Context and typed value
// objects, and obey CQS — Exists / ExistsByEmail return data (a bool), the
// mutating verbs return only the outcome. A "not found" from Get / GetByEmail is
// not a domain error; it is a repository outcome the application maps at the
// boundary.
type Repository interface {
	Create(ctx context.Context, account *Account) (*Account, error)
	Get(ctx context.Context, id ID) (*Account, error)
	Exists(ctx context.Context, id ID) (bool, error)
	Update(ctx context.Context, account *Account) error
	Delete(ctx context.Context, id ID) error
	GetByEmail(ctx context.Context, email Email) (*Account, error)
	ExistsByEmail(ctx context.Context, email Email) (bool, error)
}
