# Identity · Domain · `organization`

> **Status: done — all pieces (a) value objects, (b) entity + constructors, (c) commands, (d) queries/events/errors/repository specified.**
> Full deep-dive for the `organization` entity. Part of the Identity domain layer; see the
> [overview](../domain-layer.md), the rules in
> [`domain-layer.md`](../../../architecture/domain-layer.md), and the worked
> [`examples/identity/`](../../../architecture/examples/identity/) (which models an
> [`account`](./account.md), whose `ID` value object and lifecycle-status pattern this
> entity reuses).

## What `organization` is

The **tenant**: the isolation boundary every other Identity concept hangs off. It has a
human-facing `Name`, a stable, URL-safe `Slug` (used in URLs, sub-domains, and lookups),
and a lifecycle `Status`. Accounts join it through [`membership`](./membership.md), and
[`role`](./role.md)s are scoped to it — but the organization itself references **no other
entity**; those links are owned by the dependent side. It is the subject of creation,
renaming, re-slugging, and suspension/activation. Package:
`internal/identity/domain/organization`.

## Build order (bottom-up)

- **(a) Value objects** — `ID`, `Name`, `Slug`, `Status` ✅
- **(b) Entity struct + `Create` + `Reconstitute`** ✅
- **(c) Commands** — `Rename`, `ChangeSlug`, `Suspend`, `Activate` ✅
- **(d) Queries, events, errors, repository port** ✅

---

## (a) Value objects

One file per value object in the `organization` package. Each is immutable (unexported
field), self-validating in its constructor, with `Value()` / `Equal(other)` /
`IsZero()`, and a zero value that is invalid. Business/invariant failures return
`*domain.DomainError` via a factory in `organization/errors.go` (piece **d**).

### `ID` — `id.go` (reuse the account rules verbatim)

Identical in shape to
[`examples/identity/domain/account/id.go`](../../../architecture/examples/identity/domain/account/id.go),
re-declared in the `organization` package (VOs never cross packages).

- **Wraps:** `string` (an opaque identifier — UUID/ULID, assigned by the application, not generated in the domain).
- **Constructor:** `NewID(raw string) (ID, error)`.
- **Rules:** trim surrounding whitespace; reject empty (`NewEmptyIDError`); reject
  length `> maxIDLength` where `maxIDLength = 64` (`NewIDTooLongError(length)`).
- **Methods:** `Value() string`, `Equal(other ID) bool`, `IsZero() bool`.

### `Name` — `name.go`

The organization's human-facing display name. Free-form text; the only rules are
presence and an upper bound.

- **Wraps:** `string`.
- **Constructor:** `NewName(raw string) (Name, error)`.
- **Rules:** trim surrounding whitespace; reject empty after trimming
  (`NewEmptyNameError`); reject length `> maxNameLength` where `maxNameLength = 200`
  (counted in **runes**, not bytes, so multi-byte names are measured fairly)
  (`NewNameTooLongError(length)`). No case normalization — a display name is shown as
  the user typed it.
- **Methods:** `Value() string`, `Equal(other Name) bool`, `IsZero() bool`.

```go
// organization/name.go
const maxNameLength = 200

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

### `Slug` — `slug.go`

The organization's stable, URL-safe identifier — the string that appears in URLs,
sub-domains, and human-readable lookups (`GetBySlug`). It is normalized and
format-restricted so it is safe to place in a DNS label and stable for equality and
uniqueness.

- **Wraps:** `string`.
- **Constructor:** `NewSlug(raw string) (Slug, error)`.
- **Rules (in order):**
  1. **Trim** surrounding whitespace.
  2. **Normalize to lower case** (so `Acme` and `acme` are the same slug — equality and
     uniqueness stay stable).
  3. Reject empty after trimming (`NewEmptySlugError`).
  4. Reject length `> maxSlugLength` where `maxSlugLength = 63` (the maximum length of a
     single DNS label, so a slug can always be used as a sub-domain)
     (`NewSlugTooLongError(length)`).
  5. **Format:** the normalized value must be a **DNS-label-style token** — lowercase
     ASCII alphanumerics (`a–z`, `0–9`) separated by **single** hyphens, and it must
     **not** start or end with a hyphen and must **not** contain consecutive hyphens.
     Anything else (spaces, underscores, uppercase leftovers, dots, other punctuation)
     is rejected (`NewMalformedSlugError(value)`). Concretely it must match
     `^[a-z0-9]+(-[a-z0-9]+)*$`.
- **Methods:** `Value() string`, `Equal(other Slug) bool`, `IsZero() bool`.
- **Note:** normalization is limited to trimming + lower-casing; the constructor
  **rejects** anything that isn't already a clean slug rather than silently rewriting
  spaces/underscores into hyphens. Producing a suggested slug from a name is an
  application/UI concern, not the VO's job — the domain only accepts or refuses.

```go
// organization/slug.go
const maxSlugLength = 63

// slugPattern: lowercase alphanumeric labels joined by single hyphens,
// no leading/trailing/consecutive hyphens. (Compiled once at package init.)
var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

type Slug struct {
	value string
}

func NewSlug(raw string) (Slug, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	slugIsMissing := normalized == ""
	if slugIsMissing {
		return Slug{}, NewEmptySlugError()
	}
	slugIsTooLong := len(normalized) > maxSlugLength
	if slugIsTooLong {
		return Slug{}, NewSlugTooLongError(len(normalized))
	}
	slugIsMalformed := !slugPattern.MatchString(normalized)
	if slugIsMalformed {
		return Slug{}, NewMalformedSlugError(normalized)
	}
	return Slug{value: normalized}, nil
}

func (slug Slug) Value() string          { return slug.value }
func (slug Slug) Equal(other Slug) bool  { return slug.value == other.value }
func (slug Slug) IsZero() bool           { return slug == Slug{} }
```

> `regexp` is standard library, so it does not violate the "no third-party imports in
> `domain`" rule. Since the slug is pure ASCII, `len` (bytes) equals rune count for the
> length bound; the bound is applied before the format check for a clearer error.

### `Status` — `status.go` (enum, set by the entity)

The organization's place in its lifecycle. Like [`account.Status`](./account.md#status--statusgo-enum-mirrors-the-example--extends),
it is an **enumerated type set by the entity itself** (at creation, or by a transition
command); it is never parsed from raw external input, so it has **no `New…`
constructor**. Persistence ↔ enum mapping is an infrastructure concern.

- **Type:** `type Status int`, values via `iota`:
  - `Active` — the organization is usable. The **starting** status at creation, and
    where a `Suspended` organization lands after `Activate`.
  - `Suspended` — temporarily blocked by an operator; can be reactivated. There is no
    terminal/deleted status — removal is a repository `Delete`, not a lifecycle state.
- **Suggested helper:** `String() string` for logging/mapping (no behavior change).
- Legal transitions are specified with the commands in piece **(c)**, not here.

```go
// organization/status.go
type Status int

const (
	// Active — the organization is usable. The starting status at creation.
	Active Status = iota
	// Suspended — temporarily blocked by an operator; can be reactivated.
	Suspended
)
```

---

## (b) Entity & constructors

File: `organization/organization.go`.

### Struct

The entity embeds the base `domain.Entity[ID]` (id + recorded events) and holds its
state as value objects only. It references **no other entity** — memberships and roles
point *at* the organization from their own packages, so nothing is held here.

```go
type Organization struct {
	domain.Entity[ID]
	name   Name
	slug   Slug
	status Status
}
```

State changes only through the methods below; callers read via getters (piece **d**),
never by assigning fields.

### `Create` — the business constructor

Builds a **new** organization from already-valid value objects. Each field is
guard-claused (Fail Fast); the starting `status` is always `Active` (there is no policy
to vary it — unlike `account`, an organization has no verification step); the
organization records `OrganizationCreatedEvent` (payload defined in piece **d**).

```go
type CreateParams struct {
	ID   ID
	Name Name
	Slug Slug
}

func Create(params CreateParams) (*Organization, error) {
	idIsMissing := params.ID.IsZero()
	if idIsMissing {
		return nil, NewEmptyIDError()
	}
	nameIsMissing := params.Name.IsZero()
	if nameIsMissing {
		return nil, NewEmptyNameError()
	}
	slugIsMissing := params.Slug.IsZero()
	if slugIsMissing {
		return nil, NewEmptySlugError()
	}
	organization := &Organization{
		Entity: domain.NewEntity(params.ID),
		name:   params.Name,
		slug:   params.Slug,
		status: Active, // organizations always start active
	}
	organization.Record(NewOrganizationCreatedEvent(organization.ID(), organization.slug))
	return organization, nil
}
```

### `Reconstitute` — rebuild from persistence

The counterpart used only by the repository adapter. It takes the **full persisted
state** — a superset of `CreateParams` that also carries `Status` — and **just loads**
it into a fresh entity: no validation, no event, no policy, no error (see the
[Reconstitution rule](../../../architecture/domain-layer.md#reconstitution-rebuilding-from-persistence)).

```go
type ReconstituteParams struct {
	ID     ID
	Name   Name
	Slug   Slug
	Status Status
}

func Reconstitute(params ReconstituteParams) *Organization {
	return &Organization{
		Entity: domain.NewEntity(params.ID),
		name:   params.Name,
		slug:   params.Slug,
		status: params.Status,
	}
}
```

---

## (c) Commands

Each command is a method on `*Organization`. Following CQS, a command **mutates state
and returns only `error`** (never data). Each takes already-valid value objects (never
primitives — validation happened in the VO constructor), guards the invariant it
protects, mutates in place, and records a past-tense event. Getters are in piece **(d)**.

### `Rename(Name) error` → `OrganizationRenamedEvent`

Changes the display name. The `Name` VO is already valid, so there is nothing to
re-validate and no failure mode — but the method returns `error` for a uniform command
signature and forward compatibility. It sets the field and records the event.

```go
func (organization *Organization) Rename(name Name) error {
	organization.name = name
	organization.Record(NewOrganizationRenamedEvent(organization.ID(), name))
	return nil
}
```

### `ChangeSlug(Slug) error` → `OrganizationSlugChangedEvent`

Changes the URL-safe slug. As with `Rename`, the `Slug` VO is already valid; uniqueness
across organizations is **not** a domain invariant (it needs a store lookup) and is
enforced at the application boundary via `ExistsBySlug` before this call. The method sets
the field and records the event.

```go
func (organization *Organization) ChangeSlug(slug Slug) error {
	organization.slug = slug
	organization.Record(NewOrganizationSlugChangedEvent(organization.ID(), slug))
	return nil
}
```

### `Suspend() error` → `OrganizationSuspendedEvent`

Blocks the organization. **Guard:** rejects the call if it is already `Suspended`
(`NewAlreadySuspendedError`), so the transition is idempotent-safe and no spurious event
is recorded. Otherwise it moves to `Suspended` and records the event.

```go
func (organization *Organization) Suspend() error {
	alreadySuspended := organization.status == Suspended
	if alreadySuspended {
		return NewAlreadySuspendedError()
	}
	organization.status = Suspended
	organization.Record(NewOrganizationSuspendedEvent(organization.ID()))
	return nil
}
```

### `Activate() error` → `OrganizationActivatedEvent`

Reactivates a suspended organization. **Guard:** rejects the call if it is already
`Active` (`NewAlreadyActiveError`). Otherwise it moves to `Active` and records the event.

```go
func (organization *Organization) Activate() error {
	alreadyActive := organization.status == Active
	if alreadyActive {
		return NewAlreadyActiveError()
	}
	organization.status = Active
	organization.Record(NewOrganizationActivatedEvent(organization.ID()))
	return nil
}
```

**Legal transitions:** `Active → Suspended` (`Suspend`), `Suspended → Active`
(`Activate`). A no-op transition (suspending a suspended org, activating an active one)
is rejected by the guard above; there is no terminal status.

---

## (d) Queries, events, errors, repository

### Queries

Read-only getters on `*Organization` (CQS — they return data and never mutate). `ID()`
is promoted from the embedded base `Entity`; the rest are declared here. Callers read
through these, never by touching fields.

```go
func (organization *Organization) Name() Name     { return organization.name }
func (organization *Organization) Slug() Slug     { return organization.slug }
func (organization *Organization) Status() Status { return organization.status }
```

### Events — `organization/events.go`

All past-tense facts recorded by the organization, each carrying **ids/value objects,
never the whole entity**, each with its name in a **named constant** returned by
`EventName()` (no inline literals), each implementing the context's `domain.Event`
marker. Every event carries the organization's `ID`; state-changing ones also carry the
relevant new value.

| Event | Constant | Recorded by | Carries |
|---|---|---|---|
| `OrganizationCreatedEvent` | `organization.created` | `Create` | `ID`, `Slug` |
| `OrganizationRenamedEvent` | `organization.renamed` | `Rename` | `ID`, `Name` |
| `OrganizationSlugChangedEvent` | `organization.slug_changed` | `ChangeSlug` | `ID`, `Slug` |
| `OrganizationSuspendedEvent` | `organization.suspended` | `Suspend` | `ID` |
| `OrganizationActivatedEvent` | `organization.activated` | `Activate` | `ID` |

```go
// organization/events.go
const (
	organizationCreatedEventName     = "organization.created"
	organizationRenamedEventName     = "organization.renamed"
	organizationSlugChangedEventName = "organization.slug_changed"
	organizationSuspendedEventName   = "organization.suspended"
	organizationActivatedEventName   = "organization.activated"
)

// OrganizationCreatedEvent is recorded when an organization is created. It carries the
// organization's id and the slug it was created with, not the organization itself.
type OrganizationCreatedEvent struct {
	organizationID ID
	slug           Slug
}

func NewOrganizationCreatedEvent(organizationID ID, slug Slug) OrganizationCreatedEvent {
	return OrganizationCreatedEvent{organizationID: organizationID, slug: slug}
}

func (event OrganizationCreatedEvent) OrganizationID() ID   { return event.organizationID }
func (event OrganizationCreatedEvent) Slug() Slug           { return event.slug }
func (event OrganizationCreatedEvent) EventName() string    { return organizationCreatedEventName }

// OrganizationRenamedEvent is recorded when the display name changes; carries the new name.
type OrganizationRenamedEvent struct {
	organizationID ID
	name           Name
}

func NewOrganizationRenamedEvent(organizationID ID, name Name) OrganizationRenamedEvent {
	return OrganizationRenamedEvent{organizationID: organizationID, name: name}
}

func (event OrganizationRenamedEvent) OrganizationID() ID { return event.organizationID }
func (event OrganizationRenamedEvent) Name() Name         { return event.name }
func (event OrganizationRenamedEvent) EventName() string  { return organizationRenamedEventName }

// OrganizationSlugChangedEvent is recorded when the slug changes; carries the new slug.
type OrganizationSlugChangedEvent struct {
	organizationID ID
	slug           Slug
}

func NewOrganizationSlugChangedEvent(organizationID ID, slug Slug) OrganizationSlugChangedEvent {
	return OrganizationSlugChangedEvent{organizationID: organizationID, slug: slug}
}

func (event OrganizationSlugChangedEvent) OrganizationID() ID { return event.organizationID }
func (event OrganizationSlugChangedEvent) Slug() Slug         { return event.slug }
func (event OrganizationSlugChangedEvent) EventName() string  { return organizationSlugChangedEventName }

// OrganizationSuspendedEvent is recorded when the organization is suspended.
type OrganizationSuspendedEvent struct {
	organizationID ID
}

func NewOrganizationSuspendedEvent(organizationID ID) OrganizationSuspendedEvent {
	return OrganizationSuspendedEvent{organizationID: organizationID}
}

func (event OrganizationSuspendedEvent) OrganizationID() ID { return event.organizationID }
func (event OrganizationSuspendedEvent) EventName() string  { return organizationSuspendedEventName }

// OrganizationActivatedEvent is recorded when the organization is (re)activated.
type OrganizationActivatedEvent struct {
	organizationID ID
}

func NewOrganizationActivatedEvent(organizationID ID) OrganizationActivatedEvent {
	return OrganizationActivatedEvent{organizationID: organizationID}
}

func (event OrganizationActivatedEvent) OrganizationID() ID { return event.organizationID }
func (event OrganizationActivatedEvent) EventName() string  { return organizationActivatedEventName }
```

### Errors — `organization/errors.go`

**Business-rule / invariant violations only** — one `*domain.DomainError` factory per
failure, self-describing, Go-style names (the package carries the concept, so no
`Organization…` stutter). Lookup/persistence outcomes ("not found", "slug already
taken") are **not** here — they belong at the repository/application boundary.

- **VO errors** — `NewEmptyIDError()`, `NewIDTooLongError(length int)`,
  `NewEmptyNameError()`, `NewNameTooLongError(length int)`, `NewEmptySlugError()`,
  `NewSlugTooLongError(length int)`, `NewMalformedSlugError(value string)`.
- **Transition errors** — `NewAlreadySuspendedError()` (from `Suspend`),
  `NewAlreadyActiveError()` (from `Activate`).

```go
// organization/errors.go
func NewEmptyIDError() *domain.DomainError {
	return &domain.DomainError{Message: "organization id must not be empty"}
}

func NewIDTooLongError(length int) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("organization id must not exceed %d characters, got %d", maxIDLength, length),
	}
}

func NewEmptyNameError() *domain.DomainError {
	return &domain.DomainError{Message: "organization name must not be empty"}
}

func NewNameTooLongError(length int) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("organization name must not exceed %d characters, got %d", maxNameLength, length),
	}
}

func NewEmptySlugError() *domain.DomainError {
	return &domain.DomainError{Message: "organization slug must not be empty"}
}

func NewSlugTooLongError(length int) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("organization slug must not exceed %d characters, got %d", maxSlugLength, length),
	}
}

func NewMalformedSlugError(value string) *domain.DomainError {
	return &domain.DomainError{
		Message: fmt.Sprintf("organization slug is malformed: %q (want lowercase alphanumeric labels joined by single hyphens)", value),
	}
}

func NewAlreadySuspendedError() *domain.DomainError {
	return &domain.DomainError{Message: "organization is already suspended"}
}

func NewAlreadyActiveError() *domain.DomainError {
	return &domain.DomainError{Message: "organization is already active"}
}
```

### Repository port — `organization/repository.go`

The domain declares the interface; the concrete implementation lives in infrastructure.
**Persistence-oriented verbs** (not domain verbs). Beyond the standard CRUD set, the
organization adds `GetBySlug` / `ExistsBySlug` because the slug is a natural secondary
key (URL/sub-domain lookups, and uniqueness enforcement before `Create`/`ChangeSlug`).
Every method takes `context.Context`.

```go
// organization/repository.go
package organization

import "context"

// Repository is the organization's repository port. The domain declares the
// interface; the concrete implementation lives in infrastructure.
type Repository interface {
	Create(ctx context.Context, organization *Organization) (*Organization, error)
	Get(ctx context.Context, id ID) (*Organization, error)
	Exists(ctx context.Context, id ID) (bool, error)
	Update(ctx context.Context, organization *Organization) error
	Delete(ctx context.Context, id ID) error
	GetBySlug(ctx context.Context, slug Slug) (*Organization, error)
	ExistsBySlug(ctx context.Context, slug Slug) (bool, error)
}
```
