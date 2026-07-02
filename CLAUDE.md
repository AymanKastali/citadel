# Citadel

Citadel is an open-source authentication & IAM system in Go, built as a
**hexagonal (ports & adapters) modular monolith** — one self-contained hexagon per
bounded context. The repo is **docs-first and early stage**: the `docs/` tree is the
source of truth for how code must be structured; `internal/`, `cmd/`, and `tests/`
are not populated yet.

## Where to start

Documentation lives in `docs/`. **Start at [`docs/README.md`](docs/README.md)** — it
maps every document and gives a reading order. The canonical architecture rules are
in [`docs/architecture/hexagonal-architecture.md`](docs/architecture/hexagonal-architecture.md);
per-layer detail is in the sibling `domain-layer.md`, `application-layer.md`, and
`infrastructure-layer.md`.

## Architecture invariants (must-respect when writing code)

These are load-bearing. Follow them; see the linked docs for the full rationale.

- **Dependency rule points inward only:** `infrastructure → application → domain → (nothing)`. `domain` imports no framework/driver/transport; `application` imports `domain` only and depends on ports, never concrete adapters. See [`hexagonal-architecture.md`](docs/architecture/hexagonal-architecture.md).
- **No third-party framework imports in `domain/` or `application/`** — those layers stay pure.
- **Command–Query Separation everywhere** — a method either changes state (returns nothing/error) or returns data, never both. Applies to entities, services, ports, and adapters.
- **Repository ports live in the `domain`** (per entity), unlike all other outbound ports which live in `application`. See [`domain-layer.md`](docs/architecture/domain-layer.md).
- **Translate at every boundary** — persistence records and wire DTOs are never the domain entity; adapters map to/from the domain. See [`infrastructure-layer.md`](docs/architecture/infrastructure-layer.md).
- **State + domain events commit together** via a transactional outbox (a unit of work is allowed but never bare). Cross-context communication goes through integration events, never a direct import of another context's domain.
- **Config-varying behavior uses a domain policy** — a pure strategy interface in the entity's package, passed into the entity method *alongside* the params struct (never inside it); concrete strategies live in the domain, and the composition root selects one from config. See [`domain-layer.md`](docs/architecture/domain-layer.md).

## Scope notes

- Architecture docs are a **draft** under discussion.
- In [`docs/AUTH_FEATURES.md`](docs/AUTH_FEATURES.md), a checked box marks **intended MVP scope**, not a shipped feature.
