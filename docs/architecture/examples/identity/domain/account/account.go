package account

import "github.com/AymanKastali/citadel/internal/identity/domain"

// Account is an entity: mutable, and the single owner of the rules that govern
// an account. State changes only through its methods, never by direct field
// assignment from outside the package. It embeds domain.Entity for its id and
// recorded events.
type Account struct {
	domain.Entity[ID]
	email  Email
	status Status
}

// RegisterParams groups the already-valid value objects an account is
// registered from, so Register takes one data argument instead of a positional
// list. The email-verification policy is NOT a field here — a policy is a
// behavioral dependency, not data, so it is passed alongside this struct.
type RegisterParams struct {
	ID    ID
	Email Email
}

// Register builds a new account, rejecting any missing (zero-value) field. The
// email-verification policy is passed alongside the params struct — never
// inside it — so the caller (wired at the composition root from config) chooses
// the strategy without the entity or the params knowing which one. The policy
// decides the starting status (a query); the account records what happened.
func Register(params RegisterParams, verification EmailVerificationPolicy) (*Account, error) {
	idIsMissing := params.ID.IsZero()
	if idIsMissing {
		return nil, NewEmptyIDError()
	}
	emailIsMissing := params.Email.IsZero()
	if emailIsMissing {
		return nil, NewEmptyEmailError()
	}
	account := &Account{
		Entity: domain.NewEntity(params.ID),
		email:  params.Email,
		status: verification.InitialStatus(), // the strategy decides; the entity accommodates
	}
	account.Record(NewAccountRegistered(account.ID(), account.status))
	return account, nil
}

func (account *Account) Email() Email { return account.email }

func (account *Account) Status() Status { return account.status }
