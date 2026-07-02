package organization

import "github.com/AymanKastali/citadel/internal/identity/domain"

// Organization is an entity: mutable, and the single owner of the rules that
// govern a tenant. State changes only through its methods, never by direct field
// assignment from outside the package. It embeds domain.Entity for its id and
// recorded events. It references no other entity — memberships, roles, and
// accounts point at the organization from their own packages.
type Organization struct {
	domain.Entity[ID]
	name   Name
	slug   Slug
	status Status
}

// CreateParams groups the already-valid value objects an organization is created
// from, so Create takes one data argument instead of a positional list.
type CreateParams struct {
	ID   ID
	Name Name
	Slug Slug
}

// Create builds a new organization, rejecting any missing (zero-value) field. The
// starting status is always Active — there is no policy or verification step —
// and the organization records OrganizationCreatedEvent.
func Create(params CreateParams) (*Organization, error) {
	idIsMissing := params.ID.IsZero()
	if idIsMissing {
		return nil, NewEmptyIDError()
	}
	nameIsMissing := params.Name.IsZero()
	if nameIsMissing {
		return nil, NewEmptyNameError()
	}
	slugIsMissing := params.Slug.IsZero()
	if slugIsMissing {
		return nil, NewEmptySlugError()
	}
	organization := &Organization{
		Entity: domain.NewEntity(params.ID),
		name:   params.Name,
		slug:   params.Slug,
		status: Active, // organizations always start active
	}
	organization.Record(NewOrganizationCreatedEvent(organization.ID(), organization.slug))
	return organization, nil
}

// ReconstituteParams groups the full persisted state of an organization, as read
// back from storage. Unlike CreateParams it also carries the Status, because
// rebuilding an organization restores where it already sits in its lifecycle.
type ReconstituteParams struct {
	ID     ID
	Name   Name
	Slug   Slug
	Status Status
}

// Reconstitute rebuilds an organization from stored state (repository adapter
// only). It just loads the persisted fields into a fresh entity — no validation,
// no event, no policy.
func Reconstitute(params ReconstituteParams) *Organization {
	return &Organization{
		Entity: domain.NewEntity(params.ID),
		name:   params.Name,
		slug:   params.Slug,
		status: params.Status,
	}
}

// Rename changes the display name. The Name value object is already valid, so
// there is nothing to re-validate and no failure mode — but the method returns
// error for a uniform command signature and forward compatibility.
func (organization *Organization) Rename(name Name) error {
	organization.name = name
	organization.Record(NewOrganizationRenamedEvent(organization.ID(), name))
	return nil
}

// ChangeSlug changes the URL-safe slug. As with Rename the Slug value object is
// already valid; uniqueness across organizations is not a domain invariant (it
// needs a store lookup) and is enforced at the application boundary before this
// call.
func (organization *Organization) ChangeSlug(slug Slug) error {
	organization.slug = slug
	organization.Record(NewOrganizationSlugChangedEvent(organization.ID(), slug))
	return nil
}

// Suspend blocks the organization, moving Active -> Suspended. It rejects an
// already-Suspended organization so the transition is idempotent-by-rejection and
// no spurious event is recorded.
func (organization *Organization) Suspend() error {
	alreadySuspended := organization.status == Suspended
	if alreadySuspended {
		return NewAlreadySuspendedError()
	}
	organization.status = Suspended
	organization.Record(NewOrganizationSuspendedEvent(organization.ID()))
	return nil
}

// Activate reactivates a suspended organization, moving Suspended -> Active. It
// rejects an already-Active organization.
func (organization *Organization) Activate() error {
	alreadyActive := organization.status == Active
	if alreadyActive {
		return NewAlreadyActiveError()
	}
	organization.status = Active
	organization.Record(NewOrganizationActivatedEvent(organization.ID()))
	return nil
}

func (organization *Organization) Name() Name { return organization.name }

func (organization *Organization) Slug() Slug { return organization.slug }

func (organization *Organization) Status() Status { return organization.status }
