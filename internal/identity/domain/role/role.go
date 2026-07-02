package role

import (
	"github.com/AymanKastali/citadel/internal/identity/domain"
	"github.com/AymanKastali/citadel/internal/identity/domain/organization"
)

// Role is an entity: mutable, and the single owner of the rules that govern an
// org-scoped RBAC role. State changes only through its methods, never by direct
// field assignment from outside the package. It embeds domain.Entity for its id
// and recorded events, and references its organization by typed id only — never an
// embedded Organization entity. A role has no lifecycle status.
type Role struct {
	domain.Entity[ID]
	organizationID organization.ID
	name           Name
	description    Description
}

// CreateParams groups the already-valid value objects (plus the required typed
// organization.ID) a role is created from, so Create takes one data argument.
type CreateParams struct {
	ID             ID
	OrganizationID organization.ID
	Name           Name
	Description    Description
}

// Create builds a new role, rejecting any missing required field (id,
// organizationID, name). The Description is NOT guarded — an empty Description is a
// valid "no description" state, so it is stored as-is. On success the role records
// RoleCreatedEvent.
func Create(params CreateParams) (*Role, error) {
	idIsMissing := params.ID.IsZero()
	if idIsMissing {
		return nil, NewEmptyIDError()
	}
	organizationIDIsMissing := params.OrganizationID.IsZero()
	if organizationIDIsMissing {
		return nil, NewEmptyOrganizationIDError()
	}
	nameIsMissing := params.Name.IsZero()
	if nameIsMissing {
		return nil, NewEmptyNameError()
	}
	role := &Role{
		Entity:         domain.NewEntity(params.ID),
		organizationID: params.OrganizationID,
		name:           params.Name,
		description:    params.Description, // may be the zero-value Description — that is valid
	}
	role.Record(NewRoleCreatedEvent(role.ID(), role.organizationID, role.name))
	return role, nil
}

// ReconstituteParams groups the full persisted state of a role. A role has no
// derived lifecycle state, so its fields match CreateParams one-for-one, but it
// remains its own struct per the reconstitution rule.
type ReconstituteParams struct {
	ID             ID
	OrganizationID organization.ID
	Name           Name
	Description    Description
}

// Reconstitute rebuilds a role from stored state (repository adapter only). It
// just loads the persisted fields into a fresh entity — no validation, no event,
// no policy.
func Reconstitute(params ReconstituteParams) *Role {
	return &Role{
		Entity:         domain.NewEntity(params.ID),
		organizationID: params.OrganizationID,
		name:           params.Name,
		description:    params.Description,
	}
}

// Rename replaces the role's name with a new already-valid Name, then records
// RoleRenamedEvent. The organizationID is immutable — a role never moves between
// organizations, so no command changes it.
func (role *Role) Rename(name Name) error {
	role.name = name
	role.Record(NewRoleRenamedEvent(role.ID(), role.name))
	return nil
}

// UpdateDescription replaces the description with a new Description — which may be
// the zero value, the valid way to clear a description — then records
// RoleDescriptionUpdatedEvent.
func (role *Role) UpdateDescription(description Description) error {
	role.description = description // may be the zero-value Description — clearing is valid
	role.Record(NewRoleDescriptionUpdatedEvent(role.ID(), role.description))
	return nil
}

func (role *Role) OrganizationID() organization.ID { return role.organizationID }

func (role *Role) Name() Name { return role.name }

func (role *Role) Description() Description { return role.description }
