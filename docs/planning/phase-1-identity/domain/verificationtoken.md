# Identity · Domain · `verificationtoken`

> **Status: complete — pieces (a) value objects, (b) entity + constructors, (c) commands, and (d) queries, events, errors & repository port all done.**
> Full deep-dive for the `verificationtoken` entity. Part of the Identity domain layer;
> see the [overview](../domain-layer.md) (decision #2 — one generic token told apart by
> `Purpose`), the rules in [`domain-layer.md`](../../../architecture/domain-layer.md),
> and the sibling [`account`](./account.md) it references.

## What `verificationtoken` is

A **one-shot, short-lived token** issued against an [`account`](./account.md) to prove
control of a channel. A single generic entity covers **both** flows — it is told apart
by a `Purpose` discriminator (overview decision #2):

- `EmailVerification` — confirms the account owns its email; consuming it moves the
  account from `PendingVerification` to `Active`.
- `PasswordReset` — authorizes an out-of-band password change; consuming it lets the
  reset proceed.

The token stores **only the hash** of the secret it represents — the raw secret is
generated, sent, and later compared **in infrastructure** and **never enters the
domain** (mirrors [`account.PasswordHash`](./account.md#passwordhash--password_hashgo-new)).
It is single-use (`Consume` marks it spent) and time-bound (`ExpiresAt`). Package:
`internal/identity/domain/verificationtoken`.

## Build order (bottom-up)

- **(a) Value objects** — `ID`, `Purpose`, `TokenHash`, `ExpiresAt` ✅
- **(b) Entity struct + `Issue` + `Reconstitute`** ✅
- **(c) Commands** — `Consume` ✅
- **(d) Queries, events, errors, repository port** ✅ ← *this checkpoint*

---

## (a) Value objects

One file per value object in the `verificationtoken` package. Each is immutable
(unexported field), self-validating in its constructor, with `Value()` /
`Equal(other)` / `IsZero()`, and a zero value that is invalid. Business/invariant
failures return `*domain.DomainError` via a factory in `verificationtoken/errors.go`
(piece **d**).

### `ID` — `id.go` (mirrors `account.ID`)

Structurally identical to [`account.ID`](./account.md) — an opaque identifier
(UUID/ULID) assigned by the application, not generated in the domain.

- **Wraps:** `string`.
- **Constructor:** `NewID(raw string) (ID, error)`.
- **Rules:** trim surrounding whitespace; reject empty (`NewEmptyIDError`); reject
  length `> maxIDLength` where `maxIDLength = 64` (`NewIDTooLongError(length)`).
- **Methods:** `Value() string`, `Equal(other ID) bool`, `IsZero() bool`.

### `Purpose` — `purpose.go` (enum, set by the entity)

Which flow the token belongs to. Like [`account.Status`](./account.md#status--statusgo-enum-mirrors-the-example--extends),
`Purpose` is an **enumerated type set by the entity at issue time**, not parsed from
raw external input — the caller passes an already-chosen `Purpose` value into
`Issue`. It therefore has **no `New…` constructor**. Persistence ↔ enum mapping is an
infrastructure concern.

- **Type:** `type Purpose int`, values via `iota`:
  - `EmailVerification` — confirms the account's email address.
  - `PasswordReset` — authorizes a password reset.
- **Zero value is invalid.** The first `iota` value (`0`) is **not** a valid purpose;
  reserve `0` as an unset sentinel so a zero-value field is caught at `Issue`. Declare
  the real values from `1`:

```go
type Purpose int

const (
	purposeUnset Purpose = iota // 0 — invalid sentinel; a zero-value Purpose means "unset"
	EmailVerification
	PasswordReset
)

func (purpose Purpose) IsZero() bool { return purpose == purposeUnset }
```

- **Suggested helper:** `String() string` for logging/mapping (no behavior change).
- **Note:** because `Purpose` is entity-set rather than parsed, `Issue` guards it with
  `IsZero()` (a caller that forgot to set it) and, if desired, an `IsValid()` range
  check; there is no untrusted-string path to reject.

### `TokenHash` — `token_hash.go` (mirrors `account.PasswordHash`)

Holds an **already-hashed** secret. The raw token secret **never enters the domain**:
infrastructure generates the secret, delivers it to the user, hashes it for storage,
and later hashes the presented secret to compare. This VO only stores and structurally
compares the hash; it is **algorithm-agnostic** (SHA-256, HMAC, argon2id — the domain
neither knows nor checks). This is the same reasoning as
[`account.PasswordHash`](./account.md#passwordhash--password_hashgo-new).

- **Wraps:** `string` (the encoded hash).
- **Constructor:** `NewTokenHash(raw string) (TokenHash, error)`.
- **Rules:** reject empty (`NewEmptyTokenHashError`); reject length
  `> maxTokenHashLength` where `maxTokenHashLength = 255` (comfortably fits any encoded
  hash while rejecting junk) (`NewTokenHashTooLongError(length)`). **No trimming, no
  normalization** — hash encodings are exact and case-sensitive. **No
  format/algorithm check** — that would couple the domain to a hashing scheme.
- **Methods:** `Value() string`, `Equal(other TokenHash) bool`, `IsZero() bool`.
- **CQS/security note:** `Equal` is a plain structural comparison of two stored hashes;
  it is **not** token verification. Comparing a presented secret against the stored hash
  is infrastructure's job (`GetByHash` looks a token up by its already-hashed value),
  never here. The raw secret is never a field, argument, or event payload in this
  package.

### `ExpiresAt` — `expires_at.go` (wraps `time.Time`)

The instant after which the token is no longer usable. Wraps a stdlib `time.Time`
(stdlib `time` is allowed in the domain; third-party packages are not). The domain does
**no I/O and reads no clock** — "now" is passed into the commands/queries that need it
(piece **c**), never fetched here.

- **Wraps:** `time.Time`.
- **Constructor:** `NewExpiresAt(raw time.Time) (ExpiresAt, error)`.
- **Rules:** reject the zero time (`raw.IsZero()` → `NewZeroExpiresAtError`). No other
  bound — whether the instant is in the past is a *runtime* question answered by
  `IsExpired(now)`, not a construction invariant (a token may legitimately be
  reconstituted after it has expired).
- **Methods:** `Value() time.Time`, `Equal(other ExpiresAt) bool`, `IsZero() bool`.
  `Equal` compares via `time.Time.Equal` (never `==`, which also compares monotonic
  clock/location).

### Cross-entity reference — `account.ID`

The token references its owning account by the **typed `account.ID` value object**,
imported from the `account` package — never an embedded `*Account` (overview
relationship map: `Account 1 ─ * VerificationToken`). It is carried on the entity and
its events, never dereferenced into the account itself.

---

## (b) Entity & constructors

File: `verificationtoken/verificationtoken.go`.

### Struct

The entity embeds the base `domain.Entity[ID]` (id + recorded events) and holds its
state as value objects and one typed cross-entity id. A **`consumedAt *time.Time`**
records the single-use marker: `nil` means unconsumed, a non-nil pointer records **when**
it was consumed. A `*time.Time` is chosen over a bare `consumed bool` deliberately —
both answer "is it spent?", but the timestamp also records *when*, which is useful for
auditing and reasoning, and it maps cleanly to a nullable timestamp column.

```go
type VerificationToken struct {
	domain.Entity[ID]
	accountID  account.ID
	purpose    Purpose
	hash       TokenHash
	expiresAt  ExpiresAt
	consumedAt *time.Time // nil = unconsumed; non-nil records when it was consumed
}
```

State changes only through the methods below; callers read via getters (piece **d**),
never by assigning fields.

### `Issue` — the business constructor

Builds a **new** token from already-valid value objects. Each field is guard-claused
(Fail Fast); the token starts **unconsumed** (`consumedAt == nil`) and records
`TokenIssuedEvent` (payload defined in piece **d**). No policy is involved — `Purpose`,
`ExpiresAt`, and the hash are decided by the caller (the application, from config-driven
TTL and the flow it is running) and passed in as ready value objects.

```go
type IssueParams struct {
	ID        ID
	AccountID account.ID
	Purpose   Purpose
	Hash      TokenHash
	ExpiresAt ExpiresAt
}

func Issue(params IssueParams) (*VerificationToken, error) {
	idIsMissing := params.ID.IsZero()
	if idIsMissing {
		return nil, NewEmptyIDError()
	}
	accountIsMissing := params.AccountID.IsZero()
	if accountIsMissing {
		return nil, NewEmptyAccountIDError()
	}
	purposeIsMissing := params.Purpose.IsZero()
	if purposeIsMissing {
		return nil, NewEmptyPurposeError()
	}
	hashIsMissing := params.Hash.IsZero()
	if hashIsMissing {
		return nil, NewEmptyTokenHashError()
	}
	expiresAtIsMissing := params.ExpiresAt.IsZero()
	if expiresAtIsMissing {
		return nil, NewZeroExpiresAtError()
	}
	token := &VerificationToken{
		Entity:     domain.NewEntity(params.ID),
		accountID:  params.AccountID,
		purpose:    params.Purpose,
		hash:       params.Hash,
		expiresAt:  params.ExpiresAt,
		consumedAt: nil, // starts unconsumed
	}
	token.Record(NewTokenIssuedEvent(token.ID(), token.accountID, token.purpose))
	return token, nil
}
```

### `Reconstitute` — rebuild from persistence

The counterpart used only by the repository adapter. It takes the **full persisted
state** — a superset of `IssueParams` that also carries the `ConsumedAt` marker — and
**just loads** it into a fresh entity: no validation, no event, no policy, no error (see
the [Reconstitution rule](../../../architecture/domain-layer.md#reconstitution-rebuilding-from-persistence)).
The stored `consumed_at` is nullable, so `ConsumedAt` is a `*time.Time` and is loaded
through as-is (a spent token rebuilds as spent).

```go
type ReconstituteParams struct {
	ID         ID
	AccountID  account.ID
	Purpose    Purpose
	Hash       TokenHash
	ExpiresAt  ExpiresAt
	ConsumedAt *time.Time
}

func Reconstitute(params ReconstituteParams) *VerificationToken {
	return &VerificationToken{
		Entity:     domain.NewEntity(params.ID),
		accountID:  params.AccountID,
		purpose:    params.Purpose,
		hash:       params.Hash,
		expiresAt:  params.ExpiresAt,
		consumedAt: params.ConsumedAt,
	}
}
```

## (c) Commands

File: `verificationtoken/verificationtoken.go` (alongside the constructors). A command
**mutates state and returns only `error`** (CQS) — never data. `now time.Time` is
**passed in** by the application; the domain reads no clock and does no I/O (stdlib
`time` is allowed only as a value type).

### `Consume(now time.Time) error`

Spends the token exactly once. Guards run **before** any mutation (Fail Fast):

1. **Already consumed** (`consumedAt != nil`) → `NewAlreadyConsumedError()`. A one-shot
   token cannot be spent twice.
2. **Expired** — `now` is at or after `expiresAt` (`!now.Before(expiresAt.Value())`) →
   `NewExpiredError()`. An expired token is unusable even if never consumed.

Only if both pass does it mutate: record the consumption instant and record
`TokenConsumedEvent`. The whole-secret comparison that authorizes *reaching* this call is an
infrastructure/application concern — by the time `Consume` runs, the matching token has
already been located by its hash.

```go
func (token *VerificationToken) Consume(now time.Time) error {
	alreadyConsumed := token.consumedAt != nil
	if alreadyConsumed {
		return NewAlreadyConsumedError()
	}
	expired := token.IsExpired(now)
	if expired {
		return NewExpiredError()
	}
	consumedAt := now
	token.consumedAt = &consumedAt
	token.Record(NewTokenConsumedEvent(token.ID(), token.accountID, token.purpose))
	return nil
}
```

## (d) Queries, events, errors, repository

### Queries (no mutation — CQS)

Read-only accessors, promoted `ID()` from the base plus token-specific getters. None
mutate; the two time-aware queries take `now` in (the domain reads no clock).

```go
func (token *VerificationToken) AccountID() account.ID { return token.accountID }
func (token *VerificationToken) Purpose() Purpose       { return token.purpose }

// IsExpired reports whether now is at or after the expiry instant.
func (token *VerificationToken) IsExpired(now time.Time) bool {
	return !now.Before(token.expiresAt.Value())
}

// IsConsumed reports whether the token has already been spent.
func (token *VerificationToken) IsConsumed() bool { return token.consumedAt != nil }
```

- `AccountID()` returns the typed cross-entity id; `Purpose()` the discriminator.
- `IsExpired(now)` and `IsConsumed()` are the two runtime predicates the application
  uses (e.g. to choose an error message) and that `Consume` reuses internally.
- No getter exposes the hash for comparison beyond `Value()` on the VO, and **no getter
  ever exposes a raw secret** — none exists in this package.

### Events — `events.go`

Both events are **past-tense facts recorded by the entity** (only entities record
events), each with its name in a **named constant** (no magic strings), implementing the
context's `Event` marker. Each carries **ids only** — the token id, the account id, and
the purpose — and **never the raw secret or the hash**.

```go
package verificationtoken

const (
	tokenIssuedEventName   = "verificationtoken.issued"
	tokenConsumedEventName = "verificationtoken.consumed"
)

// TokenIssuedEvent is recorded when a token is issued against an account.
type TokenIssuedEvent struct {
	tokenID   ID
	accountID account.ID
	purpose   Purpose
}

func NewTokenIssuedEvent(tokenID ID, accountID account.ID, purpose Purpose) TokenIssuedEvent {
	return TokenIssuedEvent{tokenID: tokenID, accountID: accountID, purpose: purpose}
}

func (e TokenIssuedEvent) TokenID() ID           { return e.tokenID }
func (e TokenIssuedEvent) AccountID() account.ID { return e.accountID }
func (e TokenIssuedEvent) Purpose() Purpose      { return e.purpose }
func (e TokenIssuedEvent) EventName() string     { return tokenIssuedEventName }

// TokenConsumedEvent is recorded when a token is spent.
type TokenConsumedEvent struct {
	tokenID   ID
	accountID account.ID
	purpose   Purpose
}

func NewTokenConsumedEvent(tokenID ID, accountID account.ID, purpose Purpose) TokenConsumedEvent {
	return TokenConsumedEvent{tokenID: tokenID, accountID: accountID, purpose: purpose}
}

func (e TokenConsumedEvent) TokenID() ID           { return e.tokenID }
func (e TokenConsumedEvent) AccountID() account.ID { return e.accountID }
func (e TokenConsumedEvent) Purpose() Purpose      { return e.purpose }
func (e TokenConsumedEvent) EventName() string     { return tokenConsumedEventName }
```

### Errors — `errors.go`

Business-rule / invariant violations only, each a factory returning
`*domain.DomainError` (lookup/persistence outcomes are **not** domain errors — they
belong at the repository/application boundary). Go-style names; the package carries the
concept, so no stutter.

```go
package verificationtoken

// --- value-object / construction invariants ---

func NewEmptyIDError() *domain.DomainError {
	return &domain.DomainError{Message: "verification token id must not be empty"}
}

func NewIDTooLongError(length int) *domain.DomainError {
	return &domain.DomainError{Message: fmt.Sprintf("verification token id must be at most %d characters, got %d", maxIDLength, length)}
}

func NewEmptyAccountIDError() *domain.DomainError {
	return &domain.DomainError{Message: "verification token account id must not be empty"}
}

func NewEmptyPurposeError() *domain.DomainError {
	return &domain.DomainError{Message: "verification token purpose must be set"}
}

func NewEmptyTokenHashError() *domain.DomainError {
	return &domain.DomainError{Message: "verification token hash must not be empty"}
}

func NewTokenHashTooLongError(length int) *domain.DomainError {
	return &domain.DomainError{Message: fmt.Sprintf("verification token hash must be at most %d characters, got %d", maxTokenHashLength, length)}
}

func NewZeroExpiresAtError() *domain.DomainError {
	return &domain.DomainError{Message: "verification token expiry must not be the zero time"}
}

// --- command / lifecycle invariants ---

func NewAlreadyConsumedError() *domain.DomainError {
	return &domain.DomainError{Message: "verification token has already been consumed"}
}

func NewExpiredError() *domain.DomainError {
	return &domain.DomainError{Message: "verification token has expired"}
}
```

### Repository port — `repository.go`

Declared by the domain, implemented in `infrastructure`. **Persistence-oriented verbs**
(not domain verbs), each taking `context.Context`. Beyond the standard CRUD set this
adds two lookups the flows need: `GetByHash` (find the token a presented—already
hashed—secret maps to) and `DeleteByAccountAndPurpose` (invalidate any outstanding
tokens of one purpose for an account, e.g. before issuing a fresh one).

```go
// package verificationtoken — the package carries the concept, so no stutter
type Repository interface {
	Create(ctx context.Context, token *VerificationToken) (*VerificationToken, error)
	Get(ctx context.Context, id ID) (*VerificationToken, error)
	Exists(ctx context.Context, id ID) (bool, error)
	Update(ctx context.Context, token *VerificationToken) error
	Delete(ctx context.Context, id ID) error
	GetByHash(ctx context.Context, hash TokenHash) (*VerificationToken, error)
	DeleteByAccountAndPurpose(ctx context.Context, accountID account.ID, purpose Purpose) error
}
```
