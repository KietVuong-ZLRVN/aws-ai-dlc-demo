package domain

import "time"

// DomainEvent is the marker interface for all domain events.
type DomainEvent interface {
	EventName() string
}

// ComboCreated is raised when a new Combo is saved.
type ComboCreated struct {
	ComboId   string
	ShopperId string
	Name      string
	ItemCount int
	CreatedAt time.Time
}

func (e ComboCreated) EventName() string { return "ComboCreated" }

// ComboRenamed is raised when a Combo's name changes.
type ComboRenamed struct {
	ComboId   string
	ShopperId string
	OldName   string
	NewName   string
	RenamedAt time.Time
}

func (e ComboRenamed) EventName() string { return "ComboRenamed" }

// ComboDeleted is raised when a Combo is removed.
type ComboDeleted struct {
	ComboId   string
	ShopperId string
	DeletedAt time.Time
}

func (e ComboDeleted) EventName() string { return "ComboDeleted" }

// ComboShared is raised when a share token is generated and visibility set to public.
type ComboShared struct {
	ComboId    string
	ShopperId  string
	ShareToken string
	SharedAt   time.Time
}

func (e ComboShared) EventName() string { return "ComboShared" }

// ComboMadePrivate is raised when visibility is changed to private, revoking the share token.
type ComboMadePrivate struct {
	ComboId           string
	ShopperId         string
	RevokedShareToken string
	MadePrivateAt     time.Time
}

func (e ComboMadePrivate) EventName() string { return "ComboMadePrivate" }
