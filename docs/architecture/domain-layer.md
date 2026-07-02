# Domain Layer

> **Status: draft — under discussion. Nothing here is final yet.**

How the **domain** layer is organized within a bounded context. See [`hexagonal-architecture.md`](./hexagonal-architecture.md) for the cross-cutting rules.

## One directory per entity

One directory per entity (`account`, `user`, `order`, `product`, …), each owning its value objects, errors, and repository port. Three things sit at the `domain` root (package `domain`):

- **`domain/shared.go`** — the context's shared building blocks: `DomainError`, the `Event` marker, and the base **`Entity`** (id + recorded events) that every entity embeds.
- **Domain services** — one file per service (e.g. `domain/add_product_to_order.go`).

```
domain/                          # package domain (root)
├── shared.go                    #   DomainError + Event marker + base Entity
├── add_product_to_order.go      #   a cross-entity domain service
├── product/                     # package product
│   ├── product.go               #   the entity
│   ├── price.go                 #   a value object
│   ├── events.go                #   events recorded by the product
│   ├── errors.go                #   error factories → *domain.DomainError
│   └── repository.go            #   the repository port
├── order/
│   ├── order.go
│   ├── events.go                #   events recorded by the order
│   ├── errors.go
│   └── repository.go
├── account/
│   └── ...
└── user/
    └── ...
```

## Naming

Descriptive names from the business language — name the concept or behavior, not the mechanics. No abbreviations or framework jargon.

```go
// yes — names the business/logic violation (package `order` carries the noun)
func NewAlreadyShippedError() *domain.DomainError

// no — abbreviated, cryptic, or mechanical
func NewErr3() *domain.DomainError
```

## Entities

- Hold the **full logic of one entity** — its rules and state transitions.
- **Mutable**, but only through the entity's own methods; callers read via getters, never assign fields directly.
- **Attributes are mostly value objects**; a field with no rule may stay a primitive.
- **Reference other entities by ID, never by embedding** — hold `order.ID`, not `*Order`. (Embedding the shared base `Entity` is different — that is expected; the rule forbids embedding *another* entity.)
- **Embed the base `Entity`** (`domain.Entity[ID]`) — it carries the id and the recorded domain events and promotes `ID()`, `Record`, `PullEvents`, `DrainEvents`. Entities never re-implement this plumbing.
- **Methods take value objects, not primitives** — validation already happened, so the entity never re-validates.
- **Params grouped in a struct**, not positional arguments.

```go
// yes — a params struct of already-valid value objects
type NewProductParams struct {
    ID    ID
    Name  Name
    Price Price
}

func NewProduct(params NewProductParams) (*Product, error)

// no — scattered params, and raw primitives leak unvalidated input
func NewProduct(id string, name string, price int) (*Product, error)
```

```go
// the entity embeds the base — id and event plumbing come for free
type Product struct {
    domain.Entity[ID]
    name  Name
    price Price
}

// domain/shared.go — the base every entity embeds, id kept strongly typed
type Entity[ID comparable] struct {
    id     ID
    events []Event
}

func (entity *Entity[ID]) ID() ID              { return entity.id }
func (entity *Entity[ID]) Record(event Event)  { entity.events = append(entity.events, event) }
```

## Value objects

- Live in the entity directory (e.g. `product/price.go`).
- **Validate every invariant in the constructor** — both bounds, range, format. `Price` rejects a non-positive **and** a too-large amount. No partial validation, no other way in.
- **Always valid** — an instance can never hold invalid data, so entities built from them are valid too (**Fail Fast**).
- **No identity** — defined only by contents.
- **Immutable** via unexported fields; no setters.
- **Constructor:** `New<VO>(raw) (<VO>, error)`.
- **Raw access via `Value()`**; **equality via `Equal(other) bool`** (never `==`).
- **Zero value is invalid** — the entity constructor treats a zero-value field as an error.

```go
// product/name.go
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

## Errors

- **Business-rule / invariant violations only** — order already shipped, price not positive, name empty. Lookup and persistence outcomes ("not found", "already exists", timeouts) are **not** domain errors; they belong at the repository/application boundary.
- **One type — `DomainError`**, in the `domain` root. Differentiated by the factory and message, not a type per failure.
- **A factory per failure**, living with its concept (`product/errors.go`), returning `*domain.DomainError`.
- **Go-style names** — `New<Concept>Error`; let the package carry the concept (`order.NewAlreadyShippedError`, not `order.NewOrderAlreadyShippedError`).
- **Self-describing** — the factory owns its message; callers pass only context values.
- **No transport or classification** — no HTTP status or code. Mapping happens at the boundary.

```go
// domain/shared.go
package domain

// DomainError is the one type for every domain failure.
type DomainError struct {
    Message string // describes the failure on its own
    Err     error  // optional wrapped cause (nil when there is none)
}

func (e *DomainError) Error() string { return e.Message }
func (e *DomainError) Unwrap() error { return e.Err }
```

```go
// product/errors.go
package product

func NewEmptyNameError() *domain.DomainError {
    return &domain.DomainError{Message: "product name must not be empty"}
}

func NewInvalidPriceError(amount int) *domain.DomainError {
    return &domain.DomainError{Message: fmt.Sprintf("product price must be positive, got %d", amount)}
}
```

## Repository port

- Each entity directory declares its own **repository port** (`product/repository.go`).
- **An interface**, declared by the domain, implemented in `infrastructure`.
- **Persistence-oriented method names** — `Create`, `Get`, `Exists`, `Update`, `Delete` (not domain verbs).

```go
// package product — the package carries the concept, so no stutter
type Repository interface {
    Create(ctx context.Context, product *Product) (*Product, error)
    Get(ctx context.Context, id ID) (*Product, error)
    Exists(ctx context.Context, id ID) (bool, error)
    Update(ctx context.Context, product *Product) error
    Delete(ctx context.Context, id ID) error
}
```

## Domain services (cross-entity logic)

- Logic that **spans more than one entity** and belongs to none.
- Live at the **`domain` root**, one service per file (e.g. `domain/add_product_to_order.go`).
- **Stateless and pure** — no persistence, no I/O. Take entities/value objects as arguments, return the result.

| | Lives in | Does |
|---|---|---|
| **Domain service** | `domain` root | Pure cross-entity business logic. No persistence, no I/O. |
| **Application service** | `application/` | Orchestrates: loads via ports → calls the domain → persists via ports. |

## Domain policies (injected strategies)

A **domain policy** lets an entity method vary its behavior by a decision made
outside the entity — typically chosen per deployment from configuration — without
the entity, or its params, knowing which variant is active. It is the Strategy
pattern, kept pure and inside the domain.

- **A pure interface.** One decision, no I/O, no framework, no config. Example:
  `EmailVerificationPolicy` with `InitialStatus() Status`.
- **Lives in the entity's package** (`account/email_verification_policy.go`), not at
  the `domain` root. A policy consulted by a single entity is entity-specific; the
  root is for genuinely cross-entity domain services.
- **Its concrete strategies live in the domain too** — stateless, pure structs
  (`RequiredEmailVerification`, `OptionalEmailVerification`). No persistence, no I/O.
- **Injected into the entity method alongside the params struct — never inside it.**
  A policy is a behavioral dependency, not data, so it is a separate argument, not a
  params field. This is a deliberate addition to "methods take value objects": a
  method may *also* take an injected policy.
- **Query-shaped (CQS).** The policy method *decides and returns*; it never mutates
  the entity. The entity acts on the answer and keeps ownership of its own state.
- **The entity never chooses the strategy.** The **composition root** selects the
  concrete policy from config and injects it (see
  [`infrastructure-layer.md`](./infrastructure-layer.md)); the `application` service
  merely forwards it (see [`application-layer.md`](./application-layer.md)). Selecting
  a strategy from a config flag is wiring, not business logic — so the domain stays
  free of config knowledge and the application stays free of business rules.

| | Lives in | Injected? | Does |
|---|---|---|---|
| **Domain service** | `domain` root | No — called directly | Pure cross-entity logic across two or more entities. |
| **Domain policy** | entity package | Yes — into an entity method, alongside its params | Supplies one swappable decision the entity acts on; the concrete strategy is chosen from config at the composition root. |

```go
// account/email_verification_policy.go — a pure interface + its strategies
package account

// EmailVerificationPolicy decides the status a newly registered account starts in.
type EmailVerificationPolicy interface {
    InitialStatus() Status // query — decides, never mutates
}

type RequiredEmailVerification struct{} // chosen when EMAIL_VERIFICATION_REQUIRED=true
func (RequiredEmailVerification) InitialStatus() Status { return PendingVerification }

type OptionalEmailVerification struct{} // chosen when EMAIL_VERIFICATION_REQUIRED=false
func (OptionalEmailVerification) InitialStatus() Status { return Active }
```

```go
// account/account.go — the policy is a second argument, not a params field
func Register(params RegisterParams, verification EmailVerificationPolicy) (*Account, error) {
    // guard-clause each params field (id, email) ...
    account := &Account{
        Entity: domain.NewEntity(params.ID),
        email:  params.Email,
        status: verification.InitialStatus(), // the strategy decides; the entity accommodates
    }
    account.Record(NewAccountRegistered(account.ID(), account.status))
    return account, nil
}
```

A worked version lives in [`examples/identity/`](./examples/identity/README.md).

## Domain events

- **Only entities fire domain events** — recorded as the entity changes state. Value objects and domain services never do.
- Named in the **past tense** for a fact that happened (`OrderShipped`, `ProductRepriced`).
- **Each context declares its `Event` marker** at the domain root, in `domain/shared.go` (alongside `DomainError`).
- **Each entity has its own `events.go`** in its package (`order/events.go`, `product/events.go`), defining and recording its own events.
- **Immutable pure domain objects**, carrying **ids, never a whole entity**, implementing the context's `Event` marker.
- **No magic strings** — an event's name lives in a **named constant**; `EventName()` returns that constant, never an inline literal.
- **Recorded through the base `Entity`** — the entity calls the promoted `Record` from its own methods; the id and event slice live on the embedded base, not on each entity.
- **Exposed by two CQS-split methods** — `PullEvents()` returns a copy (query); `DrainEvents()` clears them (command). Both are promoted from the base `Entity`. One method that both returns and clears is not allowed.
- **In-context only** — no transport, no cross-context contract. Crossing a boundary is the **integration event's** job (see [`infrastructure-layer.md`](./infrastructure-layer.md)).
- **Recording is not dispatching** — the entity only records; a feature pulls, dispatches, and drains within its unit of work (see [`application-layer.md`](./application-layer.md)).

```go
// domain/shared.go (continued)
package domain

// Event marks a recorded domain fact, named in the past tense.
type Event interface {
    EventName() string
}
```

```go
// order/events.go
package order

const orderShippedEventName = "order.shipped"

// OrderShipped is recorded when an order ships. It carries the order's id.
type OrderShipped struct {
    orderID ID
}

func NewOrderShipped(orderID ID) OrderShipped { return OrderShipped{orderID: orderID} }

func (e OrderShipped) OrderID() ID       { return e.orderID }
func (e OrderShipped) EventName() string { return orderShippedEventName }
```
