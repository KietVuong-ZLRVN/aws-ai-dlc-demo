package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/cart-handoff/domain"
)

// MySQLCartHandoffRepository implements domain.CartHandoffRecordRepository using Aurora MySQL.
type MySQLCartHandoffRepository struct {
	db *sql.DB
}

func NewMySQLCartHandoffRepository(db *sql.DB) *MySQLCartHandoffRepository {
	return &MySQLCartHandoffRepository{db: db}
}

func (r *MySQLCartHandoffRepository) Save(ctx context.Context, rec *domain.CartHandoffRecord) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	src := rec.Source()
	var sourceComboID *string
	if src.Type == domain.HandoffSourceTypeSavedCombo {
		sourceComboID = &src.ComboId
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO cart_handoff_records (id, shopper_id, source_type, source_combo_id, status, recorded_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		rec.ID().String(),
		rec.ShopperID().String(),
		string(src.Type),
		sourceComboID,
		string(rec.Status()),
		rec.RecordedAt(),
	)
	if err != nil {
		return fmt.Errorf("insert handoff record: %w", err)
	}

	for _, item := range rec.AddedItems() {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO handoff_record_items (record_id, simple_sku, outcome, skip_reason)
			VALUES (?, ?, 'added', NULL)`,
			rec.ID().String(), item.SimpleSku,
		)
		if err != nil {
			return fmt.Errorf("insert added item: %w", err)
		}
	}

	for _, item := range rec.SkippedItems() {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO handoff_record_items (record_id, simple_sku, outcome, skip_reason)
			VALUES (?, ?, 'skipped', ?)`,
			rec.ID().String(), item.SimpleSku, item.Reason,
		)
		if err != nil {
			return fmt.Errorf("insert skipped item: %w", err)
		}
	}

	return tx.Commit()
}

func (r *MySQLCartHandoffRepository) FindById(ctx context.Context, id domain.CartHandoffRecordId) (*domain.CartHandoffRecord, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT hr.id, hr.shopper_id, hr.source_type, hr.source_combo_id, hr.status, hr.recorded_at,
		       hi.simple_sku, hi.outcome, hi.skip_reason
		FROM cart_handoff_records hr
		LEFT JOIN handoff_record_items hi ON hi.record_id = hr.id
		WHERE hr.id = ?`, id.String())
	if err != nil {
		return nil, fmt.Errorf("query handoff record: %w", err)
	}
	defer rows.Close()

	recs, err := scanRecords(rows)
	if err != nil {
		return nil, err
	}
	if len(recs) == 0 {
		return nil, fmt.Errorf("record not found")
	}
	return recs[0], nil
}

func (r *MySQLCartHandoffRepository) FindByShopperId(ctx context.Context, shopperID domain.ShopperId) ([]*domain.CartHandoffRecord, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT hr.id, hr.shopper_id, hr.source_type, hr.source_combo_id, hr.status, hr.recorded_at,
		       hi.simple_sku, hi.outcome, hi.skip_reason
		FROM cart_handoff_records hr
		LEFT JOIN handoff_record_items hi ON hi.record_id = hr.id
		WHERE hr.shopper_id = ?
		ORDER BY hr.recorded_at DESC`, shopperID.String())
	if err != nil {
		return nil, fmt.Errorf("query handoff records by shopper: %w", err)
	}
	defer rows.Close()

	return scanRecords(rows)
}

type recordRow struct {
	id            string
	shopperID     string
	sourceType    string
	sourceComboID *string
	status        string
	recordedAt    time.Time
}

type itemRow struct {
	simpleSku  string
	outcome    string
	skipReason *string
}

func scanRecords(rows *sql.Rows) ([]*domain.CartHandoffRecord, error) {
	order := []string{}
	byID := make(map[string]*recordRow)
	itemsByID := make(map[string][]itemRow)

	for rows.Next() {
		var rr recordRow
		var simpleSku, outcome *string
		var skipReason *string

		if err := rows.Scan(&rr.id, &rr.shopperID, &rr.sourceType, &rr.sourceComboID, &rr.status, &rr.recordedAt,
			&simpleSku, &outcome, &skipReason); err != nil {
			return nil, fmt.Errorf("scan record row: %w", err)
		}

		if _, exists := byID[rr.id]; !exists {
			clone := rr
			byID[rr.id] = &clone
			order = append(order, rr.id)
		}

		if simpleSku != nil {
			ir := itemRow{simpleSku: *simpleSku, outcome: *outcome, skipReason: skipReason}
			itemsByID[rr.id] = append(itemsByID[rr.id], ir)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make([]*domain.CartHandoffRecord, 0, len(order))
	for _, id := range order {
		rr := byID[id]

		src := domain.HandoffSource{Type: domain.HandoffSourceType(rr.sourceType)}
		if rr.sourceComboID != nil {
			src.ComboId = *rr.sourceComboID
		}

		var added []domain.CartItem
		var skipped []domain.SkippedItem

		for _, ir := range itemsByID[id] {
			if ir.outcome == "added" {
				added = append(added, domain.CartItem{SimpleSku: ir.simpleSku, Quantity: 1})
			} else {
				reason := ""
				if ir.skipReason != nil {
					reason = *ir.skipReason
				}
				skipped = append(skipped, domain.SkippedItem{SimpleSku: ir.simpleSku, Reason: reason})
			}
		}

		rec := domain.ReconstitueCartHandoffRecord(
			domain.CartHandoffRecordId(rr.id),
			domain.ShopperId(rr.shopperID),
			src,
			domain.HandoffStatus(rr.status),
			added,
			skipped,
			rr.recordedAt.UTC(),
		)
		result = append(result, rec)
	}
	return result, nil
}
