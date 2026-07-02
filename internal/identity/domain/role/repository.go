package role

import (
	"context"

	"github.com/AymanKastali/citadel/internal/identity/domain/organization"
)

// Repository is the role's repository port. The domain declares the interface;
// the concrete implementation lives in infrastructure. Its methods use persistence
// verbs (not domain verbs), take a context.Context and typed value objects, and
// obey CQS — Exists / ExistsByNameInOrganization return data (a bool), the
// mutating verbs return only the outcome. Two lookups are org-scoped:
// ListByOrganization enumerates a tenant's roles, and ExistsByNameInOrganization
// supports per-organization name-uniqueness checks at the application boundary. A
// "not found" from Get is not a domain error; it is a repository outcome the
// application maps at the boundary.
type Repository interface {
	Create(ctx context.Context, role *Role) (*Role, error)
	Get(ctx context.Context, id ID) (*Role, error)
	Exists(ctx context.Context, id ID) (bool, error)
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id ID) error
	ListByOrganization(ctx context.Context, organizationID organization.ID) ([]*Role, error)
	ExistsByNameInOrganization(ctx context.Context, organizationID organization.ID, name Name) (bool, error)
}
