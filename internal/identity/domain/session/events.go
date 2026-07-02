package session

import "github.com/AymanKastali/citadel/internal/identity/domain/account"

// Event wire names: the stable, transport-facing identifier each session event
// reports from EventName(). Grouped up top as one block rather than scattered one
// per event. Every event carries ids only — the session id and the owning account
// id — never the raw secret or its hash.
const (
	sessionOpenedEventName  = "identity.session.opened"
	sessionRotatedEventName = "identity.session.rotated"
	sessionRevokedEventName = "identity.session.revoked"
)

// SessionOpenedEvent is recorded when a session is opened against an account.
type SessionOpenedEvent struct {
	sessionID ID
	accountID account.ID
}

func NewSessionOpenedEvent(sessionID ID, accountID account.ID) SessionOpenedEvent {
	return SessionOpenedEvent{sessionID: sessionID, accountID: accountID}
}

func (event SessionOpenedEvent) SessionID() ID         { return event.sessionID }
func (event SessionOpenedEvent) AccountID() account.ID { return event.accountID }
func (event SessionOpenedEvent) EventName() string     { return sessionOpenedEventName }

// SessionRotatedEvent is recorded when a refresh session's secret is rotated in
// place. It carries ids only — never the secret or its hash.
type SessionRotatedEvent struct {
	sessionID ID
	accountID account.ID
}

func NewSessionRotatedEvent(sessionID ID, accountID account.ID) SessionRotatedEvent {
	return SessionRotatedEvent{sessionID: sessionID, accountID: accountID}
}

func (event SessionRotatedEvent) SessionID() ID         { return event.sessionID }
func (event SessionRotatedEvent) AccountID() account.ID { return event.accountID }
func (event SessionRotatedEvent) EventName() string     { return sessionRotatedEventName }

// SessionRevokedEvent is recorded when a session is revoked (logout, reuse
// detection, or an operator action).
type SessionRevokedEvent struct {
	sessionID ID
	accountID account.ID
}

func NewSessionRevokedEvent(sessionID ID, accountID account.ID) SessionRevokedEvent {
	return SessionRevokedEvent{sessionID: sessionID, accountID: accountID}
}

func (event SessionRevokedEvent) SessionID() ID         { return event.sessionID }
func (event SessionRevokedEvent) AccountID() account.ID { return event.accountID }
func (event SessionRevokedEvent) EventName() string     { return sessionRevokedEventName }
