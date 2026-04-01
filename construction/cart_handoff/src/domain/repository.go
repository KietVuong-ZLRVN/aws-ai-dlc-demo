package domain

import (
	"context"
)

// CartHandoffRecordRepository is the persistence abstraction for CartHandoffRecord.
// Records are append-only — Save is insert-only, no updates.
type CartHandoffRecordRepository interface {
	Save(ctx context.Context, record *CartHandoffRecord) error
	FindById(ctx context.Context, id CartHandoffRecordId) (*CartHandoffRecord, error)
	FindByShopperId(ctx context.Context, shopperID ShopperId) ([]*CartHandoffRecord, error)
}
