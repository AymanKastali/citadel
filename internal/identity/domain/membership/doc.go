// Package membership is the link between an account and an organization: a real
// entity with its own identity, an Active/Suspended lifecycle, and the account's
// set of roles within that organization. The account itself holds no organization
// or roles; those live here. It references the account, the organization, and each
// assigned role by typed id only (account.ID, organization.ID, []role.ID) — never
// an embedded entity. The member<->role many-to-many is held as a set of role.ID on
// the membership; the join is a persistence-only pivot table, never a domain
// entity.
//
// Lifecycle: a membership is created Active — there is no invitation flow in the
// MVP — with an empty role set. Suspend blocks it (Active -> Suspended); Activate
// lifts the block (Suspended -> Active). There is no terminal status — removal is a
// repository Delete, not a lifecycle state. AssignRole and RevokeRole add and
// remove roles from the set, guarding only duplicates/absence; the stronger
// cross-organization rule (a role must belong to the member's own organization) is
// enforced by a separate root-level domain service, not by this entity.
//
// Reading order (front page first):
//   - membership.go     the Membership entity: constructors, commands, queries
//   - id.go, status.go  the value objects the entity is built from
//   - events.go         the past-tense domain events the entity records
//   - errors.go         the domain-error factories
//   - repository.go     the persistence port (implemented in infrastructure)
//
// The package is pure domain: it imports only the standard library, the identity
// domain root, and the account, organization, and role packages (for the typed ids
// they contribute) — never a framework, driver, or transport.
package membership
