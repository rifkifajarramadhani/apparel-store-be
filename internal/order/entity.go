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
	SkuID     string
	ProductID string
	Name      string
	Size      string
	UnitPrice int
	Qty       int
}

type Order struct {
	ID        int
	UserID    int
	Status    string
	Total     int
	CreatedAt time.Time
	Items     []Item
}
