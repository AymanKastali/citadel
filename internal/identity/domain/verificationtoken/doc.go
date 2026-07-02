// Package verificationtoken is a one-shot, short-lived token issued against an
// account to prove control of a channel. A single generic entity covers both flows
// — EmailVerification (confirms the account owns its email) and PasswordReset
// (authorizes an out-of-band password change) — told apart by its Purpose
// discriminator. It references its owning account by typed account.ID only, never
// an embedded Account entity.
//
// The token stores only the hash of the secret it represents: the raw secret is
// generated, delivered, and later compared in infrastructure and never enters the
// domain (mirroring account.PasswordHash). It is single-use — Consume marks it
// spent exactly once — and time-bound via ExpiresAt. The domain reads no clock: the
// current time is passed into Consume and IsExpired by the application.
//
// Reading order (front page first):
//   - verificationtoken.go   the entity: Issue/Reconstitute, the Consume command, queries
//   - id.go, purpose.go,     the value objects the entity is built from
//     token_hash.go, expires_at.go
//   - events.go              the past-tense domain events the entity records
//   - errors.go              the domain-error factories
//   - repository.go          the persistence port (implemented in infrastructure)
//
// The package is pure domain: it imports only the standard library, the identity
// domain root, and the account package (for the typed account.ID it references) —
// never a framework, driver, or transport.
package verificationtoken
