// Package account is the authenticatable identity at the heart of the identity
// context: an email, a password hash, and a lifecycle status. It is the subject
// of registration, email verification, password change, suspension,
// reactivation, deactivation, and login.
//
// Lifecycle: a new account starts either PendingVerification (when email
// verification is required) or Active (when it is optional), as decided by the
// injected EmailVerificationPolicy. VerifyEmail moves PendingVerification ->
// Active. An operator may Suspend an Active account and Reactivate it back.
// Deactivate closes an account permanently (terminal). Login is allowed only
// while the account is Active.
//
// Reading order (front page first):
//   - account.go                     the Account entity: constructors, commands, queries
//   - email.go, id.go,               the value objects the entity is built from
//     password_hash.go, status.go
//   - email_verification_policy.go   the registration-time status strategy
//   - events.go                      the past-tense domain events the entity records
//   - errors.go                      the domain-error factories
//   - repository.go                  the persistence port (implemented in infrastructure)
//
// The package is pure domain: it imports only the standard library and the
// identity domain root, never a framework, driver, or transport.
package account
