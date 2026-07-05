package mysqladapter

import (
	"time"

	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/order"
)

type orderModel struct {
	ID        int `gorm:"primaryKey"`
	UserID    int `gorm:"index"`
	Status    string
	Total     int
	CreatedAt time.Time
	Items     []orderItemModel `gorm:"foreignKey:OrderID"`
}

func (orderModel) TableName() string { return "orders" }

func toOrder(m orderModel) order.Order {
	items := make([]order.Item, 0, len(m.Items))
	for _, item := range m.Items {
		items = append(items, order.Item{
			SkuID: item.SkuID, ProductID: item.ProductID, Name: item.Name,
			Size: item.Size, UnitPrice: item.UnitPrice, Qty: item.Qty,
		})
	}

	return order.Order{
		ID: m.ID, UserID: m.UserID, Status: m.Status, Total: m.Total,
		CreatedAt: m.CreatedAt, Items: items,
	}
}
