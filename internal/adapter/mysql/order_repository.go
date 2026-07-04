package mysqladapter

import (
	"context"
	"errors"
	"time"

	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/order"
	"gorm.io/gorm"
)

type checkoutSKURow struct {
	ID, PublicID, ProductPublicID, ProductName, SizeName string
	ProductRefID                                         uint64
	Amount                                               int
	OnHand, Reserved                                     int
}

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// Create resolves each SKU under a FOR UPDATE lock, recomputes the price from
// the catalog, decrements stock, and persists the order + items atomically.
func (r *OrderRepository) Create(ctx context.Context, userID int, lines []order.Line) (order.Order, error) {
	var created orderModel
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		items := make([]orderItemModel, 0, len(lines))
		total := 0
		for _, line := range lines {
			var sku checkoutSKURow
			now := time.Now().UTC()
			err := tx.Raw(`SELECT s.id,s.public_id,p.id product_ref_id,p.public_id product_public_id,p.name product_name,sz.name size_name,
				COALESCE((SELECT sp.amount FROM prices sp WHERE sp.sku_id=s.id AND sp.currency='IDR' AND sp.archived_at IS NULL AND sp.valid_from<=? AND (sp.valid_to IS NULL OR sp.valid_to>?) ORDER BY sp.valid_from DESC LIMIT 1),
				(SELECT pp.amount FROM prices pp WHERE pp.product_id=s.product_id AND pp.currency='IDR' AND pp.archived_at IS NULL AND pp.valid_from<=? AND (pp.valid_to IS NULL OR pp.valid_to>?) ORDER BY pp.valid_from DESC LIMIT 1),0) amount,
				ib.on_hand,ib.reserved
				FROM skus s JOIN products p ON p.id=s.product_id AND p.archived_at IS NULL
				JOIN sizes sz ON sz.id=s.size_id AND sz.archived_at IS NULL
				JOIN inventory_balances ib ON ib.sku_id=s.id
				JOIN inventory_locations il ON il.id=ib.location_id AND il.code='default' AND il.archived_at IS NULL
				WHERE s.public_id=? AND s.archived_at IS NULL FOR UPDATE`, now, now, now, now, line.SkuID).Take(&sku).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return order.ErrNotFound
				}
				return err
			}
			if sku.OnHand-sku.Reserved < line.Qty {
				return order.ErrOutOfStock
			}
			if err := tx.Exec("UPDATE inventory_balances SET on_hand=on_hand-? WHERE sku_id=? AND location_id=(SELECT id FROM inventory_locations WHERE code='default')", line.Qty, sku.ID).Error; err != nil {
				return err
			}
			items = append(items, orderItemModel{
				SkuID: sku.PublicID, ProductID: sku.ProductPublicID, Name: sku.ProductName,
				SkuRefID: publicUint64(sku.ID), ProductRefID: &sku.ProductRefID,
				Size: sku.SizeName, UnitPrice: sku.Amount, Qty: line.Qty,
			})
			total += sku.Amount * line.Qty
		}
		created = orderModel{
			UserID: userID, Status: order.StatusConfirmed, Total: total,
			CreatedAt: time.Now().UTC(), Items: items,
		}
		return tx.Create(&created).Error
	})
	if err != nil {
		return order.Order{}, err
	}
	return toOrder(created), nil
}

func publicUint64(raw string) *uint64 {
	var value uint64
	for _, char := range raw {
		if char >= '0' && char <= '9' {
			value = value*10 + uint64(char-'0')
		}
	}
	return &value
}

func (r *OrderRepository) ListByUser(ctx context.Context, userID int) ([]order.Order, error) {
	var records []orderModel
	if err := r.db.WithContext(ctx).Preload("Items").
		Where("user_id = ?", userID).Order("id DESC").Find(&records).Error; err != nil {
		return nil, err
	}
	orders := make([]order.Order, 0, len(records))
	for _, record := range records {
		orders = append(orders, toOrder(record))
	}
	return orders, nil
}

func (r *OrderRepository) GetByIDForUser(ctx context.Context, userID, orderID int) (order.Order, error) {
	var record orderModel
	if err := r.db.WithContext(ctx).Preload("Items").
		Where("user_id = ?", userID).First(&record, "id = ?", orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return order.Order{}, order.ErrNotFound
		}
		return order.Order{}, err
	}
	return toOrder(record), nil
}

var _ order.Repository = (*OrderRepository)(nil)
