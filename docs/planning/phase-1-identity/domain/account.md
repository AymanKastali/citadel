# Identity · Domain · `account`

> **Status: complete — pieces (a) value objects, (b) entity + constructors, (c) commands, and (d) queries, events, errors & repository port all done.**
> Full deep-dive for the `account` entity. Part of the Identity domain layer; see the
> [overview](../domain-layer.md), the rules in
> [`domain-layer.md`](../../../architecture/domain-layer.md), and the worked
> [`examples/identity/`](../../../architecture/examples/identity/) (the example *is* an
> account, so several parts are reused verbatim).

## What `account` is

The **authenticatable identity**: an email, a password hash, and a lifecycle status.
It is the subject of registration, email verification, password reset/change,
suspension/deactivation, and it **records each successful authentication** (its last
login). It does **not** hold an organization or roles — those live on
[`membership`](./membership.md). Package: `internal/identity/domain/account`.

## Build order (bottom-up)

- **(a) Value objects** — `ID`, `Email`, `PasswordHash`, `Status` ✅
- **(b) Entity struct + `Register` + `Reconstitute`** ✅
- **(c) Commands** — `VerifyEmail`, `ChangePassword`, `Suspend`, `Reactivate`, `Deactivate`, `Login` ✅
- **(d) Queries, events, errors, repository port** ✅ ← *this checkpoint*

---

## (a) Value objects

One file per value object in the `account` package. Each is immutable (unexported
field), self-validating in its constructor, with `Value()` / `Equal(other)` /
`IsZero()`, and a zero value that is invalid. Business/invariant failures return
`*domain.DomainError` via a factory in `account/errors.go` (piece **d**).

### `ID` — `id.go` (reuse verbatim from the example)

Identical to
[`examples/identity/domain/account/id.go`](../../../architecture/examples/identity/domain/account/id.go).

- **Wraps:** `string` (an opaque identifier — UUID/ULID, assigned by the application, not generated in the domain).
- **Constructor:** `NewID(raw string) (ID, error)`.
- **Rules:** trim surrounding whitespace; reject empty (`NewEmptyIDError`); reject
  length `> maxIDLength` where `maxIDLength = 64` (`NewIDTooLongError(length)`).
- **Methods:** `Value() string`, `Equal(other ID) bool`, `IsZero() bool`.

### `Email` — `email.go` (reuse verbatim from the example)

Identical to
[`examples/identity/domain/account/email.go`](../../../architecture/examples/identity/domain/account/email.go).

- **Wraps:** `string`.
- **Constructor:** `NewEmail(raw string) (Email, error)`.
- **Rules:** trim; reject empty (`NewEmptyEmailError`); reject length `> maxEmailLength`
  where `maxEmailLength = 254` (RFC 5321 practical limit) (`NewEmailTooLongError`);
  validate against the **RFC 5322 address grammar** via the standard library's
  `net/mail.ParseAddress`, rejecting a value that fails to parse — or that carries a
  display name / is not a bare addr-spec — as malformed (`NewMalformedEmailError`);
  **normalize to lower case** so equality and uniqueness are stable.
- **Methods:** `Value() string`, `Equal(other Email) bool`, `IsZero() bool`.
- **Note:** validation uses the stdlib `net/mail` parser (no third-party dependency, so
  the domain stays pure) and accepts only a **bare address** — no display-name or
  angle-bracket forms. Syntax is all a value object can prove; deliverability is proven
  by the verification flow, not here.

### `PasswordHash` — `password_hash.go` (new)

Holds an **already-hashed** password. The raw plaintext password **never enters the
domain**: the application's password-hasher (an outbound port) hashes on the way in and
verifies on the way out. This VO only stores and structurally compares the hash; it is
**algorithm-agnostic** (bcrypt, argon2id, scrypt — the domain neither knows nor checks).

- **Wraps:** `string` (the encoded hash, e.g. a PHC/`$2b$…` string).
- **Constructor:** `NewPasswordHash(raw string) (PasswordHash, error)`.
- **Rules:** reject empty (`NewEmptyPasswordHashError`); reject length
  `> maxPasswordHashLength` where `maxPasswordHashLength = 255` (comfortably fits any
  encoded hash while rejecting junk) (`NewPasswordHashTooLongError(length)`). **No
  trimming, no normalization** — hash encodings are exact and case-sensitive. **No
  format/algorithm check** — that would couple the domain to a hashing scheme.
- **Methods:** `Value() string`, `Equal(other PasswordHash) bool`, `IsZero() bool`.
- **CQS/security note:** `Equal` is a plain structural comparison of two stored hashes
  (e.g. for change-detection); it is **not** password verification. Verifying a
  plaintext against the hash is the hasher port's job in the application, never here.
- **Authentication boundary:** password verification cannot live in the domain — it
  needs the hashing *algorithm* (a third-party import, forbidden here) and salted
  re-derivation, which is exactly why `Equal` is not verification. So **login is an
  application use case** (`Authenticate`): it loads the account by email, calls the
  outbound `PasswordHasher.Verify(raw, account.PasswordHash())` port, then — only if
  that passes — invokes the entity's own [`Login`](#login--record-a-successful-login-status-unchanged)
  command (eligibility guard + last-login) and `session.Open`. The entity owns *its*
  slice (is this identity eligible? record the login); the app owns the credential
  check and the orchestration. Two verbs, two layers: the app **authenticates** the
  credentials, the account records the **login**.

### `Status` — `status.go` (enum, mirrors the example + extends)

The account's place in its lifecycle. Unlike the constructor-validated VOs above,
`Status` is an **enumerated type set by the entity itself** (at construction via the
`EmailVerificationPolicy`, or by a transition command); it is never parsed from raw
external input, so it has no `New…` constructor. Persistence ↔ enum mapping is an
infrastructure concern.

- **Type:** `type Status int`, values via `iota`:
  - `PendingVerification` — exists but must confirm its email before it is fully
    usable. The **starting** status when email verification is *required*.
  - `Active` — usable. The starting status when verification is *not* required, and
    where a `PendingVerification` account lands after `VerifyEmail`.
  - `Suspended` — temporarily blocked by an operator; can be reactivated.
  - `Deactivated` — closed by the user or an operator; **terminal**.
- The two *registration-time* statuses (`PendingVerification`, `Active`) are the only
  ones the `EmailVerificationPolicy` may choose; the rest are reached by transition
  commands defined in piece **(c)**.
- **Helper:** `String() string` renders the status for logging / persistence mapping
  (implemented; no behavior change).
- Legal transitions are specified with the commands in piece **(c)**, not here.

---

## (b) Entity & constructors

File: `account/account.go`.

### Struct

The entity embeds the base `domain.Entity[ID]` (id + recorded events) and holds its
state as value objects only. It references no other entity — an account's organization
and roles live on [`membership`](./membership.md), not here.

```go
type Account struct {
	domain.Entity[ID]
	email        Email
	passwordHash PasswordHash
	status       Status
	lastLoginAt  *time.Time
}
```

`lastLoginAt` is the account's only timestamp: the moment of its **last successful
authentication**. It is a pointer so `nil` cleanly means *never logged in* — distinct
from the zero `time.Time` (mirroring `verificationtoken.consumedAt`). It is unset at
registration and moves only through [`Login`](#login--record-a-successful-login-status-unchanged).

State changes only through the methods below; callers read via getters (piece **d**),
never by assigning fields.

### `Register` — the business constructor

Builds a **new** account from already-valid value objects. The
`EmailVerificationPolicy` is passed **alongside** the params struct, never inside it (a
policy is a behavioral dependency, not data). Each field is guard-claused (Fail Fast);
the policy decides the starting `status` (`RequiredEmailVerification` →
`PendingVerification`, `OptionalEmailVerification` → `Active`); the account records
`AccountRegisteredEvent` (payload defined in piece **d**).

```go
type RegisterParams struct {
	ID           ID
	Email        Email
	PasswordHash PasswordHash
}

func Register(params RegisterParams, verification EmailVerificationPolicy) (*Account, error) {
	idIsMissing := params.ID.IsZero()
	if idIsMissing {
		return nil, NewEmptyIDError()
	}
	emailIsMissing := params.Email.IsZero()
	if emailIsMissing {
		return nil, NewEmptyEmailError()
	}
	passwordHashIsMissing := params.PasswordHash.IsZero()
	if passwordHashIsMissing {
		return nil, NewEmptyPasswordHashError()
	}
	account := &Account{
		Entity:       domain.NewEntity(params.ID),
		email:        params.Email,
		passwordHash: params.PasswordHash,
		status:       verification.InitialStatus(), // the strategy decides; the entity accommodates
	}
	account.Record(NewAccountRegisteredEvent(account.ID(), account.status))
	return account, nil
}
```

### `Reconstitute` — rebuild from persistence

The counterpart used only by the repository adapter. It takes the **full persisted
state** — a superset of `RegisterParams` that also carries `Status` — and **just loads**
it into a fresh entity: no validation, no event, no policy, no error (see the
[Reconstitution rule](../../../architecture/domain-layer.md#reconstitution-rebuilding-from-persistence)).

```go
type ReconstituteParams struct {
	ID           ID
	Email        Email
	PasswordHash PasswordHash
	Status       Status
	LastLoginAt  *time.Time
}

func Reconstitute(params ReconstituteParams) *Account {
	return &Account{
		Entity:       domain.NewEntity(params.ID),
		email:        params.Email,
		passwordHash: params.PasswordHash,
		status:       params.Status,
		lastLoginAt:  params.LastLoginAt,
	}
}
```

`LastLoginAt` is carried through as-is (`nil` for an account that has never logged in);
like the rest of `Reconstitute` it is a plain load — no validation, no event.

## (c) Commands

File: `account/account.go` (alongside `Register` / `Reconstitute`).

Every command below is a **command in the CQS sense**: it mutates the account and
returns **only `error`**, never data. Each one **guards the illegal transition first**
(Fail Fast — reject before touching state), applies the change, then **records a
past-tense event** through the promoted `Record` (payloads in piece **d**). Methods take
**value objects, not primitives** — validation already happened in the value object
constructors. Most commands need no `now` — a status/hash transition is a pure state
change plus a fact — with **one exception: `Login`**, which stamps
`lastLoginAt` and therefore takes the current time as a parameter (the domain never
reads the clock itself; the application passes it in).

Recording is **not** dispatching: the command only appends the event to the base
`Entity`; a feature pulls, dispatches, and drains it inside its unit of work (see the
[Domain events rule](../../../architecture/domain-layer.md#domain-events)).

### Legal transitions

`VerifyEmail`, `Suspend`, `Reactivate`, and `Deactivate` move the account between
statuses; `ChangePassword` and `Login` do not change status but are still gated by
it. The table is the single source of truth for what each command allows — anything not
listed is rejected with the named error.

| Command | Allowed **from** | Result **to** | Rejected from → error |
|---|---|---|---|
| `VerifyEmail` | `PendingVerification` | `Active` | `Active` → `NewAlreadyVerifiedError`; `Suspended` / `Deactivated` → `NewCannotVerifyError` |
| `ChangePassword` | `PendingVerification`, `Active`, `Suspended` | *(status unchanged)* | `Deactivated` → `NewCannotChangePasswordError` |
| `Suspend` | `Active` | `Suspended` | `Suspended` → `NewAlreadySuspendedError`; `PendingVerification` / `Deactivated` → `NewCannotSuspendError` |
| `Reactivate` | `Suspended` | `Active` | any non-`Suspended` → `NewNotSuspendedError` |
| `Deactivate` | `PendingVerification`, `Active`, `Suspended` | `Deactivated` | `Deactivated` → `NewAlreadyDeactivatedError` |
| `Login` | `Active` | *(status unchanged; stamps `lastLoginAt`)* | `PendingVerification` / `Suspended` / `Deactivated` → `NewCannotLoginError` |

`Deactivated` is **terminal** — no command leaves it. Every guard reads the current
`status`, so the entity never trusts the caller to have checked first.

### `VerifyEmail` — confirm the email, `PendingVerification` → `Active`

Only a `PendingVerification` account can be verified. An already-`Active` account is a
distinct, self-describing failure from one that can never be verified (suspended or
terminal), so the two guards raise different errors.

```go
func (account *Account) VerifyEmail() error {
	alreadyVerified := account.status == Active
	if alreadyVerified {
		return NewAlreadyVerifiedError()
	}
	cannotVerify := account.status != PendingVerification
	if cannotVerify {
		return NewCannotVerifyError()
	}
	account.status = Active
	account.Record(NewEmailVerifiedEvent(account.ID()))
	return nil
}
```

### `ChangePassword` — replace the stored hash

Takes an **already-hashed** `PasswordHash` value object (the plaintext never enters the
domain; the application's hasher port produced it). **Rejected on a `Deactivated`
account** — a terminal account is closed, so mutating its credentials is meaningless and
almost always a bug or a stale request; every other status may change its password
(including `PendingVerification`, e.g. a reset before first verification, and
`Suspended`, so a forced reset can accompany an operator action). Status is unchanged;
only the hash and the recorded fact move.

```go
func (account *Account) ChangePassword(newHash PasswordHash) error {
	cannotChangePassword := account.status == Deactivated
	if cannotChangePassword {
		return NewCannotChangePasswordError()
	}
	account.passwordHash = newHash
	account.Record(NewPasswordChangedEvent(account.ID()))
	return nil
}
```

### `Suspend` — operator block, `Active` → `Suspended`

Only an `Active` account can be suspended. Suspending an already-`Suspended` account and
suspending one that cannot be suspended at all (`PendingVerification` or the terminal
`Deactivated`) are separate failures.

```go
func (account *Account) Suspend() error {
	alreadySuspended := account.status == Suspended
	if alreadySuspended {
		return NewAlreadySuspendedError()
	}
	cannotSuspend := account.status != Active
	if cannotSuspend {
		return NewCannotSuspendError()
	}
	account.status = Suspended
	account.Record(NewAccountSuspendedEvent(account.ID()))
	return nil
}
```

### `Reactivate` — lift a suspension, `Suspended` → `Active`

The inverse of `Suspend`, and the only way back to `Active` from `Suspended`. Any other
starting status is a single failure.

```go
func (account *Account) Reactivate() error {
	notSuspended := account.status != Suspended
	if notSuspended {
		return NewNotSuspendedError()
	}
	account.status = Active
	account.Record(NewAccountReactivatedEvent(account.ID()))
	return nil
}
```

### `Deactivate` — close the account, → `Deactivated` (terminal)

Reachable from any non-terminal status; the only guard is against re-deactivating an
already-`Deactivated` account, which keeps the transition idempotent-by-rejection rather
than silently re-recording the fact.

```go
func (account *Account) Deactivate() error {
	alreadyDeactivated := account.status == Deactivated
	if alreadyDeactivated {
		return NewAlreadyDeactivatedError()
	}
	account.status = Deactivated
	account.Record(NewAccountDeactivatedEvent(account.ID()))
	return nil
}
```

### `Login` — record a successful login (status unchanged)

The account's slice of the login flow. It does **not** change status and does **not**
check the password — credential verification is the application's `PasswordHasher` port
(see the [authentication boundary](#passwordhash--password_hashgo-new) note). By the time
this is called, the app's `Authenticate` use case has already verified the password;
`Login` enforces the one rule the account itself owns — **only an `Active` account may log
in** — then stamps `lastLoginAt` and records the fact. The three non-`Active` statuses
collapse to a single `NewCannotLoginError` (as `NewCannotVerifyError` already does for
`VerifyEmail`); the application may branch on `Status()` for messaging (e.g.
`PendingVerification` → offer to resend the verification email). The guard lives **inside**
the command so it can never be skipped by a caller who jumps straight to recording a login.

```go
func (account *Account) Login(now time.Time) error {
	cannotLogin := account.status != Active
	if cannotLogin {
		return NewCannotLoginError()
	}
	account.lastLoginAt = &now
	account.Record(NewAccountLoggedInEvent(account.ID(), now))
	return nil
}
```

## (d) Queries, events, errors, repository

### Queries (getters)

File: `account/account.go`. Read-only accessors — the **query** half of CQS. They return
the current value object and **never mutate**; callers read through them instead of
touching fields. (`ID()` is already promoted from the embedded `domain.Entity[ID]`, so it
is not redeclared here.)

```go
func (account *Account) Email() Email { return account.email }

func (account *Account) PasswordHash() PasswordHash { return account.passwordHash }

func (account *Account) Status() Status { return account.status }

// LastLoginAt reports the last successful authentication. The bool is false when the
// account has never logged in (lastLoginAt is nil); returning a value + ok keeps the
// caller from dereferencing a pointer into the entity's internal state.
func (account *Account) LastLoginAt() (time.Time, bool) {
	neverLoggedIn := account.lastLoginAt == nil
	if neverLoggedIn {
		return time.Time{}, false
	}
	return *account.lastLoginAt, true
}
```

`PullEvents()` (copy — query) and `DrainEvents()` (clear — command) are likewise promoted
from the base `Entity`; the account does not re-declare them.

### Events — `account/events.go`

One file for every event the account records. Each is an **immutable, past-tense** domain
object that carries the **account id plus the relevant value** — never the entity itself —
and implements the context's `Event` marker via `EventName()`, which returns a **named
constant** (no inline literals). Only the account records them, from the commands above.

Every account event carries the **account id**; `AccountRegisteredEvent` additionally
carries the **starting status**, and `AccountLoggedInEvent` the login timestamp (`at`) —
the one fact its name does not imply. The five status-transition events carry just the id,
since the new status is implied by the event's name. The wire-name constants are **grouped
in one block** at the top of the file (not scattered one per event).

```go
const (
	accountRegisteredEventName  = "account.registered"
	emailVerifiedEventName      = "account.email_verified"
	passwordChangedEventName    = "account.password_changed"
	accountSuspendedEventName   = "account.suspended"
	accountReactivatedEventName = "account.reactivated"
	accountDeactivatedEventName = "account.deactivated"
	accountLoggedInEventName    = "account.logged_in"
)
```

```go
// AccountRegisteredEvent — recorded by Register; carries the id and the status the
// account started in.
type AccountRegisteredEvent struct {
	accountID ID
	status    Status
}

func NewAccountRegisteredEvent(accountID ID, status Status) AccountRegisteredEvent {
	return AccountRegisteredEvent{accountID: accountID, status: status}
}

func (event AccountRegisteredEvent) AccountID() ID     { return event.accountID }
func (event AccountRegisteredEvent) Status() Status    { return event.status }
func (event AccountRegisteredEvent) EventName() string { return accountRegisteredEventName }
```

The five status-transition events (`EmailVerifiedEvent`, `PasswordChangedEvent`,
`AccountSuspendedEvent`, `AccountReactivatedEvent`, `AccountDeactivatedEvent`) share the
same shape — the account id only — differing only in name; e.g.:

```go
// EmailVerifiedEvent — recorded by VerifyEmail. Carries the account's id.
type EmailVerifiedEvent struct {
	accountID ID
}

func NewEmailVerifiedEvent(accountID ID) EmailVerifiedEvent {
	return EmailVerifiedEvent{accountID: accountID}
}

func (event EmailVerifiedEvent) AccountID() ID     { return event.accountID }
func (event EmailVerifiedEvent) EventName() string { return emailVerifiedEventName }
```

`AccountLoggedInEvent` adds the login timestamp:

```go
// AccountLoggedInEvent — recorded by Login. Carries the account's id and the moment of the
// successful login (the one fact the event's name does not imply).
type AccountLoggedInEvent struct {
	accountID ID
	at        time.Time
}

func NewAccountLoggedInEvent(accountID ID, at time.Time) AccountLoggedInEvent {
	return AccountLoggedInEvent{accountID: accountID, at: at}
}

func (event AccountLoggedInEvent) AccountID() ID     { return event.accountID }
func (event AccountLoggedInEvent) At() time.Time     { return event.at }
func (event AccountLoggedInEvent) EventName() string { return accountLoggedInEventName }
```

### Errors — `account/errors.go`

Business-rule / invariant violations **only** — no not-found, no already-exists, no
persistence or timeout, no HTTP status or code (those live at the repository/application
boundary). **One `*domain.DomainError` per failure**, from a **self-describing factory**
that owns its message; callers pass only context values. Names are Go-style
`New<Concept>Error`, and the `account` package already carries the noun (so
`NewAlreadyVerifiedError`, not `NewAccountAlreadyVerifiedError`).

**Value-object failures (piece a)** — already listed with their VOs; the factories live
here:

| Factory | Message intent |
|---|---|
| `NewEmptyIDError()` | account id must not be empty |
| `NewIDTooLongError(length int)` | account id must not exceed `maxIDLength` (64) characters, got `length` |
| `NewEmptyEmailError()` | account email must not be empty |
| `NewEmailTooLongError(length int)` | account email must not exceed `maxEmailLength` (254) characters, got `length` |
| `NewMalformedEmailError(value string)` | account email is malformed: `value` (quoted via `%q`) |
| `NewEmptyPasswordHashError()` | account password hash must not be empty |
| `NewPasswordHashTooLongError(length int)` | account password hash must not exceed `maxPasswordHashLength` (255) characters, got `length` |

**Transition failures (piece c)** — raised by the commands' guard clauses:

| Factory | Raised by | Message intent |
|---|---|---|
| `NewAlreadyVerifiedError()` | `VerifyEmail` | account email is already verified |
| `NewCannotVerifyError()` | `VerifyEmail` | account cannot be verified from its current status |
| `NewCannotChangePasswordError()` | `ChangePassword` | account password cannot be changed on a deactivated account |
| `NewAlreadySuspendedError()` | `Suspend` | account is already suspended |
| `NewCannotSuspendError()` | `Suspend` | only an active account can be suspended |
| `NewNotSuspendedError()` | `Reactivate` | account is not suspended, so it cannot be reactivated |
| `NewAlreadyDeactivatedError()` | `Deactivate` | account is already deactivated |
| `NewCannotLoginError()` | `Login` | account cannot log in from its current status |

The VO factories mirror
[`examples/identity/domain/account/errors.go`](../../../architecture/examples/identity/domain/account/errors.go)
(the worked example is a full-parity account, so it carries the same set). The transition
factories follow the same shape — parameterless, message-only:

```go
func NewAlreadyVerifiedError() *domain.DomainError {
	return &domain.DomainError{Message: "account email is already verified"}
}

func NewCannotVerifyError() *domain.DomainError {
	return &domain.DomainError{Message: "account cannot be verified from its current status"}
}

func NewCannotChangePasswordError() *domain.DomainError {
	return &domain.DomainError{Message: "account password cannot be changed on a deactivated account"}
}

func NewAlreadySuspendedError() *domain.DomainError {
	return &domain.DomainError{Message: "account is already suspended"}
}

func NewCannotSuspendError() *domain.DomainError {
	return &domain.DomainError{Message: "only an active account can be suspended"}
}

func NewNotSuspendedError() *domain.DomainError {
	return &domain.DomainError{Message: "account is not suspended, so it cannot be reactivated"}
}

func NewAlreadyDeactivatedError() *domain.DomainError {
	return &domain.DomainError{Message: "account is already deactivated"}
}

func NewCannotLoginError() *domain.DomainError {
	return &domain.DomainError{Message: "account cannot log in from its current status"}
}
```

### Repository port — `account/repository.go`

The account declares its own **repository port** — an interface owned by the domain,
implemented in `infrastructure`, with **persistence verbs** (not domain verbs). The
`account` package carries the concept, so the name is plain `Repository` with no stutter.
Beyond the standard CRUD set, `account` adds two **email-oriented** lookups the
authentication and registration flows need (login is by email; registration must reject a
duplicate). Every method takes a `context.Context` and typed value objects, and each obeys
CQS — `Exists` / `ExistsByEmail` return data (a `bool`), the mutating verbs return only
the outcome.

```go
package account

import "context"

// Repository is the account's repository port. The domain declares the
// interface; the concrete implementation lives in infrastructure.
type Repository interface {
	Create(ctx context.Context, account *Account) (*Account, error)
	Get(ctx context.Context, id ID) (*Account, error)
	Exists(ctx context.Context, id ID) (bool, error)
	Update(ctx context.Context, account *Account) error
	Delete(ctx context.Context, id ID) error
	GetByEmail(ctx context.Context, email Email) (*Account, error)
	ExistsByEmail(ctx context.Context, email Email) (bool, error)
}
```

A "not found" from `Get` / `GetByEmail` is **not** a domain error — it is a repository
outcome the application maps at the boundary; the domain layer defines no such error.
