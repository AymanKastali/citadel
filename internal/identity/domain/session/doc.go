// Package session is an authenticated login on one device: the persistent
// server-side record formerly called the "refresh token", renamed to name the
// thing itself (the token is merely the secret that proves possession of the
// session). It holds the owning account.ID, a Kind discriminator, the hash of its
// secret, a lifecycle Status, and an ExpiresAt. It references its owning account by
// typed account.ID only, never an embedded Account entity.
//
// A session is long-lived and mutated in place. Rotate swaps the stored secret
// hash on the same row (one Session = one login = one device); Revoke terminates
// it (terminal). Only the hash is stored — the raw secret is generated, delivered,
// hashed, and verified in infrastructure and never enters the domain (mirroring
// account.PasswordHash). The domain reads no clock: the current time is passed into
// Rotate, IsExpired, and IsActive by the application. The stored Status.Expired is
// a lazily-persisted marker; IsExpired(now) is the live source of truth for
// liveness, which is why IsActive(now) checks both.
//
// Reading order (front page first):
//   - session.go       the Session entity: Open/Reconstitute, Rotate/Revoke, queries
//   - id.go, kind.go,  the value objects the entity is built from
//     secret_hash.go, expires_at.go, status.go
//   - events.go        the past-tense domain events the entity records
//   - errors.go        the domain-error factories
//   - repository.go    the persistence port (implemented in infrastructure)
//
// The package is pure domain: it imports only the standard library, the identity
// domain root, and the account package (for the typed account.ID it references) —
// never a framework, driver, or transport.
package session
