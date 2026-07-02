# Domain code — worked example (ordering context)

> **Status: draft — under discussion.**

A reference for **how domain code must be written** in citadel: the clean-code
rules applied to the domain-layer structure. Illustrative only — not wired into
a build.

Layout follows [`domain-layer.md`](../../domain-layer.md): one directory per entity, with
`DomainError`, the `Event` marker, and the cross-entity domain service at the
`domain` root.

```
ordering/
└── domain/                       # package domain (root)
    ├── shared.go                 #   DomainError + Event marker + base Entity
    ├── add_product_to_order.go   #   domain service (cross-entity)
    ├── product/                  # package product
    │   ├── id.go  name.go  price.go   # value objects
    │   ├── events.go             #   ProductRepricedEvent — recorded by Product
    │   ├── errors.go             #   error factories → *domain.DomainError
    │   ├── product.go            #   entity
    │   └── repository.go         #   repository port
    └── order/                    # package order
        ├── id.go  quantity.go  line.go
        ├── events.go             #   OrderShippedEvent — recorded by Order
        ├── errors.go
        ├── order.go
        └── repository.go
```

## What each part demonstrates

### Value objects — `product/name.go`, `product/price.go`, `order/quantity.go`, `order/line.go`
- **Self-validating, full range** — the constructor enforces both bounds (**Fail Fast**): `Price` rejects non-positive and too-large; `Name` rejects empty and over-long. So a value object can only exist valid.
- **Immutable** — unexported fields, no setters; read via `Value()`, compare via `Equal`. `IsZero()` flags the bypassed zero value.
- **Named conditions** — each check is a named boolean (`nameIsMissing`, `amountIsZeroOrNegative`), guarded one at a time, faulty case first (**G19**).

### Entities — `product/product.go`, `order/order.go`
- **Embed the base `domain.Entity[ID]`** — id and event plumbing (`ID()`, `Record`, `PullEvents`, `DrainEvents`) are promoted, so neither entity re-declares them.
- Own their rules; state changes only through methods (**Tell-Don't-Ask**), and methods take value objects, not primitives.
- Constructors take a **params struct** and **guard-clause** each missing field.
- **Named conditions** as predicate methods (`order.hasShipped()`, `order.isEmpty()`), phrased positively (**G28**, **G29**).
- `Order.Lines()` returns a **copy**; commands (`AddLine`, `Ship`) do, queries (`Status`, `Lines`) answer (**CQS**).
- **`Reconstitute` beside each constructor** — `product.Reconstitute` / `order.Reconstitute` take a params struct of the full stored state (`order`'s includes status and lines) and just build the entity — no validation, no event, no policy; used only by the repository adapter.

### Errors — `shared.go`, `product/errors.go`, `order/errors.go`
- One `DomainError`; one **self-describing factory per violation** — business/invariant violations only, no "not found" or HTTP status.
- **Go-style names** carried by the package — `order.NewAlreadyShippedError`, not `NewOrderAlreadyShippedError`.

### Repository ports — `product/repository.go`, `order/repository.go`
- **Interfaces** declared by the domain; persistence-oriented names (`Create`, `Get`, `Exists`, `Update`, `Delete`).

### Domain events — `shared.go` (marker), `order/events.go`, `product/events.go`
- **Only entities fire events** — `Order.Ship()` records `OrderShippedEvent`, `Product.Reprice()` records `ProductRepricedEvent`; value objects and services never do.
- **The `Event` marker lives in `shared.go`** (the context's); **each entity has its own `events.go`** defining and recording its events.
- **Past-tense facts**, immutable, carrying an **id** (`order.ID`, `product.ID`), never the entity.
- **Two CQS-split methods, promoted from the base `Entity`** — `PullEvents()` reads, `DrainEvents()` clears. The application pulls, dispatches, and drains in its unit of work (see [`application-layer.md`](../../application-layer.md)).
- Crossing a boundary is the integration event's job — not shown here (pure-domain example).

### Domain service — `add_product_to_order.go`
- Cross-entity logic at the **domain root** — stateless, pure, no persistence. Coordinates the two entities; each still enforces its own rules.

## Note

Repository interfaces are named `Repository` (used as `product.Repository`,
`order.Repository`) — the package already carries the concept, so
`product.ProductRepository` would stutter.
