# Phase 1 · Identity · Domain Layer — Overview

> **Status: in progress.** This is the domain-layer **overview and index**: scope,
> design decisions, package layout, the relationship map, the shared root, and links
> to the **per-entity deep-dive documents** under [`domain/`](./domain/). Each entity
> is planned to full depth in its own doc, built bottom-up (value objects → entity,
> piece by piece). See [`../method.md`](../method.md).

The domain model for the **Identity** bounded context of the Citadel MVP. It follows
[`../../architecture/domain-layer.md`](../../architecture/domain-layer.md) and mirrors
the worked example in
[`../../architecture/examples/identity/`](../../architecture/examples/identity/).

## Scope

Derived from the checked (MVP) Identity features in
[`../../AUTH_FEATURES.md`](../../AUTH_FEATURES.md): password login, registration, email
verification, password reset, RBAC (roles + permissions in tokens), JWT access + refresh
tokens with rotation/reuse detection + configurable lifetimes, and
organizations/tenants + isolation + membership.

The **Audit** context and the application/infrastructure layers are out of scope here;
the closing [Out of scope](#out-of-scope) section lists MVP concerns handled outside the
domain.

## Design decisions

1. **Single `Account`** entity holds email, password hash, and status — password is the
   first credential type; no separate `Credential` entity for MVP.
2. **One generic `VerificationToken`** entity covers email-verification and
   password-reset, told apart by a `Purpose` discriminator.
3. **RBAC** uses two entities, `Role` and `Permission`; `Permission` owns the foreign
   key (`roleID`), giving `Role` 1—* `Permission` with no permission-id list on `Role`.
4. **Roles are org-scoped only** — every `Role` holds a required `organizationID`.
5. **No pivot/join entities.** A bare link (two foreign keys, no attributes, no
   lifecycle) is a persistence detail, not a domain entity. The member↔role
   many-to-many lives as a **set of `role.ID` on `Membership`**, mapped to a pivot
   table by the repository.
6. **Signing keys are infrastructure** — no `SigningKey` domain entity in the MVP.

**Guiding rule for references:** the dependent side owns the foreign key when it is
itself a managed entity (`Permission.roleID`); a bare link with nothing to own it
collapses into the owning entity's state (`Membership` carries `[]role.ID`).

## Package layout

```
internal/identity/domain/                 # package domain (root)
├── shared.go                             #   DomainError + Event marker + base Entity (from the example)
├── assign_role_to_membership.go          #   cross-entity domain service
├── account/                              #   the authenticatable identity
├── organization/                         #   a tenant
├── membership/                           #   account <-> org link (lifecycle + role set)
├── role/                                 #   org-scoped RBAC role
├── permission/                           #   scope granted to a role (owns roleID)
├── verificationtoken/                    #   one-shot verify / reset token
└── session/                              #   an authenticated session (Kind: Refresh | ServerSide)
```

Each entity package holds: the entity file, its value objects (one file each),
`events.go`, `errors.go`, `repository.go`, and any domain policy.

## Relationship map

```
Account 1 ─ * Membership * ─ 1 Organization    (Membership holds accountID + organizationID)
Membership * ─ * Role                            (Membership holds []role.ID; pivot table in persistence)
Organization 1 ─ * Role                          (Role holds organizationID)
Role 1 ─ * Permission                            (Permission holds roleID)
Account 1 ─ * VerificationToken                  (VerificationToken holds accountID)
Account 1 ─ * Session                            (Session holds accountID; Kind = Refresh | ServerSide; rotation mutates in place)
```

Every cross-entity reference is the target package's **typed ID value object**
(`account.ID`, `organization.ID`, `role.ID`), never an embedded entity.

## Conventions (used across the per-entity documents)

- **Value object (VO)** — self-validating, immutable, one file each in the entity
  package; constructor `New<VO>(raw) (<VO>, error)`, with `Value()` / `Equal(other)` /
  `IsZero()`. Zero value is invalid.
- **Command** — mutates state, returns only `error` (CQS).
- **Query** — returns data, never mutates (CQS).
- Constructors use business-language verbs.
- `now time.Time` is passed in by the application (the domain does no I/O; stdlib
  `time` is allowed, third-party packages are not).

---

## Domain root

Package `domain`, at `internal/identity/domain/`. Holds only the shared building
blocks every entity depends on, plus cross-entity domain services.

### `shared.go` — reused verbatim from the example

Copied unchanged from
[`../../architecture/examples/identity/domain/shared.go`](../../architecture/examples/identity/domain/shared.go).
It defines the three primitives the whole context is built on:

- **`DomainError`** — the single error type for every domain failure (message +
  optional wrapped cause). Told apart by the factory that built it, never a
  per-failure type.
- **`Event`** — the marker interface for recorded domain facts (`EventName() string`).
- **`Entity[ID comparable]`** — the generic base every entity embeds. Carries the
  entity's typed `id` and its recorded `events`, and promotes `ID()`, `Record(event)`,
  `PullEvents()` (query — returns a copy), and `DrainEvents()` (command — clears).
  Entities never re-implement this plumbing; `NewEntity(id)` builds it.

No changes are needed — the example's `shared.go` already fits this context.

### Cross-entity domain services

One file per service at the `domain` root. This context has a single one,
`assign_role_to_membership.go`, specified in [Domain service & policies](#domain-service--policies)
(Phase 9). It is noted here only so the root's contents are complete; there is no
other root-level logic.

### Tenant boundary — `organization_scope.go`

The `domain` root also holds the **tenant-boundary guard** — the domain's contribution
to multi-tenant isolation (full rationale in
[Multi-tenancy](#multi-tenancy--the-organization-boundary)). Pure and stateless:

```go
// internal/identity/domain/organization_scope.go
package domain

import "github.com/AymanKastali/citadel/internal/identity/domain/organization"

// OrganizationScoped is implemented by any entity that belongs to one organization.
type OrganizationScoped interface {
	OrganizationID() organization.ID
}

// SameOrganization reports whether two org-scoped entities share an organization.
func SameOrganization(a, b OrganizationScoped) bool {
	return a.OrganizationID().Equal(b.OrganizationID())
}

// EnsureSameOrganization rejects an operation that combines entities from different
// organizations. Every cross-entity domain service calls it.
func EnsureSameOrganization(a, b OrganizationScoped) error {
	crossesOrganizations := !SameOrganization(a, b)
	if crossesOrganizations {
		return NewCrossOrganizationError()
	}
	return nil
}
```

`NewCrossOrganizationError()` is a **root-level** factory (alongside `DomainError` in
`shared.go`) because the tenant boundary is cross-cutting, owned by no single entity:

```go
func NewCrossOrganizationError() *DomainError {
	return &DomainError{Message: "operation not allowed across organizations"}
}
```

`membership` and `role` already satisfy `OrganizationScoped` (both expose
`OrganizationID() organization.ID`) — no struct change is needed.

## Entities

Each entity is documented to full depth in its own file under [`domain/`](./domain/),
built bottom-up (value objects first, then the entity piece by piece). Statuses track
this plan's phases.

| Entity | Document | What it is | Status |
|---|---|---|---|
| `account` | [`domain/account.md`](./domain/account.md) | the authenticatable identity | ✅ done |
| `organization` | [`domain/organization.md`](./domain/organization.md) | a tenant | ✅ done |
| `membership` | [`domain/membership.md`](./domain/membership.md) | account↔org link (lifecycle + role set) | ✅ done |
| `role` | [`domain/role.md`](./domain/role.md) | org-scoped RBAC role | ✅ done |
| `permission` | [`domain/permission.md`](./domain/permission.md) | scope granted to a role (owns `roleID`) | ✅ done |
| `verificationtoken` | [`domain/verificationtoken.md`](./domain/verificationtoken.md) | one-shot verify / reset token | ✅ done |
| `session` | [`domain/session.md`](./domain/session.md) | an authenticated session (`Kind`: Refresh \| ServerSide) | ✅ done |

## Domain service & policies

### Cross-entity domain service — `assign_role_to_membership.go`

Assigning a role to a membership spans **two** entities and belongs to neither: the
membership only holds a `role.ID`, so it cannot itself check that the role belongs to
the member's organization. That check — and only that check — is the job of a domain
service at the `domain` root, mirroring the ordering example's
[`AddProductToOrder`](../../architecture/examples/ordering/domain/add_product_to_order.go).
It is stateless and pure (no persistence, no I/O): the application loads both entities
via their repository ports, calls this service, then persists.

```go
// internal/identity/domain/assign_role_to_membership.go
package domain

import (
	"github.com/AymanKastali/citadel/internal/identity/domain/membership"
	"github.com/AymanKastali/citadel/internal/identity/domain/role"
)

// AssignRoleToMembership assigns a role to a membership, enforcing the one rule that
// spans both: the role must belong to the membership's organization. It coordinates
// the two entities and translates between them; each still enforces its own rules
// (membership.AssignRole rejects a duplicate).
func AssignRoleToMembership(m *membership.Membership, r *role.Role) error {
	if err := EnsureSameOrganization(m, r); err != nil {
		return err
	}
	return m.AssignRole(r.ID())
}
```

- **The cross-org rule uses the shared guard** — `EnsureSameOrganization(m, r)` returns the root-level `NewCrossOrganizationError()` on mismatch (see [Tenant boundary](#tenant-boundary--organization_scopego) and [Multi-tenancy](#multi-tenancy--the-organization-boundary)). Both `membership` and `role` satisfy `domain.OrganizationScoped`, so the service needs no entity-specific cross-org error.
- **Duplicate-guarding stays in `membership.AssignRole`** — the service adds only the cross-entity rule; the single-entity invariant remains the entity's own.
- **No revoke service is needed** — `RevokeRole` requires no cross-entity data, so the application calls `membership.RevokeRole(role.ID)` directly.

### Domain policies

- **`EmailVerificationPolicy`** (in `account/`) is the **only** domain policy in the
  MVP — a pure strategy (`InitialStatus() Status`) with `RequiredEmailVerification` /
  `OptionalEmailVerification`, injected into `account.Register` alongside its params and
  selected from config at the composition root. See
  [`account.md`](./domain/account.md) and the worked
  [`examples/identity/`](../../architecture/examples/identity/).
- **Configurable token/session lifetimes are *not* a policy** — they are **data**: the
  application computes `ExpiresAt` from a configured duration via a Clock port and passes
  it into `verificationtoken.Issue` / `session.Open`. No strategy varies behavior, so no
  policy is warranted.

## Multi-tenancy — the organization boundary

Citadel is multi-tenant: **no operation may cross organizations.** Isolation is enforced
in two complementary layers; the domain owns only what a pure domain can.

**Which entities are scoped to what:**

| Scope | Entities | Carries |
|---|---|---|
| Organization-scoped | `role`, `membership` (and `permission` transitively via its `role`) | an immutable `organization.ID` |
| The boundary itself | `organization` | its own id |
| Account-scoped | `session`, `verificationtoken` | an `account.ID` |
| Global identity | `account` | no org (joins orgs only via `membership`) |

**Layer 1 — access-scoping (ctx + RLS).** The acting organization travels in
`context.Context`; Postgres **Row-Level Security**, keyed on it, decides which rows a
request may read or write. A bare id can never reach another tenant's row, so
**repository ports stay id-based and tenant-agnostic** (no `organizationID` parameters)
— the isolation rides in `ctx`, not in the port signatures. This is infrastructure; the
domain's repository ports are unaffected.

**Layer 2 — logic-scoping (the domain guard).** The one thing a *pure* domain can
enforce: when a single operation combines **two org-bearing entities**, a mismatch is a
domain error, via the `OrganizationScoped` interface + `EnsureSameOrganization` +
`NewCrossOrganizationError` at the root (see
[Tenant boundary](#tenant-boundary--organization_scopego)). Every cross-entity domain
service calls the guard — today `AssignRoleToMembership`; any future service that takes
two org-scoped entities must do the same.

**Invariant — `organizationID` is immutable.** No entity exposes a method to change its
organization. `role.organizationID` and `membership.organizationID` are set at
construction and never reassigned, so an entity can never migrate tenants — which is why
neither has a `ChangeOrganization` command.

**What the domain deliberately does *not* do:** it does not read the acting org from
`ctx` and it does not scope reads — those are Layer 1's job. The domain guards only the
logic it can see (entities already in hand). Together the layers are defense in depth:
RLS stops cross-tenant *access*; the guard stops cross-tenant *combination*.

## Feature coverage

Every checked (MVP) Identity feature in [`AUTH_FEATURES.md`](../../AUTH_FEATURES.md)
maps to a domain concept below, or is explicitly noted as handled outside the domain.

| MVP feature | Where in the domain | Handled outside the domain |
|---|---|---|
| Password login | `account.PasswordHash` (stored hash) + `account.Login` (eligibility guard + `lastLoginAt`) + `session.Open` (on success) | credential verification + orchestration = application **`Authenticate`** use case over the **PasswordHasher** port |
| User registration | `account.Register` (+ `EmailVerificationPolicy`) | — |
| Email verification | `verificationtoken` (`Purpose = EmailVerification`) + `account.VerifyEmail` | secret generation/email delivery = infra |
| Password reset | `verificationtoken` (`Purpose = PasswordReset`) + `account.ChangePassword` | secret generation/email delivery = infra |
| RBAC (roles + permissions) | `role`, `permission` (owns `roleID`), `membership` role set, `AssignRoleToMembership` | — |
| Scopes & permissions in tokens | `permission.Scope` (the granted scopes) | assembling scopes into a JWT = app/infra |
| JWT access tokens | — | minting/signing = app/infra (stateless; not a domain entity) |
| Refresh tokens | `session` (`Kind = Refresh`) | — |
| Refresh-token rotation + reuse detection | `session.Rotate` (mutate-in-place) + `session.Status`/`Revoke` | detecting a replayed secret + `RevokeAllByAccount` orchestration = application |
| Configurable token/session lifetimes | `ExpiresAt` on `verificationtoken` / `session` (data) | duration config + `now` = application **Clock** port |
| Organizations / tenants | `organization` | — |
| Organization membership | `membership` | — |
| Tenant data isolation | `membership` (account↔org scoping) + org-scoped `role` | row-scoping/enforcement in repository queries = infra |

## Out of scope

Deliberately excluded here so nothing is assumed covered:

- **Other layers** — application and infrastructure (this is the domain layer only).
- **The `audit` bounded context** — deferred entirely.
- **Non-entity concerns handled elsewhere:** signing keys / JWKS → infrastructure;
  password hashing/verification → application `PasswordHasher` port; `now`/lifetimes →
  application `Clock` port; JWT access-token minting → app/infra; rate limiting,
  security headers, encryption at rest, structured logging → infra cross-cutting.
- **Non-entity persistence artifacts:** the member↔role **pivot table** is a repository
  mapping detail, not a domain type.
- **Deferred session features** — the `Session` entity + `Kind` reserve room for them,
  but none are built in the MVP: `Kind = ServerSide` (cookie/server-side sessions),
  active-session listing, session-revocation UX, concurrent-session limits.
- **Deep-dive of transitions beyond the MVP** — e.g. account recovery, identity
  linking, MFA factors — all unchecked in `AUTH_FEATURES.md`.
