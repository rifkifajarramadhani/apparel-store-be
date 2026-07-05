package catalog

// Seed records describe the source fixture only; they are not runtime catalog entities.
type SeedProduct struct {
	ID, Slug, Name, Subtitle, Brand, Gender, Type, CategoryID, CategorySlug, SizeScale, Description, PublishedAt string
	CollectionIDs                                                                                                []string `json:"collectionIds"`
	BasePrice                                                                                                    int      `json:"basePrice"`
}
type SeedColourway struct {
	ID         string `json:"id"`
	ProductID  string `json:"productId"`
	StyleColor string `json:"styleColor"`
	Name       string `json:"name"`
	SwatchHex  string `json:"swatchHex"`
	Price      int
	Images     []string
}
type SeedSKU struct {
	ID          string `json:"id"`
	ColourwayID string `json:"colorwayId"`
	ProductID   string `json:"productId"`
	Size        string `json:"size"`
	SizeLabel   string `json:"sizeLabel"`
	SizeScale   string `json:"sizeScale"`
	StockQty    int    `json:"stockQty"`
	Price       int    `json:"price"`
}
type SeedCategory struct {
	ID, Slug, Name string
	ParentID       *string
	Gender         string
}
type SeedCollection struct{ ID, Slug, Name string }
type SeedSizeScale struct {
	ID    string
	Sizes []string
}
