# Identity · Domain · `membership`

> **Status: done — pieces (a) value objects, (b) entity + constructors, (c) commands, and (d) queries/events/errors/repository all complete.**
> Full deep-dive for the `membership` entity. Part of the Identity domain layer; see the
> [overview](../domain-layer.md), the rules in
> [`domain-layer.md`](../../../architecture/domain-layer.md), and the sibling
> [`account`](./account.md) deep-dive (this document mirrors its structure).

## What `membership` is

The **link between an [`account`](./account.md) and an [`organization`](./organization.md)**,
with its own lifecycle and the account's set of roles *within that organization*. It is a
**real entity, not a bare pivot**: it has an identity, an `Active`/`Suspended` status, and
its own commands and events. This is exactly why it is modelled here and not collapsed
into a join table — a bare two-foreign-key link with no attributes and no lifecycle would
*not* be a domain entity (see the [no-pivot-entities rule](../domain-layer.md#design-decisions)).

It answers "is this account a member of this org, and in what capacity?" — the account
itself holds **no** organization or roles; those live here. The member↔role many-to-many
is held as a **set of [`role.ID`](./role.md) on the membership** (see (a)); the join is a
persistence-only pivot table, never a domain entity. Package:
`internal/identity/domain/membership`.

## Build order (bottom-up)

- **(a) Value objects** — `ID`, `Status`; plus the imported typed refs (`account.ID`, `organization.ID`, `role.ID`) ✅
- **(b) Entity struct + `Create` + `Reconstitute`** ✅
- **(c) Commands** — `Suspend`, `Activate`, `AssignRole`, `RevokeRole` ✅
- **(d) Queries, events, errors, repository port** ✅

---

## (a) Value objects

The `membership` package owns only **two** value objects — `ID` and `Status`. Every
*cross-entity* reference is the **typed ID of the target package**, imported, never
redefined here and never an embedded entity (see the
[reference rule](../domain-layer.md#relationship-map)).

Each owned VO is immutable (unexported field), self-validating in its constructor, with
`Value()` / `Equal(other)` / `IsZero()`, and a zero value that is invalid.
Business/invariant failures return `*domain.DomainError` via a factory in
`membership/errors.go` (piece **d**).

### `ID` — `id.go`

Same shape as [`account.ID`](./account.md#id--idgo-reuse-verbatim-from-the-example).

- **Wraps:** `string` (an opaque identifier — UUID/ULID, assigned by the application, not generated in the domain).
- **Constructor:** `NewID(raw string) (ID, error)`.
- **Rules:** trim surrounding whitespace; reject empty (`NewEmptyIDError`); reject
  length `> maxIDLength` where `maxIDLength = 64` (`NewIDTooLongError(length)`).
- **Methods:** `Value() string`, `Equal(other ID) bool`, `IsZero() bool`.

### `Status` — `status.go` (enum)

The membership's place in its lifecycle. Like [`account.Status`](./account.md#status--statusgo-enum-mirrors-the-example--extends),
it is an **enumerated type set by the entity itself** (at construction, or by a transition
command), never parsed from raw external input — so it has no `New…` constructor.
Persistence ↔ enum mapping is an infrastructure concern.

- **Type:** `type Status int`, values via `iota`:
  - `Active` — the account is a live member of the organization. The **starting** status:
    there is no invitation flow in the MVP, so a new membership is active immediately.
  - `Suspended` — temporarily blocked by an operator; can be reactivated.
- **Suggested helper:** `String() string` for logging/mapping (no behavior change).
- Legal transitions are specified with the commands in piece **(c)**, not here.

### Cross-entity references — imported typed IDs (not defined here)

A membership points at three other entities. Each reference is the **target package's typed
ID value object**, imported — the membership never embeds another entity and never
redefines these IDs:

- **`account.ID`** — the member (imported from the [`account`](./account.md) package).
- **`organization.ID`** — the organization (imported from the [`organization`](./organization.md) package).
- **a set of `role.ID`** — the account's roles *in this organization* (imported from the
  [`role`](./role.md) package).

The member↔role relationship is **many-to-many**. It is modelled as a **set of `role.ID`
held on the membership** (set semantics: no duplicates, order irrelevant), *not* as a
`MembershipRole` pivot entity. A bare link with nothing to own it collapses into the owning
entity's state, so the membership carries `[]role.ID` and the repository maps that set to a
**persistence-only pivot table** — that pivot has no identity, no lifecycle, and no
presence in the domain model.

---

## (b) Entity & constructors

File: `membership/membership.go`.

### Struct

The entity embeds the base `domain.Entity[ID]` (id + recorded events). It holds two typed
cross-entity IDs, its own `Status`, and the role set as `[]role.ID` — a reference by ID,
never an embedded `role.Role`.

```go
type Membership struct {
	domain.Entity[ID]
	accountID      account.ID
	organizationID organization.ID
	status         Status
	roleIDs        []role.ID // set semantics — no duplicates
}
```

State changes only through the methods below; callers read via getters (piece **d**),
never by assigning fields. `roleIDs` is a **set**: the commands in (c) keep it
duplicate-free, and the `Roles()` query (d) hands back a copy so callers cannot mutate it.

### `Create` — the business constructor

Builds a **new** membership from already-valid value objects. Each field is guard-claused
(Fail Fast). There is **no invitation flow in the MVP**, so a new membership starts
`Active` with an **empty role set** (roles are added later via `AssignRole`). No policy is
involved. It records `MemberAddedEvent` (payload defined in piece **d**).

```go
type CreateParams struct {
	ID             ID
	AccountID      account.ID
	OrganizationID organization.ID
}

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
```

### `Reconstitute` — rebuild from persistence

The counterpart used only by the repository adapter. It takes the **full persisted
state** — a superset of `CreateParams` that also carries `Status` and the reconstructed
`RoleIDs` set — and **just loads** it into a fresh entity: no validation, no event, no
policy, no error (see the
[Reconstitution rule](../../../architecture/domain-layer.md#reconstitution-rebuilding-from-persistence)).
The `RoleIDs` slice is rehydrated by the repository from the pivot table.

```go
type ReconstituteParams struct {
	ID             ID
	AccountID      account.ID
	OrganizationID organization.ID
	Status         Status
	RoleIDs        []role.ID
}

func Reconstitute(params ReconstituteParams) *Membership {
	return &Membership{
		Entity:         domain.NewEntity(params.ID),
		accountID:      params.AccountID,
		organizationID: params.OrganizationID,
		status:         params.Status,
		roleIDs:        params.RoleIDs,
	}
}
```

## (c) Commands

File: `membership/membership.go` (alongside the constructors). Every command **mutates
state and returns only `error`** (CQS — never a value), guards its invariant, and records
a past-tense event on success.

### `Suspend` / `Activate` — lifecycle transitions

Guard the **redundant transition** (suspending an already-suspended membership, or
activating an already-active one) so the state change is meaningful and the event is only
fired on a real transition.

```go
func (membership *Membership) Suspend() error {
	if membership.isSuspended() {
		return NewAlreadySuspendedError()
	}
	membership.status = Suspended
	membership.Record(NewMemberSuspendedEvent(membership.ID()))
	return nil
}

func (membership *Membership) Activate() error {
	if membership.isActive() {
		return NewAlreadyActiveError()
	}
	membership.status = Active
	membership.Record(NewMemberActivatedEvent(membership.ID()))
	return nil
}

func (membership *Membership) isSuspended() bool { return membership.status == Suspended }
func (membership *Membership) isActive() bool    { return membership.status == Active }
```

### `AssignRole` / `RevokeRole` — mutating the role set

`AssignRole` adds a `role.ID` to the set, guarding against a **duplicate**; `RevokeRole`
removes one, guarding that it is **present**. Both keep the set duplicate-free and record
the corresponding event.

```go
func (membership *Membership) AssignRole(roleID role.ID) error {
	if membership.HasRole(roleID) {
		return NewRoleAlreadyAssignedError()
	}
	membership.roleIDs = append(membership.roleIDs, roleID)
	membership.Record(NewRoleAssignedEvent(membership.ID(), roleID))
	return nil
}

func (membership *Membership) RevokeRole(roleID role.ID) error {
	if !membership.HasRole(roleID) {
		return NewRoleNotAssignedError()
	}
	membership.roleIDs = removeRoleID(membership.roleIDs, roleID)
	membership.Record(NewRoleRevokedEvent(membership.ID(), roleID))
	return nil
}
```

> **Important — the cross-org check is *not* here.** `AssignRole` guards **only** the
> duplicate. The stronger invariant — *the role must belong to the member's own
> organization* — is **not** enforced by this entity, because a membership holds only the
> role's **id** (`role.ID`), not the `role.Role` entity, and so cannot see the role's
> `organizationID`. That check spans two entities and therefore lives in a dedicated
> **domain service `AssignRoleToMembership`** (documented in Phase 9), which loads the role
> and the membership, enforces same-org via the shared root guard
> `domain.EnsureSameOrganization(m, r)` (→ `domain.NewCrossOrganizationError()`), and then
> calls `membership.AssignRole(role.ID())`. The entity stays honest about what it can know
> on its own; the service owns the cross-entity rule. `Membership` satisfies
> `domain.OrganizationScoped`, and its `organizationID` is set at `Create` and is
> **immutable** — a membership never migrates tenants.

## (d) Queries, events, errors, repository

### Queries (no mutation — CQS)

All read-only; none records an event or changes state. `Roles()` returns a **copy** so
callers cannot mutate the membership's internal set behind its back (same pattern as
[`order.Lines()`](../../../architecture/examples/ordering/domain/order/order.go)).

```go
func (membership *Membership) AccountID() account.ID           { return membership.accountID }
func (membership *Membership) OrganizationID() organization.ID { return membership.organizationID }
func (membership *Membership) Status() Status                  { return membership.status }

// Roles returns a copy so callers cannot mutate the membership's role set.
func (membership *Membership) Roles() []role.ID {
	copied := make([]role.ID, len(membership.roleIDs))
	copy(copied, membership.roleIDs)
	return copied
}

// HasRole reports whether the role is in the set (used by AssignRole/RevokeRole too).
func (membership *Membership) HasRole(roleID role.ID) bool {
	for _, assigned := range membership.roleIDs {
		if assigned.Equal(roleID) {
			return true
		}
	}
	return false
}
```

### Events — `membership/events.go`

Past-tense facts, each named via a **named constant** (no magic strings), immutable, and
carrying **ids only** — never a whole entity. Each implements the context's `Event`
marker. Recorded through the promoted base `Record`.

| Event | Recorded by | Carries |
|---|---|---|
| `MemberAddedEvent` | `Create` | membership id, `account.ID`, `organization.ID` |
| `MemberSuspendedEvent` | `Suspend` | membership id |
| `MemberActivatedEvent` | `Activate` | membership id |
| `RoleAssignedEvent` | `AssignRole` | membership id + `role.ID` |
| `RoleRevokedEvent` | `RevokeRole` | membership id + `role.ID` |

```go
package membership

const (
	memberAddedEventName     = "membership.member_added"
	memberSuspendedEventName = "membership.member_suspended"
	memberActivatedEventName = "membership.member_activated"
	roleAssignedEventName    = "membership.role_assigned"
	roleRevokedEventName     = "membership.role_revoked"
)

// RoleAssignedEvent is recorded when a role is added to the membership's set.
type RoleAssignedEvent struct {
	membershipID ID
	roleID       role.ID
}

func NewRoleAssignedEvent(membershipID ID, roleID role.ID) RoleAssignedEvent {
	return RoleAssignedEvent{membershipID: membershipID, roleID: roleID}
}

func (e RoleAssignedEvent) MembershipID() ID       { return e.membershipID }
func (e RoleAssignedEvent) RoleID() role.ID        { return e.roleID }
func (e RoleAssignedEvent) EventName() string      { return roleAssignedEventName }

// MemberAddedEvent, MemberSuspendedEvent, MemberActivatedEvent, RoleRevokedEvent follow the same shape:
// unexported id fields, a New… factory, id getters, and EventName() → its constant.
```

### Errors — `membership/errors.go`

Business-rule / invariant violations only (never lookup or persistence outcomes), each a
factory returning `*domain.DomainError` with a self-describing message, Go-style
`New<Concept>Error` names (the package carries the concept, so no `Membership` stutter):

- **VO / guard (construction):** `NewEmptyIDError`, `NewIDTooLongError(length)`,
  `NewEmptyAccountIDError`, `NewEmptyOrganizationIDError`.
- **Role-set guards:** `NewRoleAlreadyAssignedError` (assign a role already in the set),
  `NewRoleNotAssignedError` (revoke a role not in the set).
- **Transition guards:** `NewAlreadySuspendedError` (suspend an already-suspended
  membership), `NewAlreadyActiveError` (activate an already-active one).

> The **cross-organization** violation (assigning a role from another org) is **not** a
> membership error: it is caught by the shared root guard
> `domain.EnsureSameOrganization`, which returns `domain.NewCrossOrganizationError()`
> (see [Multi-tenancy](../domain-layer.md#multi-tenancy--the-organization-boundary)).

```go
package membership

func NewRoleAlreadyAssignedError() *domain.DomainError {
	return &domain.DomainError{Message: "role is already assigned to this membership"}
}

func NewRoleNotAssignedError() *domain.DomainError {
	return &domain.DomainError{Message: "role is not assigned to this membership"}
}

func NewAlreadySuspendedError() *domain.DomainError {
	return &domain.DomainError{Message: "membership is already suspended"}
}
```

### Repository port — `membership/repository.go`

An interface declared by the domain, implemented in `infrastructure`, with
**persistence-oriented** method names (not domain verbs). Beyond the standard CRUD it adds
lookups by the two foreign keys — the uniqueness lookup (one membership per
account+organization) and the two list directions. The adapter is responsible for loading
and persisting the `roleIDs` set via the pivot table on `Get`/`Create`/`Update`.

```go
// package membership — the package carries the concept, so no stutter
type Repository interface {
	Create(ctx context.Context, membership *Membership) (*Membership, error)
	Get(ctx context.Context, id ID) (*Membership, error)
	Exists(ctx context.Context, id ID) (bool, error)
	Update(ctx context.Context, membership *Membership) error
	Delete(ctx context.Context, id ID) error
	GetByAccountAndOrganization(ctx context.Context, accountID account.ID, organizationID organization.ID) (*Membership, error)
	ListByAccount(ctx context.Context, accountID account.ID) ([]*Membership, error)
	ListByOrganization(ctx context.Context, organizationID organization.ID) ([]*Membership, error)
}
```
