package permission

import (
	"github.com/AymanKastali/citadel/internal/identity/domain"
	"github.com/AymanKastali/citadel/internal/identity/domain/role"
)

// Permission is an entity: the single owner of the rules that govern one
// authorization scope granted to a role. It embeds domain.Entity for its id and
// recorded events, and references the owning role by typed role.ID only — never an
// embedded Role entity. It is immutable once granted: its state is set at
// construction and there are no mutating methods. Callers read via getters, never
// by assigning fields.
type Permission struct {
	domain.Entity[ID]
	roleID role.ID
	scope  Scope
}

// GrantParams groups the already-valid value objects a permission is granted from
// (its own id, the owning role.ID, and the scope), so Grant takes one data argument
// instead of a positional list.
type GrantParams struct {
	ID     ID
	RoleID role.ID
	Scope  Scope
}

// Grant builds a new permission, rejecting any missing (zero-value) field —
// including the required roleID. The value objects are already valid
// (empty/too-long/malformed are rejected in their own constructors), so these
// guards defend the entity's own invariant, that none of its fields is the zero
// value, and return the "empty" errors. On success it records
// PermissionGrantedEvent. There is no policy — nothing about granting a permission
// varies by deployment configuration.
func Grant(params GrantParams) (*Permission, error) {
	idIsMissing := params.ID.IsZero()
	if idIsMissing {
		return nil, NewEmptyIDError()
	}
	roleIDIsMissing := params.RoleID.IsZero()
	if roleIDIsMissing {
		return nil, NewEmptyRoleIDError()
	}
	scopeIsMissing := params.Scope.IsZero()
	if scopeIsMissing {
		return nil, NewEmptyScopeError()
	}
	permission := &Permission{
		Entity: domain.NewEntity(params.ID),
		roleID: params.RoleID,
		scope:  params.Scope,
	}
	permission.Record(NewPermissionGrantedEvent(permission.ID(), permission.roleID, permission.scope))
	return permission, nil
}

// ReconstituteParams groups the full persisted state of a permission. A permission
// has no lifecycle state beyond its three construction fields, so this mirrors
// GrantParams exactly, but it remains its own struct per the reconstitution rule.
type ReconstituteParams struct {
	ID     ID
	RoleID role.ID
	Scope  Scope
}

// Reconstitute rebuilds a permission from stored state (repository adapter only).
// It just loads the persisted fields into a fresh entity — no validation, no event,
// no policy.
func Reconstitute(params ReconstituteParams) *Permission {
	return &Permission{
		Entity: domain.NewEntity(params.ID),
		roleID: params.RoleID,
		scope:  params.Scope,
	}
}

func (permission *Permission) RoleID() role.ID { return permission.roleID }

func (permission *Permission) Scope() Scope { return permission.scope }
