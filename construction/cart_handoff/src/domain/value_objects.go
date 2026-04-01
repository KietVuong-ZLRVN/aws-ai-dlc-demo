package domain

import "time"

// HandoffStatus represents the outcome of a cart handoff attempt.
type HandoffStatus string

const (
	HandoffStatusOk      HandoffStatus = "ok"
	HandoffStatusPartial HandoffStatus = "partial"
	HandoffStatusFailed  HandoffStatus = "failed"
)

// CartHandoffRecordId is the unique identity of a CartHandoffRecord.
type CartHandoffRecordId string

func (id CartHandoffRecordId) String() string { return string(id) }

// ShopperId identifies the shopper who initiated the handoff.
type ShopperId string

func (id ShopperId) String() string { return string(id) }

// HandoffSourceType is the discriminant for HandoffSource.
type HandoffSourceType string

const (
	HandoffSourceTypeSavedCombo  HandoffSourceType = "saved_combo"
	HandoffSourceTypeInlineItems HandoffSourceType = "inline_items"
)

// HandoffSource records the origin of the items being handed off.
type HandoffSource struct {
	Type    HandoffSourceType
	ComboId string     // set when Type == HandoffSourceTypeSavedCombo
	Items   []CartItem // set when Type == HandoffSourceTypeInlineItems
}

// CartItem is a single item submitted (or attempted) in a bulk cart add.
type CartItem struct {
	SimpleSku string
	Quantity  int
	Size      string
}

// SkippedItem records why a specific item was not added to the cart.
type SkippedItem struct {
	SimpleSku string
	Reason    string // "out_of_stock" | "platform_error"
}

// HandoffTimestamp records when the handoff was recorded.
type HandoffTimestamp time.Time

func (t HandoffTimestamp) Time() time.Time { return time.Time(t) }
