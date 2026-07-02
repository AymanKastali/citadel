// Package role is an org-scoped RBAC role: a named grouping of permissions within
// a single organization. It holds a required organization.ID (roles are always
// scoped to one organization — there are no global roles in the MVP), a
// human-readable Name, and an optional Description. It does not hold a list of
// permissions: the permission entity owns the foreign key (roleID), giving
// Role 1-to-* Permission. A role has no lifecycle status.
//
// Reading order (front page first):
//   - role.go                        the Role entity: constructors, commands, queries
//   - id.go, name.go, description.go  the value objects the entity is built from
//   - events.go                      the past-tense domain events the entity records
//   - errors.go                      the domain-error factories
//   - repository.go                  the persistence port (implemented in infrastructure)
//
// The package is pure domain: it imports only the standard library, the identity
// domain root, and the organization package (for the typed organization.ID it
// references) — never a framework, driver, or transport.
package role
