package membership

import (
	"github.com/AymanKastali/citadel/internal/identity/domain/account"
	"github.com/AymanKastali/citadel/internal/identity/domain/organization"
	"github.com/AymanKastali/citadel/internal/identity/domain/role"
)

// Event wire names: the stable, transport-facing identifier each membership event
// reports from EventName(). Grouped up top as one block rather than scattered one
// per event.
const (
	memberAddedEventName     = "membership.member_added"
	memberSuspendedEventName = "membership.member_suspended"
	memberActivatedEventName = "membership.member_activated"
	roleAssignedEventName    = "membership.role_assigned"
	roleRevokedEventName     = "membership.role_revoked"
)

// MemberAddedEvent is recorded when a membership is created. It carries the
// membership's id and the account and organization it links — not the entities
// themselves.
type MemberAddedEvent struct {
	membershipID   ID
	accountID      account.ID
	organizationID organization.ID
}

func NewMemberAddedEvent(membershipID ID, accountID account.ID, organizationID organization.ID) MemberAddedEvent {
	return MemberAddedEvent{membershipID: membershipID, accountID: accountID, organizationID: organizationID}
}

func (event MemberAddedEvent) MembershipID() ID { return event.membershipID }

func (event MemberAddedEvent) AccountID() account.ID { return event.accountID }

func (event MemberAddedEvent) OrganizationID() organization.ID { return event.organizationID }

func (event MemberAddedEvent) EventName() string { return memberAddedEventName }

// MemberSuspendedEvent is recorded by Suspend. It carries the membership's id.
type MemberSuspendedEvent struct {
	membershipID ID
}

func NewMemberSuspendedEvent(membershipID ID) MemberSuspendedEvent {
	return MemberSuspendedEvent{membershipID: membershipID}
}

func (event MemberSuspendedEvent) MembershipID() ID { return event.membershipID }

func (event MemberSuspendedEvent) EventName() string { return memberSuspendedEventName }

// MemberActivatedEvent is recorded by Activate. It carries the membership's id.
type MemberActivatedEvent struct {
	membershipID ID
}

func NewMemberActivatedEvent(membershipID ID) MemberActivatedEvent {
	return MemberActivatedEvent{membershipID: membershipID}
}

func (event MemberActivatedEvent) MembershipID() ID { return event.membershipID }

func (event MemberActivatedEvent) EventName() string { return memberActivatedEventName }

// RoleAssignedEvent is recorded when a role is added to the membership's set. It
// carries the membership's id and the assigned role's id.
type RoleAssignedEvent struct {
	membershipID ID
	roleID       role.ID
}

func NewRoleAssignedEvent(membershipID ID, roleID role.ID) RoleAssignedEvent {
	return RoleAssignedEvent{membershipID: membershipID, roleID: roleID}
}

func (event RoleAssignedEvent) MembershipID() ID { return event.membershipID }

func (event RoleAssignedEvent) RoleID() role.ID { return event.roleID }

func (event RoleAssignedEvent) EventName() string { return roleAssignedEventName }

// RoleRevokedEvent is recorded when a role is removed from the membership's set.
// It carries the membership's id and the revoked role's id.
type RoleRevokedEvent struct {
	membershipID ID
	roleID       role.ID
}

func NewRoleRevokedEvent(membershipID ID, roleID role.ID) RoleRevokedEvent {
	return RoleRevokedEvent{membershipID: membershipID, roleID: roleID}
}

func (event RoleRevokedEvent) MembershipID() ID { return event.membershipID }

func (event RoleRevokedEvent) RoleID() role.ID { return event.roleID }

func (event RoleRevokedEvent) EventName() string { return roleRevokedEventName }
