package permission

import (
	"context"

	"github.com/AymanKastali/citadel/internal/identity/domain/role"
)

// Repository is the permission's repository port. The domain declares the
// interface; the concrete implementation lives in infrastructure. Its methods use
// persistence verbs (not domain verbs), take a context.Context and typed value
// objects, and obey CQS — Exists / ExistsByScopeInRole / ListByRole return data,
// the mutating verbs return only the outcome. There is deliberately no Update:
// permissions are immutable once granted, so there is nothing to update — a change
// in authorization is a Delete plus a fresh Create. Two role-scoped reads support
// the RBAC use cases. A "not found" from Get is not a domain error; it is a
// repository outcome the application maps at the boundary.
type Repository interface {
	Create(ctx context.Context, permission *Permission) (*Permission, error)
	Get(ctx context.Context, id ID) (*Permission, error)
	Exists(ctx context.Context, id ID) (bool, error)
	Delete(ctx context.Context, id ID) error
	ListByRole(ctx context.Context, roleID role.ID) ([]*Permission, error)
	ExistsByScopeInRole(ctx context.Context, roleID role.ID, scope Scope) (bool, error)
}
