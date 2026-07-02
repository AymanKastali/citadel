package verificationtoken

import (
	"context"

	"github.com/AymanKastali/citadel/internal/identity/domain/account"
)

// Repository is the verification token's repository port. The domain declares the
// interface; the concrete implementation lives in infrastructure. Its methods use
// persistence verbs (not domain verbs), take a context.Context and typed value
// objects, and obey CQS — Exists / Get / GetByHash return data, the mutating verbs
// return only the outcome. Beyond the standard CRUD set it adds two lookups the
// flows need: GetByHash finds the token a presented (already hashed) secret maps
// to, and DeleteByAccountAndPurpose invalidates any outstanding tokens of one
// purpose for an account (e.g. before issuing a fresh one). A "not found" from Get
// or GetByHash is not a domain error; it is a repository outcome the application
// maps at the boundary.
type Repository interface {
	Create(ctx context.Context, token *VerificationToken) (*VerificationToken, error)
	Get(ctx context.Context, id ID) (*VerificationToken, error)
	Exists(ctx context.Context, id ID) (bool, error)
	Update(ctx context.Context, token *VerificationToken) error
	Delete(ctx context.Context, id ID) error
	GetByHash(ctx context.Context, hash TokenHash) (*VerificationToken, error)
	DeleteByAccountAndPurpose(ctx context.Context, accountID account.ID, purpose Purpose) error
}
