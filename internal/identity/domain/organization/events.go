package organization

// Event wire names: the stable, transport-facing identifier each organization
// event reports from EventName(). Grouped up top as one block rather than
// scattered one per event.
const (
	organizationCreatedEventName     = "organization.created"
	organizationRenamedEventName     = "organization.renamed"
	organizationSlugChangedEventName = "organization.slug_changed"
	organizationSuspendedEventName   = "organization.suspended"
	organizationActivatedEventName   = "organization.activated"
)

// OrganizationCreatedEvent is recorded when an organization is created. It carries
// the organization's id and the slug it was created with, not the organization
// itself.
type OrganizationCreatedEvent struct {
	organizationID ID
	slug           Slug
}

func NewOrganizationCreatedEvent(organizationID ID, slug Slug) OrganizationCreatedEvent {
	return OrganizationCreatedEvent{organizationID: organizationID, slug: slug}
}

func (event OrganizationCreatedEvent) OrganizationID() ID { return event.organizationID }

func (event OrganizationCreatedEvent) Slug() Slug { return event.slug }

func (event OrganizationCreatedEvent) EventName() string { return organizationCreatedEventName }

// OrganizationRenamedEvent is recorded by Rename. It carries the id and the new name.
type OrganizationRenamedEvent struct {
	organizationID ID
	name           Name
}

func NewOrganizationRenamedEvent(organizationID ID, name Name) OrganizationRenamedEvent {
	return OrganizationRenamedEvent{organizationID: organizationID, name: name}
}

func (event OrganizationRenamedEvent) OrganizationID() ID { return event.organizationID }

func (event OrganizationRenamedEvent) Name() Name { return event.name }

func (event OrganizationRenamedEvent) EventName() string { return organizationRenamedEventName }

// OrganizationSlugChangedEvent is recorded by ChangeSlug. It carries the id and the
// new slug.
type OrganizationSlugChangedEvent struct {
	organizationID ID
	slug           Slug
}

func NewOrganizationSlugChangedEvent(organizationID ID, slug Slug) OrganizationSlugChangedEvent {
	return OrganizationSlugChangedEvent{organizationID: organizationID, slug: slug}
}

func (event OrganizationSlugChangedEvent) OrganizationID() ID { return event.organizationID }

func (event OrganizationSlugChangedEvent) Slug() Slug { return event.slug }

func (event OrganizationSlugChangedEvent) EventName() string {
	return organizationSlugChangedEventName
}

// OrganizationSuspendedEvent is recorded by Suspend. It carries the id.
type OrganizationSuspendedEvent struct {
	organizationID ID
}

func NewOrganizationSuspendedEvent(organizationID ID) OrganizationSuspendedEvent {
	return OrganizationSuspendedEvent{organizationID: organizationID}
}

func (event OrganizationSuspendedEvent) OrganizationID() ID { return event.organizationID }

func (event OrganizationSuspendedEvent) EventName() string { return organizationSuspendedEventName }

// OrganizationActivatedEvent is recorded by Activate. It carries the id.
type OrganizationActivatedEvent struct {
	organizationID ID
}

func NewOrganizationActivatedEvent(organizationID ID) OrganizationActivatedEvent {
	return OrganizationActivatedEvent{organizationID: organizationID}
}

func (event OrganizationActivatedEvent) OrganizationID() ID { return event.organizationID }

func (event OrganizationActivatedEvent) EventName() string { return organizationActivatedEventName }
