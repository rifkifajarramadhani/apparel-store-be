package catalog

type Sku struct {
	ID        string    `json:"id"`
	Code      string    `json:"code"`
	Barcode   string    `json:"barcode,omitempty"`
	ProductID string    `json:"productId"`
	Colourway Colourway `json:"colourway"`
	Size      Size      `json:"size"`
	Price     Money     `json:"price"`
	OnHand    int       `json:"onHand"`
	Reserved  int       `json:"reserved"`
	Available int       `json:"available"`
	Assets    []Asset   `json:"assets"`
}

type SkuQuery struct {
	ProductID   string
	ColourwayID string
	Currency    string
	Cursor      string
	Limit       int
}

type InventoryAdjustment struct {
	SkuID    string `json:"skuId"`
	OnHand   int    `json:"onHand"`
	Reserved int    `json:"reserved"`
}
