package mysqladapter

type orderItemModel struct {
	ID           int `gorm:"primaryKey"`
	OrderID      int `gorm:"index"`
	SkuID        string
	ProductID    string
	SkuRefID     *uint64
	ProductRefID *uint64
	Name         string
	Size         string
	UnitPrice    int
	Qty          int
}

func (orderItemModel) TableName() string { return "order_items" }
