package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CartHandoffRecord is the aggregate root for the Cart Handoff bounded context.
// It is immutable after creation — create-only, no mutation methods.
type CartHandoffRecord struct {
	id           CartHandoffRecordId
	shopperID    ShopperId
	source       HandoffSource
	status       HandoffStatus
	addedItems   []CartItem
	skippedItems []SkippedItem
	recordedAt   time.Time
	events       []DomainEvent
}

// NewCartHandoffRecord creates an immutable CartHandoffRecord, enforcing all invariants.
func NewCartHandoffRecord(
	shopperID ShopperId,
	source HandoffSource,
	addedItems []CartItem,
	skippedItems []SkippedItem,
) (*CartHandoffRecord, error) {
	if err := validateSource(source); err != nil {
		return nil, err
	}

	status := deriveStatus(addedItems, skippedItems)
	now := time.Now().UTC()
	id := CartHandoffRecordId(uuid.NewString())

	r := &CartHandoffRecord{
		id:           id,
		shopperID:    shopperID,
		source:       source,
		status:       status,
		addedItems:   addedItems,
		skippedItems: skippedItems,
		recordedAt:   now,
	}

	if status == HandoffStatusFailed {
		reason := "platform_error"
		if len(skippedItems) > 0 {
			reason = skippedItems[0].Reason
		}
		r.events = append(r.events, CartHandoffFailed{
			RecordId:      id.String(),
			ShopperId:     shopperID.String(),
			HandoffSource: source,
			FailureReason: reason,
			Timestamp:     now,
		})
	} else {
		r.events = append(r.events, CartHandoffRecorded{
			RecordId:         id.String(),
			ShopperId:        shopperID.String(),
			HandoffSource:    source,
			Status:           status,
			AddedItemCount:   len(addedItems),
			SkippedItemCount: len(skippedItems),
			Timestamp:        now,
		})
	}

	return r, nil
}

// ReconstitueCartHandoffRecord rebuilds a record from persistence without firing events.
func ReconstitueCartHandoffRecord(
	id CartHandoffRecordId,
	shopperID ShopperId,
	source HandoffSource,
	status HandoffStatus,
	addedItems []CartItem,
	skippedItems []SkippedItem,
	recordedAt time.Time,
) *CartHandoffRecord {
	return &CartHandoffRecord{
		id: id, shopperID: shopperID, source: source,
		status: status, addedItems: addedItems, skippedItems: skippedItems, recordedAt: recordedAt,
	}
}

// PopEvents returns pending domain events and clears the list.
func (r *CartHandoffRecord) PopEvents() []DomainEvent {
	evts := r.events
	r.events = nil
	return evts
}

// Getters
func (r *CartHandoffRecord) ID() CartHandoffRecordId     { return r.id }
func (r *CartHandoffRecord) ShopperID() ShopperId        { return r.shopperID }
func (r *CartHandoffRecord) Source() HandoffSource       { return r.source }
func (r *CartHandoffRecord) Status() HandoffStatus       { return r.status }
func (r *CartHandoffRecord) AddedItems() []CartItem      { return r.addedItems }
func (r *CartHandoffRecord) SkippedItems() []SkippedItem { return r.skippedItems }
func (r *CartHandoffRecord) RecordedAt() time.Time       { return r.recordedAt }

// deriveStatus determines the HandoffStatus from the added/skipped item lists.
func deriveStatus(added []CartItem, skipped []SkippedItem) HandoffStatus {
	switch {
	case len(added) == 0:
		return HandoffStatusFailed
	case len(skipped) > 0:
		return HandoffStatusPartial
	default:
		return HandoffStatusOk
	}
}

// validateSource ensures exactly one source variant is set.
func validateSource(s HandoffSource) error {
	switch s.Type {
	case HandoffSourceTypeSavedCombo:
		if s.ComboId == "" {
			return fmt.Errorf("%w: comboId is empty for saved_combo source", ErrInvalidHandoffSource)
		}
	case HandoffSourceTypeInlineItems:
		if len(s.Items) == 0 {
			return fmt.Errorf("%w: items list is empty for inline_items source", ErrInvalidHandoffSource)
		}
	default:
		return ErrInvalidHandoffSource
	}
	return nil
}
