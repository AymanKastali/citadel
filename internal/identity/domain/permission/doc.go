// Package permission is a single authorization scope granted to a role — a string
// like "users:read" or "orgs:write". It is the dependent, owning side of the RBAC
// relationship: a permission holds a required roleID (the foreign key), so Role is
// 1-to-* Permission with no permission list on Role. It references the owning role
// by typed role.ID only — never an embedded Role entity.
//
// A permission is immutable once granted: it has no lifecycle status, no
// re-targeting, and no rename, so it exposes no commands. Changing authorization is
// expressed as delete the old permission + grant a new one, never an in-place
// mutation — which is why the repository port has no Update and the only event is
// PermissionGrantedEvent.
//
// Reading order (front page first):
//   - permission.go    the Permission entity: constructors and queries (no commands)
//   - id.go, scope.go  the value objects the entity is built from
//   - events.go        the past-tense domain event the entity records
//   - errors.go        the domain-error factories
//   - repository.go    the persistence port (implemented in infrastructure)
//
// The package is pure domain: it imports only the standard library, the identity
// domain root, and the role package (for the typed role.ID it references) — never a
// framework, driver, or transport.
package permission
