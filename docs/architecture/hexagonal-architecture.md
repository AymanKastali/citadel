# Hexagonal Architecture (Ports & Adapters)

> **Status: draft — under discussion. Nothing here is final yet.**

Three layers, dependencies pointing inward only. The system is a **modular monolith**: one hexagon (module) per bounded context, each with its own `domain/`, `application/`, and `infrastructure/`.

Per-layer detail lives in its own file:

- [`domain-layer.md`](./domain-layer.md)
- [`application-layer.md`](./application-layer.md)
- [`infrastructure-layer.md`](./infrastructure-layer.md)

## The three layers

| Layer | Role | Contains |
|---|---|---|
| **domain** | Owns 100% of decision-making and business-rule enforcement. Pure. | Entities, value objects, domain services, domain events, domain errors. No I/O, no framework, no infrastructure types. |
| **application** | Orchestrator only — knows *what* to do, not *how*. Pure. | Services (inbound port implementations), inbound & outbound port interfaces (incl. the unit of work), commands/queries/results. |
| **infrastructure** | Adapters — knows *how*. The only impure layer. | Driving adapters (transport in), driven adapters (database, mail, clock, external services), composition root. Implements the ports. |

## Dependency rule

```
infrastructure ─→ application ─→ domain ─→ (nothing)
```

- **domain** imports nothing from `application` or `infrastructure`, and no third-party framework, driver, or transport.
- **application** imports **domain only**. It declares interfaces (ports) for every concretion it needs; never imports a concrete adapter.
- **infrastructure** may import **application** and **domain**, only to implement their interfaces.
- No business logic in an adapter. No adapter calls another adapter — everything routes through a port.

## Ports

| Port | Defined in | Implemented in | Called by |
|---|---|---|---|
| **Inbound** (service) | `application` | `application` (the service type) | driving adapters in `infrastructure` |
| **Outbound** (dependency) | `application` | driven adapters in `infrastructure` | `application` |

A port speaks the domain's language, never the technology's.

> Repository ports are the exception — they are declared in the **domain** (see [`domain-layer.md`](./domain-layer.md)). All other outbound ports are declared in the `application` layer.

## Translation at every boundary

```
external input ─(driving adapter)→ Command/Query ─(service)→ Domain model
Domain model ─(outbound port)→ [driven adapter maps to] persistence / wire format
```

- The persistence record is **not** the domain entity — the driven adapter maps between them.
- Domain entities are **never** returned over the wire — the driving adapter maps to a response DTO.

## Folder structure — one hexagon per bounded context

```
<context>/                  # a bounded context = one module = one hexagon
├── domain/                 # see domain-layer.md
├── application/            # see application-layer.md
└── infrastructure/         # see infrastructure-layer.md
```

Each new bounded context gets its own sibling hexagon. Cross-context communication goes through **integration events over a broker (or a port)** — never a direct import of another context's domain.

## Events & unit of work

- **Domain event** — a fact recorded by an entity, consumed inside the context (see [`domain-layer.md`](./domain-layer.md)). **Integration event** — the public cross-context contract, translated from a domain event and published over a broker (see [`infrastructure-layer.md`](./infrastructure-layer.md)).
- **A unit of work is allowed, but never bare** — state changes and their domain events commit **together** via a **transactional outbox**, so they never drift apart.
- End-to-end, per state-changing feature:

  ```
  open UoW → load → domain decides (records events) → persist
           → pull events, write cross-context ones to the outbox → drain → commit
  relay (separate) → reads outbox → publishes integration events → other contexts consume
  ```

## Guardrails

- **Command-Query Separation, everywhere.** A method either **changes state** (a command — returns nothing, or only an error) or **returns data** (a query — no mutation), never both. This holds in every layer: entities, services, ports, adapters.
- No third-party framework import inside `domain/` or `application/`.
- No conditional on domain state inside a service or adapter.
- No adapter-to-adapter calls.
- **Behavior that varies by configuration uses a domain policy** — a pure strategy interface the domain defines and an entity method takes alongside its params; the composition root picks the concrete strategy from config (see [`domain-layer.md`](./domain-layer.md)). The choice is wiring; the decision stays in the domain.
