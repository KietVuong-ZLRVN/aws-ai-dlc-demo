package domain

import "time"

// DomainEvent is the marker interface for all domain events.
type DomainEvent interface {
	EventName() string
}

// CartHandoffRecorded is raised when a handoff completes (ok or partial).
type CartHandoffRecorded struct {
	RecordId         string
	ShopperId        string
	HandoffSource    HandoffSource
	Status           HandoffStatus
	AddedItemCount   int
	SkippedItemCount int
	Timestamp        time.Time
}

func (e CartHandoffRecorded) EventName() string { return "CartHandoffRecorded" }

// CartHandoffFailed is raised when the platform cart API rejects the entire request.
type CartHandoffFailed struct {
	RecordId      string
	ShopperId     string
	HandoffSource HandoffSource
	FailureReason string
	Timestamp     time.Time
}

func (e CartHandoffFailed) EventName() string { return "CartHandoffFailed" }
