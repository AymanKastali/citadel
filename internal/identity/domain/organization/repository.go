package organization

import "context"

// Repository is the organization's repository port. The domain declares the
// interface; the concrete implementation lives in infrastructure. Its methods use
// persistence verbs (not domain verbs), take a context.Context and typed value
// objects, and obey CQS — Exists / ExistsBySlug return data (a bool), the mutating
// verbs return only the outcome. Beyond the standard CRUD set it adds slug
// lookups: the slug is a natural secondary key (URL / sub-domain lookups, and
// uniqueness enforcement before Create / ChangeSlug). A "not found" from Get /
// GetBySlug is not a domain error; it is a repository outcome the application maps
// at the boundary.
type Repository interface {
	Create(ctx context.Context, organization *Organization) (*Organization, error)
	Get(ctx context.Context, id ID) (*Organization, error)
	Exists(ctx context.Context, id ID) (bool, error)
	Update(ctx context.Context, organization *Organization) error
	Delete(ctx context.Context, id ID) error
	GetBySlug(ctx context.Context, slug Slug) (*Organization, error)
	ExistsBySlug(ctx context.Context, slug Slug) (bool, error)
}
