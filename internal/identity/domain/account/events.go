package account

import "time"

// Event wire names: the stable, transport-facing identifier each account event
// reports from EventName(). Grouped up top as one block (the low-level naming
// detail kept together) rather than scattered one per event.
const (
	accountRegisteredEventName  = "account.registered"
	emailVerifiedEventName      = "account.email_verified"
	passwordChangedEventName    = "account.password_changed"
	accountSuspendedEventName   = "account.suspended"
	accountReactivatedEventName = "account.reactivated"
	accountDeactivatedEventName = "account.deactivated"
	accountLoggedInEventName    = "account.logged_in"
)

// AccountRegisteredEvent is recorded when an account is registered. It is an
// immutable domain event and carries the account's id and the status it started
// in, not the account itself.
type AccountRegisteredEvent struct {
	accountID ID
	status    Status
}

func NewAccountRegisteredEvent(accountID ID, status Status) AccountRegisteredEvent {
	return AccountRegisteredEvent{accountID: accountID, status: status}
}

func (event AccountRegisteredEvent) AccountID() ID { return event.accountID }

func (event AccountRegisteredEvent) Status() Status { return event.status }

func (event AccountRegisteredEvent) EventName() string { return accountRegisteredEventName }

// EmailVerifiedEvent is recorded by VerifyEmail. It carries the account's id.
type EmailVerifiedEvent struct {
	accountID ID
}

func NewEmailVerifiedEvent(accountID ID) EmailVerifiedEvent {
	return EmailVerifiedEvent{accountID: accountID}
}

func (event EmailVerifiedEvent) AccountID() ID { return event.accountID }

func (event EmailVerifiedEvent) EventName() string { return emailVerifiedEventName }

// PasswordChangedEvent is recorded by ChangePassword. It carries only the id; it
// never carries the old or new hash, so the fact is safe to dispatch and store.
type PasswordChangedEvent struct {
	accountID ID
}

func NewPasswordChangedEvent(accountID ID) PasswordChangedEvent {
	return PasswordChangedEvent{accountID: accountID}
}

func (event PasswordChangedEvent) AccountID() ID { return event.accountID }

func (event PasswordChangedEvent) EventName() string { return passwordChangedEventName }

// AccountSuspendedEvent is recorded by Suspend. It carries the account's id.
type AccountSuspendedEvent struct {
	accountID ID
}

func NewAccountSuspendedEvent(accountID ID) AccountSuspendedEvent {
	return AccountSuspendedEvent{accountID: accountID}
}

func (event AccountSuspendedEvent) AccountID() ID { return event.accountID }

func (event AccountSuspendedEvent) EventName() string { return accountSuspendedEventName }

// AccountReactivatedEvent is recorded by Reactivate. It carries the account's id.
type AccountReactivatedEvent struct {
	accountID ID
}

func NewAccountReactivatedEvent(accountID ID) AccountReactivatedEvent {
	return AccountReactivatedEvent{accountID: accountID}
}

func (event AccountReactivatedEvent) AccountID() ID { return event.accountID }

func (event AccountReactivatedEvent) EventName() string { return accountReactivatedEventName }

// AccountDeactivatedEvent is recorded by Deactivate. It carries the account's id.
type AccountDeactivatedEvent struct {
	accountID ID
}

func NewAccountDeactivatedEvent(accountID ID) AccountDeactivatedEvent {
	return AccountDeactivatedEvent{accountID: accountID}
}

func (event AccountDeactivatedEvent) AccountID() ID { return event.accountID }

func (event AccountDeactivatedEvent) EventName() string { return accountDeactivatedEventName }

// AccountLoggedInEvent is recorded by Login. It carries the account's id and the
// moment of the successful login (the one fact the event's name does not already
// imply).
type AccountLoggedInEvent struct {
	accountID ID
	at        time.Time
}

func NewAccountLoggedInEvent(accountID ID, at time.Time) AccountLoggedInEvent {
	return AccountLoggedInEvent{accountID: accountID, at: at}
}

func (event AccountLoggedInEvent) AccountID() ID { return event.accountID }

func (event AccountLoggedInEvent) At() time.Time { return event.at }

func (event AccountLoggedInEvent) EventName() string { return accountLoggedInEventName }
