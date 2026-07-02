package role

import "github.com/AymanKastali/citadel/internal/identity/domain/organization"

// Event wire names: the stable, transport-facing identifier each role event
// reports from EventName(). Grouped up top as one block rather than scattered one
// per event.
const (
	roleCreatedEventName            = "role.created"
	roleRenamedEventName            = "role.renamed"
	roleDescriptionUpdatedEventName = "role.description_updated"
)

// RoleCreatedEvent is recorded when a role is created. It carries the role's id,
// its organization's id, and its name — not the role itself.
type RoleCreatedEvent struct {
	roleID         ID
	organizationID organization.ID
	name           Name
}

func NewRoleCreatedEvent(roleID ID, organizationID organization.ID, name Name) RoleCreatedEvent {
	return RoleCreatedEvent{roleID: roleID, organizationID: organizationID, name: name}
}

func (event RoleCreatedEvent) RoleID() ID { return event.roleID }

func (event RoleCreatedEvent) OrganizationID() organization.ID { return event.organizationID }

func (event RoleCreatedEvent) Name() Name { return event.name }

func (event RoleCreatedEvent) EventName() string { return roleCreatedEventName }

// RoleRenamedEvent is recorded by Rename. It carries the role's id and the new name.
type RoleRenamedEvent struct {
	roleID ID
	name   Name
}

func NewRoleRenamedEvent(roleID ID, name Name) RoleRenamedEvent {
	return RoleRenamedEvent{roleID: roleID, name: name}
}

func (event RoleRenamedEvent) RoleID() ID { return event.roleID }

func (event RoleRenamedEvent) Name() Name { return event.name }

func (event RoleRenamedEvent) EventName() string { return roleRenamedEventName }

// RoleDescriptionUpdatedEvent is recorded by UpdateDescription. It carries the
// role's id and the new description (which may be the zero-value Description).
type RoleDescriptionUpdatedEvent struct {
	roleID      ID
	description Description
}

func NewRoleDescriptionUpdatedEvent(roleID ID, description Description) RoleDescriptionUpdatedEvent {
	return RoleDescriptionUpdatedEvent{roleID: roleID, description: description}
}

func (event RoleDescriptionUpdatedEvent) RoleID() ID { return event.roleID }

func (event RoleDescriptionUpdatedEvent) Description() Description { return event.description }

func (event RoleDescriptionUpdatedEvent) EventName() string { return roleDescriptionUpdatedEventName }
