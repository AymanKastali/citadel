// Package organization is the tenant at the center of the identity context: the
// isolation boundary every other identity concept is scoped to. It has a
// human-facing Name, a stable, URL-safe Slug (used in URLs, sub-domains, and
// lookups), and a lifecycle Status. It references no other entity — memberships,
// roles, and accounts point at the organization from their own packages.
//
// Lifecycle: an organization is created Active. Suspend blocks it (Active ->
// Suspended); Activate lifts the block (Suspended -> Active). There is no terminal
// status — removal is a repository Delete, not a lifecycle state.
//
// Reading order (front page first):
//   - organization.go                the Organization entity: constructors, commands, queries
//   - id.go, name.go, slug.go,       the value objects the entity is built from
//     status.go
//   - events.go                      the past-tense domain events the entity records
//   - errors.go                      the domain-error factories
//   - repository.go                  the persistence port (implemented in infrastructure)
//
// The package is pure domain: it imports only the standard library and the
// identity domain root, never a framework, driver, or transport.
package organization
