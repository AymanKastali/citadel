# Domain code — worked example (identity context)

> **Status: draft — under discussion.**

A reference for **how a domain policy is written** in citadel: a pure interface,
injected into an entity method, that lets one deployment differ from another
without the entity knowing which variant is active. The worked entity is a full
`account` — it carries its whole lifecycle, so this doubles as a complete worked
entity. Illustrative only — not wired into a build.

The motivating case: at registration, email verification may be **required** or
**not**, decided by a `.env` value. `account.Register` takes an
`EmailVerificationPolicy` **alongside** its params struct; the composition root
picks the concrete strategy from config and injects it.

Layout follows [`domain-layer.md`](../../domain-layer.md): one directory per
entity, with `DomainError`, the `Event` marker, and the base `Entity` at the
`domain` root. A single entity (`account`) here, so there is no cross-entity
domain service.

```
identity/
└── domain/                              # package domain (root)
    ├── shared.go                        #   DomainError + Event marker + base Entity
    └── account/                         # package account
        ├── doc.go                       #   package headline + reading order
        ├── id.go  email.go              # value objects
        ├── password_hash.go             #   value object (already-hashed password)
        ├── status.go                    #   the account lifecycle (an enum)
        ├── email_verification_policy.go #   the domain policy + its two strategies
        ├── events.go                    #   the account's domain events (…Event)
        ├── errors.go                    #   error factories → *domain.DomainError
        ├── account.go                   #   entity + Register/Reconstitute + commands
        └── repository.go                #   repository port
```

## What each part demonstrates

### Domain policy — `account/email_verification_policy.go`
- **A pure interface (a Strategy)** — `EmailVerificationPolicy` has one method,
  `InitialStatus() Status`. No I/O, no framework, no config.
- **Lives in the entity package**, not at the domain root — a policy consulted by
  a single entity is entity-specific, unlike a cross-entity domain service.
- **Concrete strategies live in the domain too** — `RequiredEmailVerification` and
  `OptionalEmailVerification`, each a stateless struct returning its decision.
- **Query-shaped (CQS)** — the policy *decides and returns*; it never mutates the
  account. The entity acts on the answer and keeps ownership of its own state.
- **Chosen at the composition root, not by the entity** — infrastructure reads
  `EMAIL_VERIFICATION_REQUIRED` and injects the matching strategy. Selecting a
  strategy from config is wiring, not business logic; the decision itself stays in
  the domain.

### Entity — `account/account.go`
- **The policy is passed alongside the params struct, never inside it** —
  `Register(params RegisterParams, verification EmailVerificationPolicy)`. A policy
  is a behavioral dependency, not data, so it is not a `RegisterParams` field.
- Constructor takes a **params struct** of already-valid value objects and
  **guard-clauses** each missing field (**Fail Fast**).
- **The entity accommodates the strategy** — it sets its own `status` from
  `verification.InitialStatus()` and records `AccountRegisteredEvent`. The strategy
  decides; the entity owns the state change and the event.
- Embeds the base `domain.Entity[ID]`; **commands** (`VerifyEmail`, `ChangePassword`,
  `Suspend`, `Reactivate`, `Deactivate`, `Login`) mutate and record an event, while
  **queries** (`Email`, `PasswordHash`, `Status`, `LastLoginAt`) answer (**CQS**). Each
  command guards its illegal transitions first (Fail Fast).
- **`Reconstitute` sits beside `Register`** — it takes a `ReconstituteParams` (the full
  stored state, including `Status` and `LastLoginAt`) and just loads it into the entity: no
  validation, no event, no policy. The repository adapter uses it to rebuild a stored
  account. See [`domain-layer.md`](../../domain-layer.md).

### Value objects — `account/id.go`, `account/email.go`, `account/password_hash.go`
- **Self-validating, full range** — `ID` rejects empty and over-long; `Email` rejects
  empty, over-long, and non-RFC-5322 addresses (validated with the stdlib `net/mail`
  parser, bare address only) and normalizes to lower case; `PasswordHash` stores an
  already-hashed value (plaintext never enters the domain), rejecting empty and over-long
  with no format check.
- **Immutable** — unexported fields, read via `Value()`, compare via `Equal`;
  `IsZero()` flags the bypassed zero value.

### Status — `account/status.go`
- A small lifecycle enum. Only the two **registration-time** statuses
  (`PendingVerification`, `Active`) are ever chosen by the policy; the rest are
  reached by later transitions the account owns. A `String()` helper renders the
  status for logging / persistence mapping (no behavior change).

### Errors — `account/errors.go`, `shared.go`
- One `DomainError`; one **self-describing factory per violation** — business /
  invariant violations only, no "not found" or HTTP status.

### Domain events — `account/events.go`
- **Only the entity fires them** — one per command (`AccountRegisteredEvent`,
  `EmailVerifiedEvent`, `PasswordChangedEvent`, `AccountSuspendedEvent`,
  `AccountReactivatedEvent`, `AccountDeactivatedEvent`, `AccountLoggedInEvent`), never
  the entity itself.
- **Each carries the account id**; `AccountRegisteredEvent` also carries the starting
  status and `AccountLoggedInEvent` the login timestamp.
- **Past-tense facts**, immutable, with their wire names in **one grouped `const` block**.
- Pulled and drained by the application in its unit of work (see
  [`application-layer.md`](../../application-layer.md)).

### Repository port — `account/repository.go`
- **An interface** declared by the domain, with persistence-oriented names
  (`Create`, `Get`, `Exists`, `Update`, `Delete`, plus the email lookups `GetByEmail` /
  `ExistsByEmail`).

## Where the strategy is chosen

The domain never picks a strategy. The composition root does, from config —
see [`application-layer.md`](../../application-layer.md) for the service that
forwards the injected policy into `Register`, and
[`infrastructure-layer.md`](../../infrastructure-layer.md) for reading
`EMAIL_VERIFICATION_REQUIRED` and selecting the concrete policy.
