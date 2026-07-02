// Package domain holds the ordering context's domain model: entities, value
// objects, domain errors, and the cross-entity domain services that live at
// the domain root.
package domain

// DomainError is the single type for every domain failure — a violated
// business rule or invariant, and only that. It is told apart by the factory
// that built it and by its message, never by a per-failure type.
type DomainError struct {
	Message string // describes the failure on its own
	Err     error  // optional wrapped cause; nil when there is none
}

func (e *DomainError) Error() string { return e.Message }

func (e *DomainError) Unwrap() error { return e.Err }

// Event marks a recorded domain fact. By convention an event is named in the
// past tense — it describes something that has already happened.
type Event interface {
	EventName() string
}

// Entity is the base every domain entity embeds: it carries the entity's id and
// the domain events it has recorded. ID stays strongly typed (product.ID,
// order.ID) through the type parameter, so embedding does not weaken identity.
type Entity[ID comparable] struct {
	id     ID
	events []Event
}

func NewEntity[ID comparable](id ID) Entity[ID] { return Entity[ID]{id: id} }

func (entity *Entity[ID]) ID() ID { return entity.id }

// Record appends a domain event as the entity changes state. It is called only
// by the entity's own methods — recording is not dispatching.
func (entity *Entity[ID]) Record(event Event) { entity.events = append(entity.events, event) }

// PullEvents returns a copy of the recorded events (query — reads, never mutates).
func (entity *Entity[ID]) PullEvents() []Event {
	copied := make([]Event, len(entity.events))
	copy(copied, entity.events)
	return copied
}

// DrainEvents clears the recorded events (command — mutates, never returns).
func (entity *Entity[ID]) DrainEvents() { entity.events = nil }
