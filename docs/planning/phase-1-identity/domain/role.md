# Identity · Domain · `role`

> **Status: complete — pieces (a) value objects, (b) entity + constructors, (c) commands, and (d) queries, events, errors & repository port all done.**
> Full deep-dive for the `role` entity. Part of the Identity domain layer; see the
> [overview](../domain-layer.md), the rules in
> [`domain-layer.md`](../../../architecture/domain-layer.md), and the sibling
> [`account`](./account.md) deep-dive (the pattern this document mirrors).

## What `role` is

An **org-scoped RBAC role**: a named grouping of permissions within a single
organization. It carries a required `organizationID`, a human-readable `Name`, and an
optional `Description`. A role is **always scoped to one organization** — there are no
global roles in the MVP. It does **not** hold a list of permissions: the
[`permission`](./permission.md) entity owns the foreign key (`roleID`), giving
`Role` 1—* `Permission` with no permission-id list on `Role` (see the
[design decisions](../domain-layer.md#design-decisions)). Package:
`internal/identity/domain/role`.

## Build order (bottom-up)

- **(a) Value objects** — `ID`, `Name`, `Description` (+ typed cross-entity ref `organization.ID`) ✅
- **(b) Entity struct + `Create` + `Reconstitute`** ✅
- **(c) Commands** — `Rename`, `UpdateDescription` ✅
- **(d) Queries, events, errors, repository port** ✅

---

## (a) Value objects

One file per value object in the `role` package. Each is immutable (unexported
field), self-validating in its constructor, with `Value()` / `Equal(other)` /
`IsZero()`. Business/invariant failures return `*domain.DomainError` via a factory in
`role/errors.go` (piece **d**).

### `ID` — `id.go`

Same shape as [`account.ID`](./account.md#id--idgo-reuse-verbatim-from-the-example).

- **Wraps:** `string` (an opaque identifier — UUID/ULID, assigned by the application, not generated in the domain).
- **Constructor:** `NewID(raw string) (ID, error)`.
- **Rules:** trim surrounding whitespace; reject empty (`NewEmptyIDError`); reject
  length `> maxIDLength` where `maxIDLength = 64` (`NewIDTooLongError(length)`).
- **Methods:** `Value() string`, `Equal(other ID) bool`, `IsZero() bool`.

### `Name` — `name.go`

The role's human-readable name (e.g. `"Administrator"`, `"Billing Manager"`). A
required VO with the standard "zero value is invalid" contract.

- **Wraps:** `string`.
- **Constructor:** `NewName(raw string) (Name, error)`.
- **Rules:** trim surrounding whitespace; reject empty (`NewEmptyNameError`); reject
  length `> maxNameLength` where `maxNameLength = 100` (`NewNameTooLongError(length)`).
- **Methods:** `Value() string`, `Equal(other Name) bool`, `IsZero() bool`.

```go
// role/name.go
const maxNameLength = 100

type Name struct {
	value string // unexported → immutable, only set via NewName
}

func NewName(raw string) (Name, error) {
	trimmed := strings.TrimSpace(raw)
	nameIsMissing := trimmed == ""
	if nameIsMissing {
		return Name{}, NewEmptyNameError()
	}
	nameIsTooLong := utf8.RuneCountInString(trimmed) > maxNameLength
	if nameIsTooLong {
		return Name{}, NewNameTooLongError(utf8.RuneCountInString(trimmed))
	}
	return Name{value: trimmed}, nil
}

func (name Name) Value() string         { return name.value }
func (name Name) Equal(other Name) bool { return name.value == other.value }
func (name Name) IsZero() bool          { return name == Name{} }
```

### `Description` — `description.go` (optional VO)

A free-text description of what the role is for. **This VO is optional**, which makes it
a deliberate deviation from the standard "zero value is invalid" rule that every other
VO in this context follows.

- **Wraps:** `string`.
- **Constructor:** `NewDescription(raw string) (Description, error)`.
- **Rules:** trim surrounding whitespace; **an empty (or whitespace-only) value is
  valid** and yields the zero-value `Description` — it is *not* an error. Bound the max
  only: reject length `> maxDescriptionLength` where `maxDescriptionLength = 500`
  (`NewDescriptionTooLongError(length)`).
- **Methods:** `Value() string`, `Equal(other Description) bool`, `IsZero() bool`.

**How this deviates from "zero value invalid" — and why it is safe:**

- For every *required* VO (`ID`, `Name`, `organization.ID`), the entity constructor
  guards `IsZero()` and Fail-Fasts. **`Description` is exempt from that guard** — the
  entity accepts a zero-value `Description` as a legitimate "no description" state.
- The constructor therefore **does not reject empty**; it only enforces the upper
  bound. An empty string trims to `""` and is returned as `Description{}` with no error.
- Because empty is allowed, `IsZero()` here answers the business question "was a
  description provided?" — it is *informational*, not an invalidity signal. Contrast
  `Name.IsZero()`, which the entity treats as an error.
- The self-validating + immutable + `Value()`/`Equal()`/`IsZero()` contract is otherwise
  unchanged; the *only* relaxation is that emptiness is a valid value, so a still-bounded
  max length is the sole invariant.

```go
// role/description.go
const maxDescriptionLength = 500

type Description struct {
	value string // unexported → immutable, only set via NewDescription
}

// NewDescription accepts an empty value — an absent description is valid.
// It enforces only the upper bound; empty trims to the zero-value Description.
func NewDescription(raw string) (Description, error) {
	trimmed := strings.TrimSpace(raw)
	descriptionIsTooLong := utf8.RuneCountInString(trimmed) > maxDescriptionLength
	if descriptionIsTooLong {
		return Description{}, NewDescriptionTooLongError(utf8.RuneCountInString(trimmed))
	}
	return Description{value: trimmed}, nil // trimmed may be "" — that is a valid Description
}

func (description Description) Value() string                { return description.value }
func (description Description) Equal(other Description) bool { return description.value == other.value }
func (description Description) IsZero() bool                 { return description == Description{} }
```

### Cross-entity reference — `organization.ID`

A role is **org-scoped**, so it holds a **required** `organization.ID` — the *typed ID
value object imported from the [`organization`](./organization.md) package*, never an
embedded `Organization` entity (per the
[reference-by-ID rule](../../../architecture/domain-layer.md#entities)). It is not
declared in the `role` package; it is imported and stored as a field on the entity
(piece **b**), and the constructor guards it as required.

---

## (b) Entity & constructors

File: `role/role.go`.

### Struct

The entity embeds the base `domain.Entity[ID]` (id + recorded events) and holds its
state as value objects plus the required typed `organization.ID`. It references the
organization by ID only, never by embedding.

```go
type Role struct {
	domain.Entity[ID]
	organizationID organization.ID
	name           Name
	description    Description
}
```

State changes only through the methods below; callers read via getters (piece **d**),
never by assigning fields.

### `Create` — the business constructor

Builds a **new** role from already-valid value objects and the required typed
`organization.ID`. The required fields are guard-claused (Fail Fast): `id`,
`organizationID`, and `name`. **`description` is not guarded** — an empty
`Description` is a valid "no description" state (see piece **a**), so it is stored as-is.
On success the role records `RoleCreatedEvent` (payload defined in piece **d**).

```go
type CreateParams struct {
	ID             ID
	OrganizationID organization.ID
	Name           Name
	Description    Description
}

func Create(params CreateParams) (*Role, error) {
	idIsMissing := params.ID.IsZero()
	if idIsMissing {
		return nil, NewEmptyIDError()
	}
	organizationIDIsMissing := params.OrganizationID.IsZero()
	if organizationIDIsMissing {
		return nil, NewEmptyOrganizationIDError()
	}
	nameIsMissing := params.Name.IsZero()
	if nameIsMissing {
		return nil, NewEmptyNameError()
	}
	role := &Role{
		Entity:         domain.NewEntity(params.ID),
		organizationID: params.OrganizationID,
		name:           params.Name,
		description:    params.Description, // may be the zero-value Description — that is valid
	}
	role.Record(NewRoleCreatedEvent(role.ID(), role.organizationID, role.name))
	return role, nil
}
```

### `Reconstitute` — rebuild from persistence

The counterpart used only by the repository adapter. It takes the **full persisted
state** and **just loads** it into a fresh entity: no validation, no event, no policy,
no error (see the
[Reconstitution rule](../../../architecture/domain-layer.md#reconstitution-rebuilding-from-persistence)).
Here the reconstitute params happen to match the create params one-for-one (a role has
no derived lifecycle state), but it remains its own struct per the rule.

```go
type ReconstituteParams struct {
	ID             ID
	OrganizationID organization.ID
	Name           Name
	Description    Description
}

func Reconstitute(params ReconstituteParams) *Role {
	return &Role{
		Entity:         domain.NewEntity(params.ID),
		organizationID: params.OrganizationID,
		name:           params.Name,
		description:    params.Description,
	}
}
```

## (c) Commands

Each command mutates state and returns **only `error`** (CQS — a command never returns
data). Each takes already-valid value objects, applies its change, and records exactly
one past-tense event (piece **d**). The `organizationID` is **immutable after
creation** — a role never moves between organizations, so no command changes it.

### `Rename` — change the role's name

Replaces the name with a new already-valid `Name`, then records `RoleRenamedEvent`.

```go
func (role *Role) Rename(name Name) error {
	role.name = name
	role.Record(NewRoleRenamedEvent(role.ID(), role.name))
	return nil
}
```

### `UpdateDescription` — change (or clear) the description

Replaces the description with a new `Description` — which **may be the zero value**, the
valid way to *clear* a description — then records `RoleDescriptionUpdatedEvent`.

```go
func (role *Role) UpdateDescription(description Description) error {
	role.description = description // may be the zero-value Description — clearing is valid
	role.Record(NewRoleDescriptionUpdatedEvent(role.ID(), role.description))
	return nil
}
```

## (d) Queries, events, errors, repository

### Queries (getters)

Read-only accessors returning value objects (CQS — no mutation). Callers never touch
fields directly.

```go
func (role *Role) OrganizationID() organization.ID { return role.organizationID }
func (role *Role) Name() Name                       { return role.name }
func (role *Role) Description() Description          { return role.description }
```

(`ID()` is promoted from the embedded base `Entity[ID]`.)

Because `OrganizationID()` returns `organization.ID`, `Role` satisfies
`domain.OrganizationScoped`, so it plugs directly into the shared
`domain.EnsureSameOrganization` guard used by cross-entity services (e.g.
`AssignRoleToMembership`). Its `organizationID` is immutable (set only at `Create`), so
a role never migrates tenants — see
[Multi-tenancy](../domain-layer.md#multi-tenancy--the-organization-boundary).

### Events — `role/events.go`

Past-tense facts, each with its name in a **named constant** returned by `EventName()`
(no inline literals). Each event carries the role's **id** plus the relevant changed
value(s) — ids and value objects only, never the whole entity. All implement the
context's `Event` marker from `domain/shared.go`.

| Event | Recorded by | Payload |
|---|---|---|
| `RoleCreatedEvent` | `Create` | role `ID`, `organization.ID`, `Name` |
| `RoleRenamedEvent` | `Rename` | role `ID`, new `Name` |
| `RoleDescriptionUpdatedEvent` | `UpdateDescription` | role `ID`, new `Description` |

```go
// role/events.go
package role

const (
	roleCreatedEventName            = "role.created"
	roleRenamedEventName            = "role.renamed"
	roleDescriptionUpdatedEventName = "role.description_updated"
)

// RoleCreatedEvent is recorded when a role is created. It carries the role's id,
// its organization's id, and its name.
type RoleCreatedEvent struct {
	roleID         ID
	organizationID organization.ID
	name           Name
}

func NewRoleCreatedEvent(roleID ID, organizationID organization.ID, name Name) RoleCreatedEvent {
	return RoleCreatedEvent{roleID: roleID, organizationID: organizationID, name: name}
}

func (e RoleCreatedEvent) RoleID() ID                       { return e.roleID }
func (e RoleCreatedEvent) OrganizationID() organization.ID  { return e.organizationID }
func (e RoleCreatedEvent) Name() Name                       { return e.name }
func (e RoleCreatedEvent) EventName() string                { return roleCreatedEventName }

// RoleRenamedEvent is recorded when a role is renamed. It carries the role's id and new name.
type RoleRenamedEvent struct {
	roleID ID
	name   Name
}

func NewRoleRenamedEvent(roleID ID, name Name) RoleRenamedEvent {
	return RoleRenamedEvent{roleID: roleID, name: name}
}

func (e RoleRenamedEvent) RoleID() ID          { return e.roleID }
func (e RoleRenamedEvent) Name() Name          { return e.name }
func (e RoleRenamedEvent) EventName() string   { return roleRenamedEventName }

// RoleDescriptionUpdatedEvent is recorded when a role's description changes. It carries the
// role's id and the new description (which may be the zero-value Description).
type RoleDescriptionUpdatedEvent struct {
	roleID      ID
	description Description
}

func NewRoleDescriptionUpdatedEvent(roleID ID, description Description) RoleDescriptionUpdatedEvent {
	return RoleDescriptionUpdatedEvent{roleID: roleID, description: description}
}

func (e RoleDescriptionUpdatedEvent) RoleID() ID              { return e.roleID }
func (e RoleDescriptionUpdatedEvent) Description() Description { return e.description }
func (e RoleDescriptionUpdatedEvent) EventName() string       { return roleDescriptionUpdatedEventName }
```

### Errors — `role/errors.go`

Business/invariant violations **only** — VO and constructor guards. Lookup/persistence
outcomes ("not found", "already exists", timeouts) are **not** domain errors; they live
at the repository/application boundary. Each factory returns `*domain.DomainError` with
its own self-describing message; names are Go-style (`New<Concept>Error`, the package
carries the `role` concept).

| Factory | Raised by | Why |
|---|---|---|
| `NewEmptyIDError()` | `ID` ctor, `Create` guard | id trimmed to empty / missing |
| `NewIDTooLongError(length)` | `ID` ctor | id longer than `maxIDLength` |
| `NewEmptyNameError()` | `Name` ctor, `Create` guard | name trimmed to empty / missing |
| `NewNameTooLongError(length)` | `Name` ctor | name longer than `maxNameLength` |
| `NewDescriptionTooLongError(length)` | `Description` ctor | description longer than `maxDescriptionLength` (empty is *not* an error) |
| `NewEmptyOrganizationIDError()` | `Create` guard | required `organizationID` missing (org-scoped) |

There is deliberately **no** "empty description" error — an absent description is a valid
value, so the only description invariant is its max length.

```go
// role/errors.go
package role

func NewEmptyNameError() *domain.DomainError {
	return &domain.DomainError{Message: "role name must not be empty"}
}

func NewNameTooLongError(length int) *domain.DomainError {
	return &domain.DomainError{Message: fmt.Sprintf("role name must be at most %d characters, got %d", maxNameLength, length)}
}

func NewDescriptionTooLongError(length int) *domain.DomainError {
	return &domain.DomainError{Message: fmt.Sprintf("role description must be at most %d characters, got %d", maxDescriptionLength, length)}
}

func NewEmptyOrganizationIDError() *domain.DomainError {
	return &domain.DomainError{Message: "role organization id must not be empty"}
}

// NewEmptyIDError / NewIDTooLongError mirror the account package's id errors.
```

### Repository port — `role/repository.go`

Declared by the domain, implemented in `infrastructure`. Persistence-oriented method
names (not domain verbs), each taking a `context.Context`. Two lookups are org-scoped —
`ListByOrganization` (enumerate a tenant's roles) and `ExistsByNameInOrganization`
(support per-organization name-uniqueness checks at the application boundary).

```go
// package role — the package carries the concept, so no stutter
type Repository interface {
	Create(ctx context.Context, role *Role) (*Role, error)
	Get(ctx context.Context, id ID) (*Role, error)
	Exists(ctx context.Context, id ID) (bool, error)
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id ID) error
	ListByOrganization(ctx context.Context, organizationID organization.ID) ([]*Role, error)
	ExistsByNameInOrganization(ctx context.Context, organizationID organization.ID, name Name) (bool, error)
}
```
