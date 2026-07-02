package membership

import (
	"context"

	"github.com/AymanKastali/citadel/internal/identity/domain/account"
	"github.com/AymanKastali/citadel/internal/identity/domain/organization"
)

// Repository is the membership's repository port. The domain declares the
// interface; the concrete implementation lives in infrastructure. Its methods use
// persistence verbs (not domain verbs), take a context.Context and typed value
// objects, and obey CQS — Exists / GetByAccountAndOrganization / ListByAccount /
// ListByOrganization return data, the mutating verbs return only the outcome.
// GetByAccountAndOrganization backs the uniqueness lookup (one membership per
// account+organization); ListByAccount and ListByOrganization enumerate the two
// list directions. The adapter is responsible for loading and persisting the
// roleIDs set via the pivot table on Get, Create, and Update. A "not found" from
// Get or GetByAccountAndOrganization is not a domain error; it is a repository
// outcome the application maps at the boundary.
type Repository interface {
	Create(ctx context.Context, membership *Membership) (*Membership, error)
	Get(ctx context.Context, id ID) (*Membership, error)
	Exists(ctx context.Context, id ID) (bool, error)
	Update(ctx context.Context, membership *Membership) error
	Delete(ctx context.Context, id ID) error
	GetByAccountAndOrganization(ctx context.Context, accountID account.ID, organizationID organization.ID) (*Membership, error)
	ListByAccount(ctx context.Context, accountID account.ID) ([]*Membership, error)
	ListByOrganization(ctx context.Context, organizationID organization.ID) ([]*Membership, error)
}
