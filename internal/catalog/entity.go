package catalog

// Field shapes mirror the frontend types in src/types/catalog.ts. Prices are
// integers in minor currency units. IDs are natural string keys (style codes,
// style-colors, slugs), not autoincrement integers.

type Swatch struct {
	StyleColor string `json:"styleColor"`
	Hex        string `json:"hex"`
}

type Product struct {
	ID                string   `json:"id"`
	Slug              string   `json:"slug"`
	Name              string   `json:"name"`
	Subtitle          string   `json:"subtitle"`
	Brand             string   `json:"brand"`
	Gender            string   `json:"gender"`
	Type              string   `json:"type"`
	CategoryID        string   `json:"categoryId"`
	CategorySlug      string   `json:"categorySlug"`
	CollectionIDs     []string `json:"collectionIds"`
	SizeScale         string   `json:"sizeScale"`
	BasePrice         int      `json:"basePrice"`
	MinPrice          int      `json:"minPrice"`
	MaxPrice          int      `json:"maxPrice"`
	Badges            []string `json:"badges"`
	ColorwayCount     int      `json:"colorwayCount"`
	ColorFamilies     []string `json:"colorFamilies"`
	Swatches          []Swatch `json:"swatches"`
	ThumbnailURL      string   `json:"thumbnailUrl"`
	HoverImageURL     string   `json:"hoverImageUrl"`
	DefaultColorwayID string   `json:"defaultColorwayId"`
	Sizes             []string `json:"sizes"`
	Description       string   `json:"description"`
	PublishedAt       string   `json:"publishedAt"`
}

type Colorway struct {
	ID          string   `json:"id"`
	ProductID   string   `json:"productId"`
	StyleColor  string   `json:"styleColor"`
	Name        string   `json:"name"`
	ColorFamily string   `json:"colorFamily"`
	SwatchHex   string   `json:"swatchHex"`
	Price       int      `json:"price"`
	IsDefault   bool     `json:"isDefault"`
	OnSale      bool     `json:"onSale"`
	Images      []string `json:"images"`
	Skus        []Sku    `json:"skus,omitempty"` // present when fetched with ?_embed=skus
}

type Sku struct {
	ID         string `json:"id"`
	ColorwayID string `json:"colorwayId"`
	ProductID  string `json:"productId"`
	Size       string `json:"size"`
	SizeLabel  string `json:"sizeLabel"`
	SizeScale  string `json:"sizeScale"`
	InStock    bool   `json:"inStock"`
	StockQty   int    `json:"stockQty"`
	Price      int    `json:"price"`
}

type Category struct {
	ID       string  `json:"id"`
	Slug     string  `json:"slug"`
	Name     string  `json:"name"`
	ParentID *string `json:"parentId"`
	Gender   string  `json:"gender"`
	Level    int     `json:"level"`
}

type Collection struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type SizeScale struct {
	ID    string   `json:"id"`
	Sizes []string `json:"sizes"`
}

// Aggregate is the {product, colorways, skus} bundle the admin editor submits
// and reads back. Mirrors ProductAggregateInputSchema on the frontend.
type Aggregate struct {
	Product   Product    `json:"product"`
	Colorways []Colorway `json:"colorways"`
	Skus      []Sku      `json:"skus"`
}

// ProductQuery holds the coarse server-side filters the storefront sends.
// Only the params src/lib/api.ts actually builds are supported.
type ProductQuery struct {
	CategoryID   string
	CategorySlug string
	Gender       string
	Slug         string
	Q            string
	MinPrice     *int
	MaxPrice     *int
	SortBy       string // "" | "publishedAt" | "minPrice"
	SortDesc     bool
	Page         int // 0 => unpaginated
	Limit        int // 0 => unpaginated
}
