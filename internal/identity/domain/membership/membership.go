package membership

import (
	"github.com/AymanKastali/citadel/internal/identity/domain"
	"github.com/AymanKastali/citadel/internal/identity/domain/account"
	"github.com/AymanKastali/citadel/internal/identity/domain/organization"
	"github.com/AymanKastali/citadel/internal/identity/domain/role"
)

// Membership is an entity: mutable, and the single owner of the rules that govern
// the link between an account and an organization. State changes only through its
// methods, never by direct field assignment from outside the package. It embeds
// domain.Entity for its id and recorded events, and references the account, the
// organization, and the account's roles within that organization by typed id only
// — never an embedded Account, Organization, or Role entity.
type Membership struct {
	domain.Entity[ID]
	accountID      account.ID
	organizationID organization.ID
	status         Status
	roleIDs        []role.ID // set semantics — no duplicates
}

// CreateParams groups the already-valid value objects a membership is created
// from, so Create takes one data argument instead of a positional list.
type CreateParams struct {
	ID             ID
	AccountID      account.ID
	OrganizationID organization.ID
}

// Create builds a new membership, rejecting any missing (zero-value) field. There
// is no invitation flow in the MVP, so a new membership always starts Active with
// an empty role set — roles are added later via AssignRole. On success it records
// MemberAddedEvent.
func Create(params CreateParams) (*Membership, error) {
	idIsMissing := params.ID.IsZero()
	if idIsMissing {
		return nil, NewEmptyIDError()
	}
	accountIsMissing := params.AccountID.IsZero()
	if accountIsMissing {
		return nil, NewEmptyAccountIDError()
	}
	organizationIsMissing := params.OrganizationID.IsZero()
	if organizationIsMissing {
		return nil, NewEmptyOrganizationIDError()
	}
	membership := &Membership{
		Entity:         domain.NewEntity(params.ID),
		accountID:      params.AccountID,
		organizationID: params.OrganizationID,
		status:         Active,      // no invitation flow in MVP — active immediately
		roleIDs:        []role.ID{}, // roles are assigned later, never at creation
	}
	membership.Record(NewMemberAddedEvent(membership.ID(), membership.accountID, membership.organizationID))
	return membership, nil
}

// ReconstituteParams groups the full persisted state of a membership, as read back
// from storage. Unlike CreateParams it also carries the Status and the RoleIDs
// set, because rebuilding a membership restores where it already sits in its
// lifecycle and which roles it already holds. RoleIDs is rehydrated by the
// repository from the pivot table.
type ReconstituteParams struct {
	ID             ID
	AccountID      account.ID
	OrganizationID organization.ID
	Status         Status
	RoleIDs        []role.ID
}

// Reconstitute rebuilds a membership from stored state (repository adapter only).
// It just loads the persisted fields into a fresh entity — no validation, no
// event, no policy.
func Reconstitute(params ReconstituteParams) *Membership {
	return &Membership{
		Entity:         domain.NewEntity(params.ID),
		accountID:      params.AccountID,
		organizationID: params.OrganizationID,
		status:         params.Status,
		roleIDs:        params.RoleIDs,
	}
}

// Suspend blocks the membership, moving Active -> Suspended. It rejects an
// already-Suspended membership so the transition is idempotent-by-rejection and no
// spurious event is recorded.
func (membership *Membership) Suspend() error {
	if membership.isSuspended() {
		return NewAlreadySuspendedError()
	}
	membership.status = Suspended
	membership.Record(NewMemberSuspendedEvent(membership.ID()))
	return nil
}

// Activate reactivates a suspended membership, moving Suspended -> Active. It
// rejects an already-Active membership.
func (membership *Membership) Activate() error {
	if membership.isActive() {
		return NewAlreadyActiveError()
	}
	membership.status = Active
	membership.Record(NewMemberActivatedEvent(membership.ID()))
	return nil
}

func (membership *Membership) isSuspended() bool { return membership.status == Suspended }

func (membership *Membership) isActive() bool { return membership.status == Active }

// AssignRole adds a role to the membership's role set, rejecting a duplicate so
// the set stays free of repeats. On success it records RoleAssignedEvent.
//
// It does NOT check that the role belongs to the membership's own organization —
// a membership holds only a role.ID, not the role.Role entity, so it cannot see
// the role's organizationID. That cross-entity rule is the job of the root-level
// domain service AssignRoleToMembership; this method guards only the single-entity
// invariant it can see on its own.
func (membership *Membership) AssignRole(roleID role.ID) error {
	if membership.HasRole(roleID) {
		return NewRoleAlreadyAssignedError()
	}
	membership.roleIDs = append(membership.roleIDs, roleID)
	membership.Record(NewRoleAssignedEvent(membership.ID(), roleID))
	return nil
}

// RevokeRole removes a role from the membership's role set, rejecting a role that
// is not present. On success it records RoleRevokedEvent.
func (membership *Membership) RevokeRole(roleID role.ID) error {
	if !membership.HasRole(roleID) {
		return NewRoleNotAssignedError()
	}
	membership.roleIDs = removeRoleID(membership.roleIDs, roleID)
	membership.Record(NewRoleRevokedEvent(membership.ID(), roleID))
	return nil
}

// removeRoleID returns a new slice containing every id in roleIDs except target,
// preserving order. It allocates a fresh backing array rather than filtering
// roleIDs in place, so no caller holding a copy of the old slice header can
// observe the mutation through shared backing storage.
func removeRoleID(roleIDs []role.ID, target role.ID) []role.ID {
	filtered := make([]role.ID, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		if roleID.Equal(target) {
			continue
		}
		filtered = append(filtered, roleID)
	}
	return filtered
}

func (membership *Membership) AccountID() account.ID { return membership.accountID }

func (membership *Membership) OrganizationID() organization.ID { return membership.organizationID }

func (membership *Membership) Status() Status { return membership.status }

// Roles returns a copy so callers cannot mutate the membership's internal role set
// behind its back.
func (membership *Membership) Roles() []role.ID {
	copied := make([]role.ID, len(membership.roleIDs))
	copy(copied, membership.roleIDs)
	return copied
}

// HasRole reports whether roleID is in the membership's role set. It is used by
// AssignRole and RevokeRole to guard their invariants, and exported for callers
// that need to check membership without mutating it.
func (membership *Membership) HasRole(roleID role.ID) bool {
	for _, assigned := range membership.roleIDs {
		if assigned.Equal(roleID) {
			return true
		}
	}
	return false
}
