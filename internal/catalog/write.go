package catalog

// ProductAggregate is the admin editor's write payload: a product plus its
// colourways and SKUs, referencing dimensions by business key (style code,
// brand name, category/colourway slug, size scale + code).
type ProductAggregate struct {
	Product    ProductWrite
	Colourways []ColourwayWrite
	Skus       []SkuWrite
	Images     []ImageWrite
}

type ProductWrite struct {
	ID           string // style_code
	Slug         string
	Name         string
	Subtitle     string
	Brand        string // brand name
	Gender       string
	ProductType  string
	Description  string
	CategorySlug string
	SizeScale    string
	BasePrice    int64
	PublishedAt  string // YYYY-MM-DD
}

type ColourwayWrite struct {
	ID        string // within-payload correlation key, not persisted
	Name      string
	SwatchHex string
	Price     int64
	IsDefault bool
}

// ImageWrite is one product image; ColourwayID (a business-style colourway
// id), when set, scopes the image to that colourway. Empty means the image
// is shared across all of the product's colourways.
type ImageWrite struct {
	URL         string
	ColourwayID string
}

type SkuWrite struct {
	ID          string // sku_code
	ColourwayID string
	Size        string // size code
	SizeScale   string
	StockQty    int
	Price       int64
}
