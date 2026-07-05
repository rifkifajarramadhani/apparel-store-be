package mysqladapter

import (
	"context"
	"time"

	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/asset"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/colourway"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/pagination"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/price"
	appsize "github.com/rifkifajarramadhani/golang-clean-architecture/internal/size"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/sku"
)

func (r *SKURepository) List(ctx context.Context, q sku.Query) (pagination.CursorPage[sku.SKU], error) {
	now := time.Now().UTC()
	sql := `
		SELECT
			s.public_id AS id,
			s.sku_code AS code,
			COALESCE(s.barcode, '') AS barcode,
			p.style_code AS product_id,
			c.public_id AS colourway_id,
			c.name AS colourway_name,
			c.hex_code,
			sz.public_id AS size_id,
			ss.code AS scale_code,
			sz.code AS size_code,
			sz.name AS size_name,
			sz.sort_order,
			s.on_hand,
			s.reserved,
			COALESCE(
				(
					SELECT sp.amount
					FROM prices AS sp
					WHERE
						sp.sku_id = s.id
						AND sp.currency = ?
						AND sp.archived_at IS NULL
						AND sp.valid_from <= ?
						AND (sp.valid_to IS NULL OR sp.valid_to > ?)
					ORDER BY sp.valid_from DESC
					LIMIT 1
				),
				(
					SELECT pp.amount
					FROM prices AS pp
					WHERE
						pp.product_id = s.product_id
						AND pp.currency = ?
						AND pp.archived_at IS NULL
						AND pp.valid_from <= ?
						AND (pp.valid_to IS NULL OR pp.valid_to > ?)
					ORDER BY pp.valid_from DESC
					LIMIT 1
				),
				0
			) AS amount
		FROM skus AS s
		JOIN products AS p ON p.id = s.product_id AND p.archived_at IS NULL
		JOIN colourways AS c ON c.id = s.colourway_id AND c.archived_at IS NULL
		JOIN sizes AS sz ON sz.id = s.size_id AND sz.archived_at IS NULL
		JOIN size_scales AS ss ON ss.id = sz.size_scale_id
		WHERE s.archived_at IS NULL`
	args := []any{q.Currency, now, now, q.Currency, now, now}
	if q.ProductID != "" {
		sql += " AND (p.public_id=? OR p.style_code=?)"
		args = append(args, q.ProductID, q.ProductID)
	}

	if q.ColourwayID != "" {
		sql += " AND c.public_id=?"
		args = append(args, q.ColourwayID)
	}

	if q.Cursor != "" {
		sql += " AND s.public_id>?"
		args = append(args, q.Cursor)
	}

	sql += " ORDER BY s.public_id LIMIT ?"
	args = append(args, q.Limit+1)

	var rows []skuRow
	if err := r.db.WithContext(ctx).Raw(sql, args...).Scan(&rows).Error; err != nil {
		return pagination.CursorPage[sku.SKU]{}, err
	}

	hasMore := len(rows) > q.Limit
	if hasMore {
		rows = rows[:q.Limit]
	}

	items := make([]sku.SKU, 0, len(rows))
	for _, row := range rows {
		items = append(items, sku.SKU{ID: row.ID, Code: row.Code, Barcode: row.Barcode, ProductID: row.ProductID, Colourway: colourway.Colourway{ID: row.ColourwayID, Name: row.ColourwayName, HexCode: row.HexCode}, Size: appsize.Size{ID: row.SizeID, ScaleCode: row.ScaleCode, Code: row.SizeCode, Name: row.SizeName, SortOrder: row.SortOrder}, Price: price.Money{Currency: q.Currency, Amount: row.Amount, CompareAtAmount: row.CompareAtAmount}, OnHand: row.OnHand, Reserved: row.Reserved, Available: row.OnHand - row.Reserved, Assets: []asset.Asset{}})
	}

	if err := r.hydrateSkuAssets(ctx, items); err != nil {
		return pagination.CursorPage[sku.SKU]{}, err
	}

	page := pagination.CursorPage[sku.SKU]{Items: items}
	if hasMore {
		page.NextCursor = items[len(items)-1].ID
	}

	return page, nil
}

func (r *SKURepository) SetInventory(ctx context.Context, in sku.InventoryAdjustment) error {
	result := r.db.WithContext(ctx).Exec("UPDATE skus SET on_hand=?, reserved=? WHERE public_id=? AND archived_at IS NULL", in.OnHand, in.Reserved, in.SKUID)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected != 0 {
		return nil
	}

	var count int64
	if err := r.db.WithContext(ctx).Table("skus").Where("public_id=? AND archived_at IS NULL", in.SKUID).Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		return sku.ErrNotFound
	}

	return nil
}

func (r *SKURepository) hydrateSkuAssets(ctx context.Context, skus []sku.SKU) error {
	if len(skus) == 0 {
		return nil
	}

	positions := make(map[string]int, len(skus))
	ids := make([]string, len(skus))
	for i := range skus {
		ids[i] = skus[i].ID
		positions[skus[i].ID] = i
	}

	var assets []skuAssetRow
	query := `
		SELECT
			s.public_id AS sku_id,
			a.public_id,
			a.media_type,
			a.cdn_url AS url,
			COALESCE(a.alt_text, '') AS alt_text,
			a.role,
			a.sort_order
		FROM assets AS a
		JOIN skus AS s ON s.id = a.sku_id
		WHERE
			a.archived_at IS NULL
			AND s.public_id IN ?
		ORDER BY
			a.role,
			a.sort_order
	`
	if err := r.db.WithContext(ctx).Raw(query, ids).Scan(&assets).Error; err != nil {
		return err
	}

	for _, row := range assets {
		i := positions[row.SkuID]
		skus[i].Assets = append(skus[i].Assets, asset.Asset{ID: row.PublicID, MediaType: row.MediaType, URL: row.URL, AltText: row.AltText, Role: row.Role, SortOrder: row.SortOrder})
	}

	return nil
}
