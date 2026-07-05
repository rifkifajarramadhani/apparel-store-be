package catalogseed

type Product struct {
	ID, Slug, Name, Subtitle, Brand, Gender, Type, CategoryID, CategorySlug, SizeScale, Description, PublishedAt string
	CollectionIDs                                                                                                []string `json:"collectionIds"`
	BasePrice                                                                                                    int      `json:"basePrice"`
}

type Colourway struct {
	ID         string `json:"id"`
	ProductID  string `json:"productId"`
	StyleColor string `json:"styleColor"`
	Name       string `json:"name"`
	SwatchHex  string `json:"swatchHex"`
	Price      int
	Images     []string
}

type SKU struct {
	ID          string `json:"id"`
	ColourwayID string `json:"colorwayId"`
	ProductID   string `json:"productId"`
	Size        string `json:"size"`
	SizeLabel   string `json:"sizeLabel"`
	SizeScale   string `json:"sizeScale"`
	StockQty    int    `json:"stockQty"`
	Price       int    `json:"price"`
}

type Category struct {
	ID, Slug, Name string
	ParentID       *string
	Gender         string
}

type Collection struct{ ID, Slug, Name string }

type SizeScale struct {
	ID    string
	Sizes []string
}
