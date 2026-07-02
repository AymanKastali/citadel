# Citadel — Planning

> **Status: living.** Planning docs are written *before* code and reviewed
> phase by phase. They are the implementation spec Claude Code reads to write Go.

This tree turns the architecture (see [`../architecture/`](../architecture/)) and
the MVP feature set (see [`../AUTH_FEATURES.md`](../AUTH_FEATURES.md)) into concrete,
buildable plans. Where `docs/architecture/` says *how any code must be structured*,
`docs/planning/` says *what to build, in what order*, for a specific phase.

## How planning works

The method — layer-by-layer, skeleton-then-deep-dive, and the phase/checkpoint
workflow — is documented separately in
[**`method.md`**](./method.md). Read it first.

## Index

| Doc | Phase | Layer | Status |
|---|---|---|---|
| [`method.md`](./method.md) | — | — | The planning method |
| [`phase-1-identity/domain-layer.md`](./phase-1-identity/domain-layer.md) | 1 (MVP) | Identity · Domain | Skeleton in progress |

More docs (Identity domain deep-dives, then application/infrastructure) are added as
each pass is approved.
