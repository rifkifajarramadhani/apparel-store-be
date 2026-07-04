package mysqladapter

import (
	"context"
	"errors"
	"time"

	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/order"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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
			var sku skuModel
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				First(&sku, "id = ?", line.SkuID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return order.ErrNotFound
				}
				return err
			}
			if sku.StockQty < line.Qty {
				return order.ErrOutOfStock
			}
			var product productModel
			if err := tx.Select("name").First(&product, "id = ?", sku.ProductID).Error; err != nil {
				return err
			}
			newQty := sku.StockQty - line.Qty
			if err := tx.Model(&skuModel{}).Where("id = ?", sku.ID).Updates(map[string]any{
				"stock_qty": newQty, "in_stock": newQty > 0,
			}).Error; err != nil {
				return err
			}
			items = append(items, orderItemModel{
				SkuID: sku.ID, ProductID: sku.ProductID, Name: product.Name,
				Size: sku.Size, UnitPrice: sku.Price, Qty: line.Qty,
			})
			total += sku.Price * line.Qty
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
