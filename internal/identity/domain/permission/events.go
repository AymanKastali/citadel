package permission

import "github.com/AymanKastali/citadel/internal/identity/domain/role"

// Event wire names: the stable, transport-facing identifier each permission event
// reports from EventName(). Grouped up top as one block rather than scattered one
// per event. A permission has exactly one fact — it was granted — so there is a
// single name here. (Deletion is a persistence outcome driven by the application,
// not a state transition of a still-living entity, so there is no revoked event in
// the MVP.)
const permissionGrantedEventName = "permission.granted"

// PermissionGrantedEvent is recorded when a scope is granted to a role. It carries
// the permission's id, the owning role's id, and the granted scope value — not the
// entities themselves.
type PermissionGrantedEvent struct {
	permissionID ID
	roleID       role.ID
	scope        Scope
}

func NewPermissionGrantedEvent(permissionID ID, roleID role.ID, scope Scope) PermissionGrantedEvent {
	return PermissionGrantedEvent{permissionID: permissionID, roleID: roleID, scope: scope}
}

func (event PermissionGrantedEvent) PermissionID() ID { return event.permissionID }

func (event PermissionGrantedEvent) RoleID() role.ID { return event.roleID }

func (event PermissionGrantedEvent) Scope() Scope { return event.scope }

func (event PermissionGrantedEvent) EventName() string { return permissionGrantedEventName }
