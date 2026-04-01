package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/KietVuong-ZLRVN/aws-ai-dlc-demo/combo-portfolio/domain"
)

// MySQLComboRepository implements domain.ComboRepository using Aurora MySQL.
type MySQLComboRepository struct {
	db *sql.DB
}

func NewMySQLComboRepository(db *sql.DB) *MySQLComboRepository {
	return &MySQLComboRepository{db: db}
}

func (r *MySQLComboRepository) Save(ctx context.Context, combo *domain.Combo) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var shareToken *string
	if combo.ShareToken() != nil {
		s := combo.ShareToken().String()
		shareToken = &s
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO combos (id, shopper_id, name, visibility, share_token, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE name=VALUES(name), visibility=VALUES(visibility),
		  share_token=VALUES(share_token), updated_at=VALUES(updated_at)`,
		combo.ID().String(),
		combo.ShopperID().String(),
		combo.Name().String(),
		string(combo.Visibility()),
		shareToken,
		combo.CreatedAt(),
		combo.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("upsert combo: %w", err)
	}

	if _, err = tx.ExecContext(ctx, `DELETE FROM combo_items WHERE combo_id = ?`, combo.ID().String()); err != nil {
		return fmt.Errorf("delete old items: %w", err)
	}

	for i, item := range combo.Items() {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO combo_items (combo_id, config_sku, simple_sku, name, image_url, price, sort_order)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			combo.ID().String(),
			item.ConfigSku,
			item.SimpleSku,
			item.Name,
			item.ImageUrl,
			item.Price,
			i,
		)
		if err != nil {
			return fmt.Errorf("insert item: %w", err)
		}
	}

	return tx.Commit()
}

func (r *MySQLComboRepository) FindById(ctx context.Context, id domain.ComboId) (*domain.Combo, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id, c.shopper_id, c.name, c.visibility, c.share_token, c.created_at, c.updated_at,
		       ci.config_sku, ci.simple_sku, ci.name, ci.image_url, ci.price
		FROM combos c
		LEFT JOIN combo_items ci ON ci.combo_id = c.id
		WHERE c.id = ?
		ORDER BY ci.sort_order`, id.String())
	if err != nil {
		return nil, fmt.Errorf("query combo by id: %w", err)
	}
	defer rows.Close()

	combos, err := scanCombos(rows)
	if err != nil {
		return nil, err
	}
	if len(combos) == 0 {
		return nil, domain.ErrComboNotFound
	}
	return combos[0], nil
}

func (r *MySQLComboRepository) FindByShopperId(ctx context.Context, shopperID domain.ShopperId) ([]*domain.Combo, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id, c.shopper_id, c.name, c.visibility, c.share_token, c.created_at, c.updated_at,
		       ci.config_sku, ci.simple_sku, ci.name, ci.image_url, ci.price
		FROM combos c
		LEFT JOIN combo_items ci ON ci.combo_id = c.id
		WHERE c.shopper_id = ?
		ORDER BY c.created_at DESC, ci.sort_order`, shopperID.String())
	if err != nil {
		return nil, fmt.Errorf("query combos by shopper: %w", err)
	}
	defer rows.Close()

	return scanCombos(rows)
}

func (r *MySQLComboRepository) FindByShareToken(ctx context.Context, token domain.ShareToken) (*domain.Combo, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id, c.shopper_id, c.name, c.visibility, c.share_token, c.created_at, c.updated_at,
		       ci.config_sku, ci.simple_sku, ci.name, ci.image_url, ci.price
		FROM combos c
		LEFT JOIN combo_items ci ON ci.combo_id = c.id
		WHERE c.share_token = ?
		ORDER BY ci.sort_order`, token.String())
	if err != nil {
		return nil, fmt.Errorf("query combo by share token: %w", err)
	}
	defer rows.Close()

	combos, err := scanCombos(rows)
	if err != nil {
		return nil, err
	}
	if len(combos) == 0 {
		return nil, domain.ErrComboNotFound
	}
	return combos[0], nil
}

func (r *MySQLComboRepository) Delete(ctx context.Context, id domain.ComboId) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM combos WHERE id = ?`, id.String())
	if err != nil {
		return fmt.Errorf("delete combo: %w", err)
	}
	return nil
}

// scanCombos reads joined combo+item rows and de-duplicates by combo ID.
func scanCombos(rows *sql.Rows) ([]*domain.Combo, error) {
	type comboRow struct {
		id         string
		shopperID  string
		name       string
		visibility string
		shareToken *string
		createdAt  time.Time
		updatedAt  time.Time
		items      []domain.ComboItem
	}

	order := []string{}
	byID := make(map[string]*comboRow)

	for rows.Next() {
		var row comboRow
		var configSku, simpleSku, itemName, imageUrl *string
		var price *float64

		if err := rows.Scan(
			&row.id, &row.shopperID, &row.name, &row.visibility,
			&row.shareToken, &row.createdAt, &row.updatedAt,
			&configSku, &simpleSku, &itemName, &imageUrl, &price,
		); err != nil {
			return nil, fmt.Errorf("scan combo row: %w", err)
		}

		if _, exists := byID[row.id]; !exists {
			clone := row
			byID[row.id] = &clone
			order = append(order, row.id)
		}

		if configSku != nil {
			byID[row.id].items = append(byID[row.id].items, domain.ComboItem{
				ConfigSku: *configSku,
				SimpleSku: *simpleSku,
				Name:      *itemName,
				ImageUrl:  *imageUrl,
				Price:     *price,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make([]*domain.Combo, 0, len(order))
	for _, id := range order {
		row := byID[id]
		var shareToken *domain.ShareToken
		if row.shareToken != nil {
			t := domain.ShareToken(*row.shareToken)
			shareToken = &t
		}
		result = append(result, domain.ReconstitueCombo(
			domain.ComboId(row.id),
			domain.ShopperId(row.shopperID),
			domain.ComboName(row.name),
			row.items,
			domain.Visibility(row.visibility),
			shareToken,
			row.createdAt.UTC(),
			row.updatedAt.UTC(),
		))
	}
	return result, nil
}
