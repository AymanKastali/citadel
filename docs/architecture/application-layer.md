# Application Layer

> **Status: draft — under discussion. Nothing here is final yet.**

How the **application** layer is organized within a bounded context. It is a pure orchestrator: it knows *what* to do, not *how*. See [`hexagonal-architecture.md`](./hexagonal-architecture.md) for the cross-cutting rules.

## Structure — ports by role

Everything lives under a single `ports/` tree, grouped by technical role:

```
application/
└── ports/
    ├── inbound/                # inbound port interfaces + their Command/Query/Result
    ├── outbound/               # outbound port interfaces (UnitOfWork, Clock, EmailSender, gateways)
    └── services/               # service implementations of the inbound ports
```

- **`inbound/`** (package `inbound`) — the action interfaces the outside calls (`PlaceOrder`, `GetOrder`), together with their DTOs. Since they share one package, the DTOs are feature-prefixed: `PlaceOrderCommand`, `GetOrderQuery`, `OrderResult`.
- **`outbound/`** (package `outbound`) — the interfaces the core needs from the world: `UnitOfWork`, `Clock`, `EmailSender`, gateways. **Repository ports are the exception** — they live in the domain (see [`domain-layer.md`](./domain-layer.md)).
- **`services/`** (package `services`) — the service structs that implement the inbound ports (`PlaceOrderService`, `GetOrderService`), built via `New(...)`.

## The service is an orchestrator only

The application enforces no business rule itself; it wires the domain to its ports.

**Command (state-changing) — always:**

**open unit of work → load (outbound port) → hand data to the domain → domain decides (records events) → persist (outbound port) → pull & enqueue events → drain → commit.**

**Query (read-only) — simpler:**

**load (outbound port) → map to `Result`.** No unit of work, no domain events — nothing changes state.

## Unit of work + domain events

- A **unit of work** makes one feature's writes atomic. It is an **outbound port**, implemented over the DB transaction in `infrastructure`.
- **Closure form** — the transaction can't leak, commit/rollback is automatic:

  ```go
  // application/ports/outbound/unit_of_work.go
  package outbound

  type UnitOfWork interface {
      Do(ctx context.Context, fn func(ctx context.Context) error) error
  }
  ```

  (An explicit `Begin` / `Commit` / `Rollback` trio is the fallback.)
- **Allowed, but never bare.** A state-changing feature uses the unit of work **and** domain events together: `entity.PullEvents()` → write cross-context ones to the **outbox** → `entity.DrainEvents()`, all inside the unit of work so state and events commit atomically. A plain transaction that commits only rows is not allowed.
- **Read-only features need no unit of work.**
- **Domain-event handlers are application services** — in-context reactions, each running its own orchestration.
- Publishing to other contexts happens at the boundary, not here — drained events become **integration events** (see [`infrastructure-layer.md`](./infrastructure-layer.md)).

### Precondition check vs. business rule

- **Application checks** only what needs no domain knowledge — input well-formed, referenced record exists, port returned not-found. These gate whether we ask the domain at all.
- **The domain decides** validity, permission, and state transitions.
- A check that could change with the business belongs in the domain.

## Inbound ports and services

An inbound port is an **action interface** in `inbound/`; its implementation is a **service** in `services/`. Each has one `Handle` method, split by feature kind under **CQS**:

- **Command:** `Handle(ctx, PlaceOrderCommand) error` — changes state, returns only an error. The id is caller-supplied on the command.
- **Query:** `Handle(ctx, GetOrderQuery) (OrderResult, error)` — returns data, mutates nothing.

```go
// application/ports/inbound/place_order.go
package inbound

// PlaceOrder places an order. Driving adapters depend on this interface.
type PlaceOrder interface {
    Handle(ctx context.Context, cmd PlaceOrderCommand) error
}

// PlaceOrderCommand is the input into the service; the caller supplies the id.
type PlaceOrderCommand struct {
    OrderID string
}
```

The service is orchestration only — the domain makes every decision:

```go
// application/ports/services/place_order.go
package services

type PlaceOrderService struct {
    uow    outbound.UnitOfWork
    orders order.Repository
}

func NewPlaceOrderService(uow outbound.UnitOfWork, orders order.Repository) *PlaceOrderService {
    return &PlaceOrderService{uow: uow, orders: orders}
}

func (s *PlaceOrderService) Handle(ctx context.Context, cmd inbound.PlaceOrderCommand) error {
    return s.uow.Do(ctx, func(ctx context.Context) error {
        newOrder, err := order.NewOrder(cmd.toParams()) // domain builds & validates
        if err != nil {
            return err
        }
        if err := s.orders.Create(ctx, newOrder); err != nil { // persist via port
            return err
        }
        events := newOrder.PullEvents()        // query — read recorded events
        // write cross-context events to the outbox here (same transaction)
        newOrder.DrainEvents()                 // command — clear them
        return nil
    })
}
```

### Injecting a domain policy

When an entity method takes a **domain policy** (an injected strategy — see
[`domain-layer.md`](./domain-layer.md)), the service is **constructor-injected** with
the concrete policy and passes it into the entity method **alongside** the command's
params. The service makes no decision itself — it never inspects config or branches on
which strategy is active; it just forwards what the composition root wired in.

```go
// application/ports/services/register_account.go
package services

type RegisterAccountService struct {
    uow          outbound.UnitOfWork
    accounts     account.Repository
    verification account.EmailVerificationPolicy // wired at the composition root, from config
}

func NewRegisterAccountService(
    uow outbound.UnitOfWork,
    accounts account.Repository,
    verification account.EmailVerificationPolicy,
) *RegisterAccountService {
    return &RegisterAccountService{uow: uow, accounts: accounts, verification: verification}
}

func (s *RegisterAccountService) Handle(ctx context.Context, cmd inbound.RegisterAccountCommand) error {
    return s.uow.Do(ctx, func(ctx context.Context) error {
        // policy passed ALONGSIDE the params — the domain decides, the service only forwards
        newAccount, err := account.Register(cmd.toParams(), s.verification)
        if err != nil {
            return err
        }
        if _, err := s.accounts.Create(ctx, newAccount); err != nil {
            return err
        }
        newAccount.PullEvents()  // write cross-context events to the outbox here
        newAccount.DrainEvents()
        return nil
    })
}
```

## Naming

- **Inbound port:** an action interface in `inbound/` — `PlaceOrder`, `GetOrder`.
- **Service:** the implementation in `services/` — `PlaceOrderService`, `GetOrderService`, built via `New...`.
- **Outbound port:** the capability, no `Port` suffix — `UnitOfWork`, `Clock`, `EmailSender`, `PaymentGateway`.
- **Edge/DTO** (in `inbound/`, feature-prefixed): `PlaceOrderCommand` (into a command port), `GetOrderQuery` (into a query port), `OrderResult` (read data out of a query port).
