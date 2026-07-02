# Identity · Domain · `session`

> **Status: complete — pieces (a) value objects, (b) entity + constructors, (c) commands, and (d) queries, events, errors & repository port all done.**
> Full deep-dive for the `session` entity. Part of the Identity domain layer; see the
> [overview](../domain-layer.md) (esp. the *"Session unifies refresh tokens"* design
> note) and the rules in
> [`domain-layer.md`](../../../architecture/domain-layer.md). Mirrors the
> [`account`](./account.md) deep-dive; several parts (`ID`, `SecretHash`) are modeled
> directly on account's `ID` / `PasswordHash`.

## What `session` is

An **authenticated session** — the persistent server-side record of one login on one
device. It is the concept formerly called the "refresh token", **renamed to name the
thing itself**: the token is merely the secret that proves possession of the session.
It holds the owning `account.ID`, a `Kind` discriminator, the **hash** of its secret, a
lifecycle `Status`, and an expiry.

A session is **long-lived and mutated in place**: rotation swaps the stored secret hash
on the *same* row for the whole duration of the login, so **one Session *is* one login
(one device)**. There is exactly one live row per device, not a chain of one-shot rows.
The `Kind` is `Refresh` for the MVP; `ServerSide` is **reserved** for a future stateful
opaque-session mode and is never instantiated yet. Package:
`internal/identity/domain/session`.

## Build order (bottom-up)

- **(a) Value objects** — `ID`, `Kind`, `SecretHash`, `ExpiresAt`, `Status` ✅
- **(b) Entity struct + `Open` + `Reconstitute`** ✅
- **(c) Commands** — `Rotate`, `Revoke` ✅
- **(d) Queries, events, errors, repository port** ✅ ← *this checkpoint*

---

## (a) Value objects

One file per value object in the `session` package. The constructor-validated ones
(`ID`, `SecretHash`, `ExpiresAt`) are immutable (unexported field), self-validating in
their constructor, with `Value()` / `Equal(other)` / `IsZero()`, and a zero value that
is invalid. The enums (`Kind`, `Status`) are **set by the entity, never parsed** from
raw external input, so they carry no `New…` constructor. Business/invariant failures
return `*domain.DomainError` via a factory in `session/errors.go` (piece **d**).

### `ID` — `id.go` (modeled on account's `ID`)

Identical in shape to
[`account.ID`](./account.md#id--idgo-reuse-verbatim-from-the-example).

- **Wraps:** `string` (an opaque identifier — UUID/ULID, assigned by the application, not generated in the domain).
- **Constructor:** `NewID(raw string) (ID, error)`.
- **Rules:** trim surrounding whitespace; reject empty (`NewEmptyIDError`); reject
  length `> maxIDLength` where `maxIDLength = 64` (`NewIDTooLongError(length)`).
- **Methods:** `Value() string`, `Equal(other ID) bool`, `IsZero() bool`.

### `Kind` — `kind.go` (enum, set by the entity)

Discriminates what mechanism the session backs. Like `Status`, it is an **enumerated
type set by the entity** (chosen at `Open` from the caller's `OpenParams.Kind`, itself
produced from config/wiring — not parsed from a wire string here); it has no `New…`
constructor. Persistence ↔ enum mapping is an infrastructure concern.

- **Type:** `type Kind int`, values via `iota`:
  - `Refresh` — a refresh-token session: the client holds the secret and presents it to
    rotate/refresh. **The only `Kind` used in the MVP.**
  - `ServerSide` — **reserved.** A future stateful/opaque server-side session mode. It
    is defined so the discriminator and its persistence mapping are stable, but the
    domain **never instantiates it** in the MVP, and `Rotate` is guarded to reject it
    (piece **c**).
- **Suggested helper:** `String() string` for logging/mapping (no behavior change).

### `SecretHash` — `secret_hash.go` (modeled on account's `PasswordHash`)

Holds the **hash of the session/refresh secret**. The raw secret **never enters the
domain**: the application's secret-hasher (an outbound port) hashes the freshly issued
secret on the way in and verifies a presented secret on the way out. This VO only stores
and structurally compares the hash; it is **algorithm-agnostic** (SHA-256, HMAC,
argon2id — the domain neither knows nor checks). Modeled directly on
[`account.PasswordHash`](./account.md#passwordhash--password_hashgo-new).

- **Wraps:** `string` (the encoded hash).
- **Constructor:** `NewSecretHash(raw string) (SecretHash, error)`.
- **Rules:** reject empty (`NewEmptySecretHashError`); reject length
  `> maxSecretHashLength` where `maxSecretHashLength = 255` (comfortably fits any encoded
  hash while rejecting junk) (`NewSecretHashTooLongError(length)`). **No trimming, no
  normalization** — hash encodings are exact and case-sensitive. **No format/algorithm
  check** — that would couple the domain to a hashing scheme.
- **Methods:** `Value() string`, `Equal(other SecretHash) bool`, `IsZero() bool`.
- **CQS/security note:** `Equal` is a plain structural comparison of two stored hashes
  (e.g. change-detection during rotation); it is **not** secret verification. Verifying a
  presented plaintext secret against the hash is the hasher port's job in the
  application, never here. The raw secret is **never stored, logged, or carried in an
  event** — only its hash.

### `ExpiresAt` — `expires_at.go` (wraps `time.Time`)

The moment the session stops being valid. Wraps the stdlib `time.Time` (stdlib `time`
is allowed; no third-party packages) so expiry is a first-class, self-validating value.

- **Wraps:** `time.Time`.
- **Constructor:** `NewExpiresAt(raw time.Time) (ExpiresAt, error)`.
- **Rules:** reject the zero time (`NewZeroExpiresAtError`) — a session must always carry
  a concrete expiry. No other bound here: the lifetime is chosen by the application from
  config, and "already in the past" is evaluated *live* against `now` (see
  `IsExpired`), not fixed at construction.
- **Methods:** `Value() time.Time`, `Equal(other ExpiresAt) bool`, `IsZero() bool`.
- **Note:** the domain never reads the clock; every expiry check takes `now time.Time`
  from the application (see piece **d**).

### `Status` — `status.go` (enum, set by the entity)

The session's place in its lifecycle. Like account's `Status`, it is an **enumerated
type set by the entity itself** (at `Open`, or by a transition command); it is never
parsed from raw external input, so it has no `New…` constructor. Persistence ↔ enum
mapping is an infrastructure concern.

- **Type:** `type Status int`, values via `iota`:
  - `Active` — usable: may be rotated and accepted. The **starting** status at `Open`.
  - `Revoked` — invalidated (logout, reuse detection, or an operator action);
    **terminal**. Cannot be rotated or re-activated.
  - `Expired` — the stored status once the session is known to be past its
    `ExpiresAt`. This is a **lazily-persisted convenience marker**, not the source of
    truth for liveness — see the note on [`Expired` vs `IsExpired`](#expired-status-vs-computed-isexpirednow) below.
- **Suggested helper:** `String() string` for logging/mapping (no behavior change).
- Legal transitions are specified with the commands in piece **(c)**, not here.

### Cross-entity reference — `account.ID`

The owning account is referenced by its **typed ID value object** `account.ID`
(imported from the `account` package), never by embedding the entity — per the
[relationship map](../domain-layer.md#relationship-map). It is validated by `account`'s
own constructor; `session` treats it as an already-valid value and only guards that it
is non-zero at `Open`.

---

## (b) Entity & constructors

File: `session/session.go`.

### Struct

The entity embeds the base `domain.Entity[ID]` (id + recorded events) and holds its
state as value objects, plus the owning account referenced by typed id.

```go
type Session struct {
	domain.Entity[ID]
	accountID  account.ID
	kind       Kind
	secretHash SecretHash
	status     Status
	expiresAt  ExpiresAt
}
```

State changes only through the methods below; callers read via getters (piece **d**),
never by assigning fields.

### `Open` — the business constructor

Builds a **new** session from already-valid value objects. Each field is guard-claused
(Fail Fast); the session starts `Active`; it records `SessionOpenedEvent` (payload defined in
piece **d**). No policy is needed — the caller supplies the `Kind` and the
config-derived `ExpiresAt`.

```go
type OpenParams struct {
	ID         ID
	AccountID  account.ID
	Kind       Kind
	SecretHash SecretHash
	ExpiresAt  ExpiresAt
}

func Open(params OpenParams) (*Session, error) {
	idIsMissing := params.ID.IsZero()
	if idIsMissing {
		return nil, NewEmptyIDError()
	}
	accountIsMissing := params.AccountID.IsZero()
	if accountIsMissing {
		return nil, NewEmptyAccountIDError()
	}
	secretHashIsMissing := params.SecretHash.IsZero()
	if secretHashIsMissing {
		return nil, NewEmptySecretHashError()
	}
	expiresAtIsMissing := params.ExpiresAt.IsZero()
	if expiresAtIsMissing {
		return nil, NewZeroExpiresAtError()
	}
	session := &Session{
		Entity:     domain.NewEntity(params.ID),
		accountID:  params.AccountID,
		kind:       params.Kind,
		secretHash: params.SecretHash,
		status:     Active, // sessions always start usable
		expiresAt:  params.ExpiresAt,
	}
	session.Record(NewSessionOpenedEvent(session.ID(), session.accountID))
	return session, nil
}
```

### `Reconstitute` — rebuild from persistence

The counterpart used only by the repository adapter. It takes the **full persisted
state** — a superset of `OpenParams` that also carries `Status` — and **just loads** it
into a fresh entity: no validation, no event, no policy, no error (see the
[Reconstitution rule](../../../architecture/domain-layer.md#reconstitution-rebuilding-from-persistence)).

```go
type ReconstituteParams struct {
	ID         ID
	AccountID  account.ID
	Kind       Kind
	SecretHash SecretHash
	Status     Status
	ExpiresAt  ExpiresAt
}

func Reconstitute(params ReconstituteParams) *Session {
	return &Session{
		Entity:     domain.NewEntity(params.ID),
		accountID:  params.AccountID,
		kind:       params.Kind,
		secretHash: params.SecretHash,
		status:     params.Status,
		expiresAt:  params.ExpiresAt,
	}
}
```

---

## (c) Commands

Two commands, each CQS (mutate state, return only `error`), each guard-claused, each
recording exactly one event. Both mutate the **same** row in place — there is no
"create a successor session" step.

### `Rotate` — swap the secret hash in place

The heart of the mutate-in-place design. On each refresh the application hashes a
**newly issued** secret and hands the hash to `Rotate`, which replaces the stored one on
the same session. The old secret's hash is gone; presenting the old secret afterwards
will no longer match (the basis for reuse detection — see the note below).

- **Guard `Kind == Refresh`** — only refresh sessions rotate; a `ServerSide` (reserved)
  session must not (`NewCannotRotateNonRefreshError`).
- **Guard the session is usable** — `status == Active` (`NewSessionNotActiveError`) and
  not past expiry (`NewExpiredError`). A revoked or expired session cannot be rotated.
- **Swap** `secretHash` to `newSecretHash`; **status stays `Active`**; **expiry is
  unchanged** (lifetime extension, if any, is an application decision that would pass a
  fresh `ExpiresAt` — not modeled here for the MVP).
- **Record `SessionRotatedEvent`** (carries session id + account id — **never** the secret).

```go
func (session *Session) Rotate(newSecretHash SecretHash, now time.Time) error {
	notRefresh := session.kind != Refresh
	if notRefresh {
		return NewCannotRotateNonRefreshError()
	}
	notActive := session.status != Active
	if notActive {
		return NewSessionNotActiveError()
	}
	expired := session.IsExpired(now)
	if expired {
		return NewExpiredError()
	}
	session.secretHash = newSecretHash
	session.Record(NewSessionRotatedEvent(session.ID(), session.accountID))
	return nil
}
```

> `now` is passed in (the domain reads no clock) so the "not expired" guard is checked
> live, independent of whether the stored `Status` has been lazily set to `Expired`.

### `Revoke` — invalidate this session (per device)

Terminates one session — used for logout on a device, and for **reuse detection**
(revoke the compromised device's session). Idempotency is a guard, not a silent no-op:
re-revoking is a business error.

- **Guard not already revoked** — `status != Revoked` (`NewAlreadyRevokedError`).
- **Set** `status = Revoked` (terminal).
- **Record `SessionRevokedEvent`** (carries session id + account id).

```go
func (session *Session) Revoke() error {
	alreadyRevoked := session.status == Revoked
	if alreadyRevoked {
		return NewAlreadyRevokedError()
	}
	session.status = Revoked
	session.Record(NewSessionRevokedEvent(session.ID(), session.accountID))
	return nil
}
```

> "Log out everywhere" is **not** a per-entity command — it spans every session of an
> account and is driven by the application via the repository's `RevokeAllByAccount`
> (piece **d**), which loads each session and calls `Revoke()`.

---

## (d) Queries, events, errors, repository

### Queries (no mutation — CQS)

Read-only accessors and derived predicates. The expiry predicates take `now time.Time`
from the application; the entity never reads the clock.

```go
func (session *Session) AccountID() account.ID  { return session.accountID }
func (session *Session) Kind() Kind              { return session.kind }
func (session *Session) Status() Status          { return session.status }
func (session *Session) SecretHash() SecretHash  { return session.secretHash }

// IsExpired: live, clock-based — true once now is at/after the expiry instant,
// regardless of the stored Status.
func (session *Session) IsExpired(now time.Time) bool {
	return !now.Before(session.expiresAt.Value())
}

// IsActive: usable right now — Active status AND not yet expired.
func (session *Session) IsActive(now time.Time) bool {
	return session.status == Active && !session.IsExpired(now)
}
```

- **`SecretHash()`** exposes only the stored **hash** — the raw secret is never in the
  domain, so nothing can leak it. It exists for the application to compare a
  hashed-presented secret against the stored hash during rotation/verification.
- **`IsActive`** is the single question the application asks before honoring a session:
  it combines the stored `Status` with a **live** expiry check, so a not-yet-swept
  `Active` row whose `ExpiresAt` has passed is still correctly reported as *not* active.

### `Expired` status vs computed `IsExpired(now)`

These are deliberately two things and must stay consistent:

- **`IsExpired(now)` is the source of truth for liveness** — a pure, clock-based
  computation against `ExpiresAt`. It is always correct the instant expiry passes,
  without any write.
- **`Status.Expired` is a lazily-persisted marker** — a session may sit as `Active` in
  storage after its `ExpiresAt`; a background sweep (or the next load) may *later* set
  the stored status to `Expired`. That write is an optimization/bookkeeping step, not
  the definition of expiry.
- **Therefore `IsActive(now)` checks both** — `status == Active` **and**
  `!IsExpired(now)`. Never trust `Status` alone for liveness: an `Active`-but-past-expiry
  session is *not* active. Conversely `Rotate` guards on `IsExpired(now)` (live), not on
  `status == Expired`, so a lagging status marker can never let a stale session rotate.

### Events

File: `session/events.go`. Past-tense facts, each name in a **named constant** returned
by `EventName()` (no magic strings), each an immutable pure object carrying **ids only —
the session id and the account id, never the raw secret or its hash**, implementing the
context's `Event` marker. Recorded via the promoted `Record` from the base `Entity`.

| Event | Recorded by | Payload |
|---|---|---|
| `SessionOpenedEvent` | `Open` | session id, account id |
| `SessionRotatedEvent` | `Rotate` | session id, account id |
| `SessionRevokedEvent` | `Revoke` | session id, account id |

```go
const sessionRotatedEventName = "identity.session.rotated"

// SessionRotatedEvent is recorded when a refresh session's secret is rotated in place.
// It carries ids only — never the secret or its hash.
type SessionRotatedEvent struct {
	sessionID ID
	accountID account.ID
}

func NewSessionRotatedEvent(sessionID ID, accountID account.ID) SessionRotatedEvent {
	return SessionRotatedEvent{sessionID: sessionID, accountID: accountID}
}

func (e SessionRotatedEvent) SessionID() ID          { return e.sessionID }
func (e SessionRotatedEvent) AccountID() account.ID  { return e.accountID }
func (e SessionRotatedEvent) EventName() string      { return sessionRotatedEventName }
```

`SessionOpenedEvent` and `SessionRevokedEvent` follow the same shape (own name constant, same
id-only payload).

### Errors

File: `session/errors.go`. Business/invariant violations only, one factory per failure,
each returning `*domain.DomainError`, Go-style `New<Concept>Error` names (the `session`
package carries the concept). "Not found" / "already exists" are **not** here — they
belong at the repository/application boundary.

| Factory | Raised when |
|---|---|
| `NewEmptyIDError()` | `ID` is zero at `Open` (also reused by `ID`'s constructor) |
| `NewIDTooLongError(length int)` | `ID` exceeds `maxIDLength` |
| `NewEmptyAccountIDError()` | `AccountID` is zero at `Open` |
| `NewEmptySecretHashError()` | `SecretHash` is empty (constructor / `Open`) |
| `NewSecretHashTooLongError(length int)` | `SecretHash` exceeds `maxSecretHashLength` |
| `NewZeroExpiresAtError()` | `ExpiresAt` is the zero time (constructor / `Open`) |
| `NewCannotRotateNonRefreshError()` | `Rotate` called on a non-`Refresh` (`ServerSide`) session |
| `NewSessionNotActiveError()` | `Rotate` called while `status != Active` |
| `NewExpiredError()` | `Rotate` called on a session past its expiry |
| `NewAlreadyRevokedError()` | `Revoke` called on an already-`Revoked` session |

```go
// session/errors.go
package session

func NewCannotRotateNonRefreshError() *domain.DomainError {
	return &domain.DomainError{Message: "only a refresh session may be rotated"}
}

func NewSessionNotActiveError() *domain.DomainError {
	return &domain.DomainError{Message: "session is not active"}
}

func NewExpiredError() *domain.DomainError {
	return &domain.DomainError{Message: "session has expired"}
}

func NewAlreadyRevokedError() *domain.DomainError {
	return &domain.DomainError{Message: "session is already revoked"}
}
```

### Repository port

File: `session/repository.go`. An interface declared by the domain, implemented in
`infrastructure`, with **persistence-oriented method names** (not domain verbs) and
`context.Context` on every method. It adds session-specific lookups
(`GetBySecretHash`, `ListByAccount`) and the bulk `RevokeAllByAccount` on top of the
standard CRUD set.

```go
// package session — the package carries the concept, so no stutter
type Repository interface {
	Create(ctx context.Context, session *Session) (*Session, error)
	Get(ctx context.Context, id ID) (*Session, error)
	Exists(ctx context.Context, id ID) (bool, error)
	Update(ctx context.Context, session *Session) error
	Delete(ctx context.Context, id ID) error

	// GetBySecretHash resolves the session presenting a given secret hash — the
	// lookup behind refresh/rotation and reuse detection.
	GetBySecretHash(ctx context.Context, secretHash SecretHash) (*Session, error)
	// ListByAccount returns every session for an account (e.g. to enumerate devices).
	ListByAccount(ctx context.Context, accountID account.ID) ([]*Session, error)
	// RevokeAllByAccount backs "log out everywhere": the application loads each
	// session and calls Revoke() within its unit of work.
	RevokeAllByAccount(ctx context.Context, accountID account.ID) error
}
```

- `GetBySecretHash` reflects the mutate-in-place model: because rotation swaps the hash
  on the same row, the *current* secret resolves to the one live session; the *previous*
  (superseded) secret no longer resolves — which is exactly how the application detects a
  replay and triggers `Revoke()`.
- `Update` is the primary persistence verb for this entity — `Rotate` and `Revoke` both
  mutate the existing row, so the application persists via `Update`, not `Create`.

---

## Rotation & reuse detection (design note)

- **One row per login.** A session is opened once (`Open`) and lives for the whole
  login. `Rotate` **swaps the secret hash on that same row**; there is no chain of
  one-shot tokens. A Session *is* the login on one device.
- **The entity supplies the primitives.** It exposes `Status()`, `IsExpired(now)`, and
  `IsActive(now)`, and enforces the transitions (`Rotate` guarded to `Kind == Refresh`
  + active + unexpired; `Revoke` guarded against double-revoke). It does **not** decide
  policy or detect replays — that is orchestration.
- **Reuse detection is per device.** The application resolves the presented secret via
  `GetBySecretHash`. A *current* secret rotates normally. A *superseded* secret
  (previously rotated away) no longer resolves to a live session — a replay — so the
  application calls `Revoke()` on **that one session**, cutting off the compromised
  device without touching the account's other sessions.
- **Log out everywhere is explicit and account-wide.** It is *not* a session command; it
  is the application driving `RevokeAllByAccount` (loading each session and calling
  `Revoke()`), used on password change, account suspension, or a user "sign out of all
  devices" action.
