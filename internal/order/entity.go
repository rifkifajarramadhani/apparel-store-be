package order

import "time"

const (
	StatusPending   = "pending"
	StatusConfirmed = "confirmed"
)

// Line is a requested purchase quantity for a SKU. Price is resolved
// server-side from the catalog, never taken from the client.
type Line struct {
	SkuID string
	Qty   int
}

type Item struct {
	SkuID     string `json:"skuId"`
	ProductID string `json:"productId"`
	Name      string `json:"name"`
	Size      string `json:"size"`
	UnitPrice int    `json:"unitPrice"`
	Qty       int    `json:"qty"`
}

type Order struct {
	ID        int       `json:"id"`
	UserID    int       `json:"userId"`
	Status    string    `json:"status"`
	Total     int       `json:"total"`
	CreatedAt time.Time `json:"createdAt"`
	Items     []Item    `json:"items"`
}
