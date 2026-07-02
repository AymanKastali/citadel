package session

import (
	"context"

	"github.com/AymanKastali/citadel/internal/identity/domain/account"
)

// Repository is the session's repository port. The domain declares the interface;
// the concrete implementation lives in infrastructure. Its methods use persistence
// verbs (not domain verbs), take a context.Context and typed value objects, and
// obey CQS — Exists / Get / GetBySecretHash / ListByAccount return data, the
// mutating verbs return only the outcome. Update is the primary persistence verb
// for this entity: Rotate and Revoke both mutate the existing row, so the
// application persists via Update, not Create.
//
// Beyond the standard CRUD set it adds session-specific lookups. GetBySecretHash
// reflects the mutate-in-place model: because rotation swaps the hash on the same
// row, the current secret resolves to the one live session while a superseded
// secret no longer resolves — which is exactly how the application detects a replay
// and triggers Revoke(). RevokeAllByAccount backs "log out everywhere": the
// application loads each session and calls Revoke() within its unit of work. A
// "not found" from Get or GetBySecretHash is not a domain error; it is a repository
// outcome the application maps at the boundary.
type Repository interface {
	Create(ctx context.Context, session *Session) (*Session, error)
	Get(ctx context.Context, id ID) (*Session, error)
	Exists(ctx context.Context, id ID) (bool, error)
	Update(ctx context.Context, session *Session) error
	Delete(ctx context.Context, id ID) error
	GetBySecretHash(ctx context.Context, secretHash SecretHash) (*Session, error)
	ListByAccount(ctx context.Context, accountID account.ID) ([]*Session, error)
	RevokeAllByAccount(ctx context.Context, accountID account.ID) error
}
