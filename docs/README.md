# Citadel — Documentation

**Citadel** is an open-source authentication and identity & access management (IAM)
system written in Go. It is built as a **hexagonal (ports & adapters) modular
monolith**: one self-contained hexagon per bounded context, each with its own
`domain/`, `application/`, and `infrastructure/` layers and dependencies pointing
inward only.

The project is **docs-first and early stage** — this `docs/` tree defines the
architecture and the target feature set before the implementation lands. Treat
these documents as the source of truth for how code must be structured.

## Status

- The **architecture docs are a draft** under discussion (each carries a status
  banner). The rules they describe are the intended design.
- In [`AUTH_FEATURES.md`](./AUTH_FEATURES.md), a checked box (`[x]`) marks a feature
  as **intended MVP scope** — not a shipped or implemented feature. Everything
  unchecked is deferred post-MVP.

## Doc map

| Document | What it covers |
|---|---|
| [`architecture/hexagonal-architecture.md`](./architecture/hexagonal-architecture.md) | The cross-cutting rules: the three layers, the dependency rule, ports, boundary translation, events & unit of work, and the global guardrails. **Start here.** |
| [`architecture/domain-layer.md`](./architecture/domain-layer.md) | How the pure domain layer is organized: entities, value objects, domain errors, repository ports, domain services, and domain events. |
| [`architecture/application-layer.md`](./architecture/application-layer.md) | The orchestration layer: inbound/outbound ports, services, the unit-of-work + domain-event flow, and precondition-vs-business-rule split. |
| [`architecture/infrastructure-layer.md`](./architecture/infrastructure-layer.md) | The only impure layer: driving/driven adapters, persistence mapping, the transactional outbox, and integration events. |
| [`architecture/examples/ordering/`](./architecture/examples/ordering/README.md) | A worked, illustrative domain example (an `ordering` context) showing the domain-layer rules applied to real Go code. |
| [`architecture/examples/identity/`](./architecture/examples/identity/README.md) | A worked example (an `identity` context) showing a **domain policy** — an injected strategy chosen from config — via account registration and email verification. |
| [`AUTH_FEATURES.md`](./AUTH_FEATURES.md) | A vendor-agnostic inventory of modern auth/IAM features, with MVP scope marked by checkboxes. |

## Suggested reading order

1. [`architecture/hexagonal-architecture.md`](./architecture/hexagonal-architecture.md) — the overall shape and rules.
2. [`architecture/domain-layer.md`](./architecture/domain-layer.md) — the innermost layer.
3. [`architecture/application-layer.md`](./architecture/application-layer.md) — how features orchestrate the domain.
4. [`architecture/infrastructure-layer.md`](./architecture/infrastructure-layer.md) — how the outside world plugs in.
5. [`architecture/examples/ordering/`](./architecture/examples/ordering/README.md) — the rules in practice.
6. [`architecture/examples/identity/`](./architecture/examples/identity/README.md) — a domain policy (injected strategy) in practice.
7. [`AUTH_FEATURES.md`](./AUTH_FEATURES.md) — what Citadel aims to build.
