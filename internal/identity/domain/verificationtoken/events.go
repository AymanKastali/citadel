package verificationtoken

import "github.com/AymanKastali/citadel/internal/identity/domain/account"

// Event wire names: the stable, transport-facing identifier each token event
// reports from EventName(). Grouped up top as one block rather than scattered one
// per event. Both events carry ids only — the token id, the owning account id, and
// the purpose — never the raw secret or the hash.
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

func (event TokenIssuedEvent) TokenID() ID           { return event.tokenID }
func (event TokenIssuedEvent) AccountID() account.ID { return event.accountID }
func (event TokenIssuedEvent) Purpose() Purpose      { return event.purpose }
func (event TokenIssuedEvent) EventName() string     { return tokenIssuedEventName }

// TokenConsumedEvent is recorded when a token is spent.
type TokenConsumedEvent struct {
	tokenID   ID
	accountID account.ID
	purpose   Purpose
}

func NewTokenConsumedEvent(tokenID ID, accountID account.ID, purpose Purpose) TokenConsumedEvent {
	return TokenConsumedEvent{tokenID: tokenID, accountID: accountID, purpose: purpose}
}

func (event TokenConsumedEvent) TokenID() ID           { return event.tokenID }
func (event TokenConsumedEvent) AccountID() account.ID { return event.accountID }
func (event TokenConsumedEvent) Purpose() Purpose      { return event.purpose }
func (event TokenConsumedEvent) EventName() string     { return tokenConsumedEventName }
