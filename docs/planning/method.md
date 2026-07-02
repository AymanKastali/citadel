# The Planning Method

> **Status: living.** How every Citadel planning doc is produced and reviewed.

Planning docs are written *before* code and are the implementation spec Claude Code
reads to write Go. This document defines the method they all follow.

## Layer-by-layer, inside-out

Planning proceeds one layer at a time, in the direction the dependency rule allows
(`infrastructure → application → domain → nothing`), so each layer is planned only
after the layer it depends on is settled:

1. **Domain** first — entities, value objects, events, repository ports, domain
   services and policies. Nothing above it is planned until the domain is done.
2. **Application** next — inbound/outbound ports, services, unit-of-work flow.
3. **Infrastructure** last — adapters, persistence mapping, composition root.

## Documentation first, code last

Every planning phase produces **documentation**. No code is written until the whole
system is documented layer by layer (domain → application → infrastructure). Writing
code is the **final step** of the effort, not of any single phase — and it follows the
same bottom-up order the docs were written in.

## One entity at a time, built bottom-up

A layer starts with an **overview document** (scope, design decisions, package layout,
the relationship map, the shared root, and an index). Then **each entity is planned to
full depth in its own document**, and finished before the next begins. There is no
separate skeleton pass — an entity's document is built **bottom-up, piece by piece**:

1. **Value objects** — each one's wrapped type, *all* validation rules (bounds, format,
   normalization), constructor, `Value()`/`Equal()`/`IsZero()`, and the errors it raises.
2. **Entity + constructors** — struct fields (value objects + typed-ID references +
   the embedded base); the business constructor (params struct + any injected policy)
   that records the creation event; and a **`Reconstitute`** that just loads the entity
   from full persisted state (same params-in/entity-out shape; no validation, no event,
   no policy — used only by the repository adapter).
3. **Commands** — each state transition: its guards/invariants, the change, the event.
4. **Queries, events, errors, repository** — getters; full event definitions; the
   error catalog; and the repository port.

## Phases and checkpoints

A planning doc is built in **numbered phases**, and each entity's pieces above are
their own steps. After each piece, work **stops for review** ("check and decide")
before the next begins. This keeps each reviewable unit small and stops a wrong
assumption from propagating.

## Verification

Planning docs are **reviewed, not compiled**. Every phase is checked against the
relevant rules in [`../architecture/`](../architecture/) — for the domain layer, that
is [`../architecture/domain-layer.md`](../architecture/domain-layer.md) and the worked
[`../architecture/examples/`](../architecture/examples/). Typical checks: one directory
per entity, self-validating value objects, references by typed ID (never embedding),
Command–Query Separation on every method, repository ports using persistence verbs and
living in the entity package, past-tense events with the `Event` type suffix and named
wire-name constants, and config-varying behavior expressed as an injected domain policy.

## Scope discipline

Each planning doc names what it deliberately leaves out (other layers, other bounded
contexts, concerns handled elsewhere) so nothing is silently dropped and no reviewer
assumes coverage that is not there.
