package account

const accountRegisteredEventName = "account.registered"

// AccountRegistered is recorded when an account is registered. It is an
// immutable domain event and carries the account's id and the status it started
// in, not the account itself.
type AccountRegistered struct {
	accountID ID
	status    Status
}

func NewAccountRegistered(accountID ID, status Status) AccountRegistered {
	return AccountRegistered{accountID: accountID, status: status}
}

func (event AccountRegistered) AccountID() ID { return event.accountID }

func (event AccountRegistered) Status() Status { return event.status }

func (event AccountRegistered) EventName() string { return accountRegisteredEventName }
