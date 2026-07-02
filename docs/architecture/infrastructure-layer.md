# Infrastructure Layer

> **Status: draft — under discussion. Nothing here is final yet.**

How the **infrastructure** layer is organized within a bounded context. It is the only impure layer: adapters translate a specific technology to the ports, and the composition root wires everything. See [`hexagonal-architecture.md`](./hexagonal-architecture.md) for the cross-cutting rules.

## Structure

```
infrastructure/
├── adapters/
│   ├── inbound/                # driving adapters → call inbound ports
│   └── outbound/               # driven adapters → implement outbound ports
└── config/                     # composition root: build adapters, wire ports
```

- **Driving adapters** receive external input (transport, jobs, CLI), map it to a `Command`/`Query`, and call an inbound port.
- **Driven adapters** implement the outbound ports (repositories, mail, clock, external services).
- **Adapters are dumb translators.** No business logic in an adapter, and no adapter calls another adapter — everything routes through a port.

## Mapping

- The **persistence model stays in the adapter** and is mapped to/from the domain entity — the persistence record is not the domain entity.
- Domain entities are **never returned over the wire** — the driving adapter maps to a response DTO.

## Composition root — selecting domain policies

- A **domain policy** is an injected strategy the domain defines but does not choose
  (see [`domain-layer.md`](./domain-layer.md)). The **composition root** (`config/`)
  reads configuration — a `.env` / environment variable — and picks the concrete
  strategy to inject into the application service.
- **This is wiring, not business logic.** Mapping a flag to a strategy is a `switch`
  in the composition root; the decision the strategy encodes stays in the domain. No
  adapter or service branches on the config value at request time.

```go
// infrastructure/config — reads EMAIL_VERIFICATION_REQUIRED, selects a domain policy
func emailVerificationPolicy(cfg Config) account.EmailVerificationPolicy {
    if cfg.EmailVerificationRequired {
        return account.RequiredEmailVerification{}
    }
    return account.OptionalEmailVerification{}
}

// ... then inject it when building the service:
// services.NewRegisterAccountService(uow, accounts, emailVerificationPolicy(cfg))
```

## Unit of work, outbox, and integration events

- The **unit of work port is implemented over the DB transaction** (a driven adapter); its writes commit or roll back together.
- **Transactional outbox** — integration events are written in the **same transaction** as the state change, so there is no dual write.
- **Relay / publisher** — a separate driven adapter reads the outbox, publishes to the broker, and marks events sent, keeping publishing off the request's critical path.
- **Inbound messaging adapters** — driving adapters that consume other contexts' integration events and map them to a `Command`.
- **Translation at the boundary** — the publisher maps a **domain event** (domain types) to an **integration event** (the public contract: stable, serializable, versioned, primitives/DTO). Only events other contexts care about become integration events.
- **Cross-context communication goes through integration events (or a port)** — never a direct import of another context's domain.

## Naming

- **Driving adapter:** by transport — `<Transport>OrderHandler`.
- **Driven adapter:** technology + capability — `<Tech>OrderRepository`, `<Tech>EmailSender`.
- **Persistence model:** stays in the adapter.
- **Edge/DTO:** `Request` / `Response` at the transport edge.
