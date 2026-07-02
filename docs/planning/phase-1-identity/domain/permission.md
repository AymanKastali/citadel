# Identity · Domain · `permission`

> **Status: complete — pieces (a) value objects, (b) entity + constructors, (c) commands, and (d) queries, events, errors & repository port all done.**
> Full deep-dive for the `permission` entity. Part of the Identity domain layer; see the
> [overview](../domain-layer.md) (RBAC decision: `Permission` owns the `roleID` FK,
> giving `Role` 1—* `Permission`), the rules in
> [`domain-layer.md`](../../../architecture/domain-layer.md), and the sibling
> [`account.md`](./account.md) whose structure this mirrors.

## What `permission` is

A **single authorization scope granted to a role** — a string like `users:read` or
`orgs:write`. It is the dependent, owning side of the RBAC relationship: a permission
holds a **required `roleID`** (the foreign key), so `Role` is 1—* `Permission` with no
permission list on `Role` (see the [overview](../domain-layer.md)). A permission is
**immutable once granted**: it carries no lifecycle beyond existence, so the only ways
to change authorization are to grant a new one or delete an existing one. Package:
`internal/identity/domain/permission`.

## Build order (bottom-up)

- **(a) Value objects** — `ID`, `Scope` (+ typed `role.ID` reference) ✅
- **(b) Entity struct + `Grant` + `Reconstitute`** ✅
- **(c) Commands** — none (a permission is immutable; change = delete + grant) ✅
- **(d) Queries, events, errors, repository port** ✅

---

## (a) Value objects

One file per value object in the `permission` package. Each is immutable (unexported
field), self-validating in its constructor, with `Value()` / `Equal(other)` /
`IsZero()`, and a zero value that is invalid. Business/invariant failures return
`*domain.DomainError` via a factory in `permission/errors.go` (piece **d**).

### `ID` — `id.go` (same shape as `account.ID`)

The permission's own identity, mirroring
[`account.ID`](./account.md#id--idgo-reuse-verbatim-from-the-example).

- **Wraps:** `string` (an opaque identifier — UUID/ULID, assigned by the application,
  not generated in the domain).
- **Constructor:** `NewID(raw string) (ID, error)`.
- **Rules:** trim surrounding whitespace; reject empty (`NewEmptyIDError`); reject
  length `> maxIDLength` where `maxIDLength = 64` (`NewIDTooLongError(length)`).
- **Methods:** `Value() string`, `Equal(other ID) bool`, `IsZero() bool`.

### `Scope` — `scope.go` (new)

The permission string itself — the resource/action pair the role is authorized for.
This is the heart of the entity: it must be stable for equality and uniqueness (a role
must not hold the same scope twice, checked in piece **d**), so it is normalized and
strictly formatted.

- **Wraps:** `string` (the encoded scope, e.g. `users:read`).
- **Constructor:** `NewScope(raw string) (Scope, error)`.
- **Rules:**
  - trim surrounding whitespace;
  - reject empty (`NewEmptyScopeError`);
  - **normalize to lower case** *before* the length and format checks, so equality and
    per-role uniqueness are case-insensitive and stable;
  - reject length `> maxScopeLength` where `maxScopeLength = 100`
    (`NewScopeTooLongError(length)`) — comfortably fits any realistic
    `resource:action` (or deeper) scope while rejecting junk;
  - reject anything not matching the **scope format** below as malformed
    (`NewMalformedScopeError(value)`).
- **Format:** `^[a-z0-9]+(:[a-z0-9]+)+$` — **one or more** colon-separated segments,
  each a non-empty run of lowercase alphanumerics, with **at least two** segments
  (i.e. at least one colon). Justification:
  - the trailing `(:[a-z0-9]+)+` (a `+`, not a `*`) forces a minimum of two segments,
    so a bare `users` with no action is rejected — a scope is always a
    `resource:action` pair (deeper scopes like `orgs:members:invite` are allowed);
  - segments are `[a-z0-9]+`, so empty segments (`users:`, `:read`, `users::read`) and
    leading/trailing colons are rejected, and there is no separator ambiguity;
  - the alphabet is intentionally narrow (lowercase alphanumerics only — no `_`, `-`,
    `.`, or wildcards) to keep scopes unambiguous and normalization total; the lower
    case is guaranteed by the normalization step above, so the pattern anchors it
    rather than doing case folding.
  - The regexp is compiled once as a package-level `var` from a `const` pattern (stdlib
    `regexp` only — no third-party import).
- **Methods:** `Value() string`, `Equal(other Scope) bool`, `IsZero() bool`.

### Cross-entity reference — `role.ID` (typed, imported)

A permission holds a **required `roleID`** of type
[`role.ID`](./role.md), imported from the sibling `role` package. It is referenced by
its **typed ID value object**, never by embedding the `Role` entity (see the
[relationship map](../domain-layer.md#relationship-map)). The `role.ID` arrives already
valid; the entity constructor only guards that it is non-zero (piece **b**).

---

## (b) Entity & constructors

File: `permission/permission.go`.

### Struct

The entity embeds the base `domain.Entity[ID]` (id + recorded events) and holds its
state as value objects only. It references the owning role by typed id — never by
embedding the entity.

```go
type Permission struct {
	domain.Entity[ID]
	roleID role.ID
	scope  Scope
}
```

State is set once at construction; there are no mutating methods (piece **c**). Callers
read via getters (piece **d**), never by assigning fields.

### `Grant` — the business constructor

Builds a **new** permission from already-valid value objects: the permission's own
`ID`, the owning `role.ID`, and the `Scope`. Each field is guard-claused (Fail Fast) —
including the required `roleID` — and the permission records `PermissionGrantedEvent`
(payload defined in piece **d**). There is no policy: nothing about granting a
permission varies by deployment configuration.

```go
type GrantParams struct {
	ID     ID
	RoleID role.ID
	Scope  Scope
}

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
```

> **On the guard clauses.** The value objects are already valid (empty/too-long/malformed
> scope, empty id are all rejected in their own constructors, piece **a**). The guards
> here defend the *entity's* invariant — that none of its fields is the zero value — and
> so return the "empty" errors (`NewEmptyIDError`, `NewEmptyRoleIDError`,
> `NewEmptyScopeError`), never the format/length ones.

### `Reconstitute` — rebuild from persistence

The counterpart used only by the repository adapter. Because a permission has no
lifecycle state beyond its three construction fields, `ReconstituteParams` mirrors
`GrantParams` exactly. It **just loads** the stored fields into a fresh entity: no
validation, no event, no policy, no error (see the
[Reconstitution rule](../../../architecture/domain-layer.md#reconstitution-rebuilding-from-persistence)).

```go
type ReconstituteParams struct {
	ID     ID
	RoleID role.ID
	Scope  Scope
}

func Reconstitute(params ReconstituteParams) *Permission {
	return &Permission{
		Entity: domain.NewEntity(params.ID),
		roleID: params.RoleID,
		scope:  params.Scope,
	}
}
```

## (c) Commands

**None.** A permission is **immutable once granted**: it has no status, no
re-targeting, and no rename. Neither its `scope` nor its `roleID` may change after
`Grant` — reassigning a scope to a different role, or changing the scope string, is a
different authorization fact, expressed as **delete the old permission + grant a new
one**, not an in-place mutation.

Consequences that ripple through the rest of the entity:

- there are **no state-transition events** beyond `PermissionGrantedEvent` (piece **d**);
- the repository port has **no `Update`** (piece **d**) — there is nothing to update;
- the entity exposes only getters, and the base's `PullEvents`/`DrainEvents`.

## (d) Queries, events, errors, repository

### Queries (CQS — return data, never mutate)

`ID()` is promoted from the embedded base. The entity adds one getter per field:

```go
func (permission *Permission) RoleID() role.ID { return permission.roleID }
func (permission *Permission) Scope() Scope     { return permission.scope }
```

There is no query that also mutates; event access is the base's CQS-split
`PullEvents()` (copy) / `DrainEvents()` (clear).

### Events — `permission/events.go`

Only the entity records events, past-tense, named via a constant, carrying **ids and
values, never a whole entity**. A permission has exactly one fact: it was granted.

```go
const permissionGrantedEventName = "permission.granted"

// PermissionGrantedEvent is recorded when a scope is granted to a role. It carries the
// permission's id, the owning role's id, and the granted scope value.
type PermissionGrantedEvent struct {
	permissionID ID
	roleID       role.ID
	scope        Scope
}

func NewPermissionGrantedEvent(permissionID ID, roleID role.ID, scope Scope) PermissionGrantedEvent {
	return PermissionGrantedEvent{permissionID: permissionID, roleID: roleID, scope: scope}
}

func (e PermissionGrantedEvent) PermissionID() ID    { return e.permissionID }
func (e PermissionGrantedEvent) RoleID() role.ID     { return e.roleID }
func (e PermissionGrantedEvent) Scope() Scope        { return e.scope }
func (e PermissionGrantedEvent) EventName() string   { return permissionGrantedEventName }
```

(Deletion is a persistence outcome driven by the application, not a state transition of
a still-living entity, so there is no `PermissionRevoked` domain event in the MVP.)

### Errors — `permission/errors.go`

Business/invariant violations only, each a factory returning `*domain.DomainError`
(lookup/persistence outcomes — not found, already exists — belong at the
repository/application boundary, not here). Go-style names; the package carries the
concept.

```go
func NewEmptyIDError() *domain.DomainError {
	return &domain.DomainError{Message: "permission id must not be empty"}
}

func NewEmptyRoleIDError() *domain.DomainError {
	return &domain.DomainError{Message: "permission role id must not be empty"}
}

func NewEmptyScopeError() *domain.DomainError {
	return &domain.DomainError{Message: "permission scope must not be empty"}
}

func NewScopeTooLongError(length int) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("permission scope must be at most %d characters, got %d", maxScopeLength, length),
	}
}

func NewMalformedScopeError(value string) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("permission scope %q is malformed: expected colon-separated lowercase alphanumeric segments (e.g. \"users:read\")", value),
	}
}
```

`NewEmptyIDError` / `NewIDTooLongError` for the `ID` VO follow `account.ID` verbatim.

### Repository port — `permission/repository.go`

Declared by the domain, implemented in `infrastructure`; persistence-oriented method
names, `context.Context` first. **There is no `Update`** — permissions are immutable
(piece **c**), so there is nothing to update. Two role-scoped reads support the RBAC
use cases: listing a role's permissions, and enforcing per-role scope uniqueness at the
application boundary.

```go
// package permission — the package carries the concept, so no stutter
type Repository interface {
	Create(ctx context.Context, permission *Permission) (*Permission, error)
	Get(ctx context.Context, id ID) (*Permission, error)
	Exists(ctx context.Context, id ID) (bool, error)
	Delete(ctx context.Context, id ID) error
	ListByRole(ctx context.Context, roleID role.ID) ([]*Permission, error)
	ExistsByScopeInRole(ctx context.Context, roleID role.ID, scope Scope) (bool, error)
}
```

- **`ListByRole`** — all permissions owned by a role (the 1—* read; drives token/claim
  assembly and role detail views).
- **`ExistsByScopeInRole`** — the per-role uniqueness check (a role must not hold the
  same scope twice); the application consults it before `Grant`. "Already exists" is a
  boundary outcome, not a domain error, which is why this is a repository query rather
  than an entity invariant.
