package mysqladapter

import "github.com/rifkifajarramadhani/golang-clean-architecture/internal/catalog"

// Catalog tables use natural string primary keys. Slice/struct fields are
// stored as JSON columns via GORM's json serializer.

type productModel struct {
	ID                string   `gorm:"primaryKey;size:64"`
	Slug              string   `gorm:"index;size:191"`
	Name              string   `gorm:"size:191"`
	Subtitle          string   `gorm:"size:191"`
	Brand             string   `gorm:"size:80"`
	Gender            string   `gorm:"index;size:16"`
	Type              string   `gorm:"column:type;size:80"`
	CategoryID        string   `gorm:"index;size:64"`
	CategorySlug      string   `gorm:"index;size:64"`
	CollectionIDs     []string `gorm:"serializer:json"`
	SizeScale         string   `gorm:"size:64"`
	BasePrice         int
	MinPrice          int `gorm:"index"`
	MaxPrice          int
	Badges            []string `gorm:"serializer:json"`
	ColorwayCount     int
	ColorFamilies     []string         `gorm:"serializer:json"`
	Swatches          []catalog.Swatch `gorm:"serializer:json"`
	ThumbnailURL      string           `gorm:"size:512"`
	HoverImageURL     string           `gorm:"size:512"`
	DefaultColorwayID string           `gorm:"size:64"`
	Sizes             []string         `gorm:"serializer:json"`
	Description       string           `gorm:"type:text"`
	PublishedAt       string           `gorm:"index;size:32"`
}

func (productModel) TableName() string { return "products" }

type colorwayModel struct {
	ID          string `gorm:"primaryKey;size:64"`
	ProductID   string `gorm:"index;size:64"`
	StyleColor  string `gorm:"size:64"`
	Name        string `gorm:"size:191"`
	ColorFamily string `gorm:"size:40"`
	SwatchHex   string `gorm:"size:16"`
	Price       int
	IsDefault   bool
	OnSale      bool
	Images      []string `gorm:"serializer:json"`
}

func (colorwayModel) TableName() string { return "colorways" }

type skuModel struct {
	ID         string `gorm:"primaryKey;size:80"`
	ColorwayID string `gorm:"index;size:64"`
	ProductID  string `gorm:"index;size:64"`
	Size       string `gorm:"size:16"`
	SizeLabel  string `gorm:"size:32"`
	SizeScale  string `gorm:"size:64"`
	InStock    bool
	StockQty   int
	Price      int
}

func (skuModel) TableName() string { return "skus" }

type categoryModel struct {
	ID       string  `gorm:"primaryKey;size:64"`
	Slug     string  `gorm:"size:64"`
	Name     string  `gorm:"size:80"`
	ParentID *string `gorm:"index;size:64"`
	Gender   string  `gorm:"size:16"`
	Level    int
}

func (categoryModel) TableName() string { return "categories" }

type collectionModel struct {
	ID   string `gorm:"primaryKey;size:64"`
	Slug string `gorm:"size:64"`
	Name string `gorm:"size:80"`
}

func (collectionModel) TableName() string { return "collections" }

type sizeScaleModel struct {
	ID    string   `gorm:"primaryKey;size:64"`
	Sizes []string `gorm:"serializer:json"`
}

func (sizeScaleModel) TableName() string { return "size_scales" }

// ── model <-> domain ────────────────────────────────────────────────────────

func toProduct(m productModel) catalog.Product {
	return catalog.Product{
		ID: m.ID, Slug: m.Slug, Name: m.Name, Subtitle: m.Subtitle, Brand: m.Brand,
		Gender: m.Gender, Type: m.Type, CategoryID: m.CategoryID, CategorySlug: m.CategorySlug,
		CollectionIDs: orEmpty(m.CollectionIDs), SizeScale: m.SizeScale, BasePrice: m.BasePrice,
		MinPrice: m.MinPrice, MaxPrice: m.MaxPrice, Badges: orEmpty(m.Badges),
		ColorwayCount: m.ColorwayCount, ColorFamilies: orEmpty(m.ColorFamilies),
		Swatches: orEmptySwatches(m.Swatches), ThumbnailURL: m.ThumbnailURL, HoverImageURL: m.HoverImageURL,
		DefaultColorwayID: m.DefaultColorwayID, Sizes: orEmpty(m.Sizes),
		Description: m.Description, PublishedAt: m.PublishedAt,
	}
}

func fromProduct(p catalog.Product) productModel {
	return productModel{
		ID: p.ID, Slug: p.Slug, Name: p.Name, Subtitle: p.Subtitle, Brand: p.Brand,
		Gender: p.Gender, Type: p.Type, CategoryID: p.CategoryID, CategorySlug: p.CategorySlug,
		CollectionIDs: orEmpty(p.CollectionIDs), SizeScale: p.SizeScale, BasePrice: p.BasePrice,
		MinPrice: p.MinPrice, MaxPrice: p.MaxPrice, Badges: orEmpty(p.Badges),
		ColorwayCount: p.ColorwayCount, ColorFamilies: orEmpty(p.ColorFamilies),
		Swatches: p.Swatches, ThumbnailURL: p.ThumbnailURL, HoverImageURL: p.HoverImageURL,
		DefaultColorwayID: p.DefaultColorwayID, Sizes: orEmpty(p.Sizes),
		Description: p.Description, PublishedAt: p.PublishedAt,
	}
}

func toColorway(m colorwayModel) catalog.Colorway {
	return catalog.Colorway{
		ID: m.ID, ProductID: m.ProductID, StyleColor: m.StyleColor, Name: m.Name,
		ColorFamily: m.ColorFamily, SwatchHex: m.SwatchHex, Price: m.Price,
		IsDefault: m.IsDefault, OnSale: m.OnSale, Images: orEmpty(m.Images),
	}
}

func fromColorway(c catalog.Colorway) colorwayModel {
	return colorwayModel{
		ID: c.ID, ProductID: c.ProductID, StyleColor: c.StyleColor, Name: c.Name,
		ColorFamily: c.ColorFamily, SwatchHex: c.SwatchHex, Price: c.Price,
		IsDefault: c.IsDefault, OnSale: c.OnSale, Images: orEmpty(c.Images),
	}
}

func toSku(m skuModel) catalog.Sku {
	return catalog.Sku{
		ID: m.ID, ColorwayID: m.ColorwayID, ProductID: m.ProductID, Size: m.Size,
		SizeLabel: m.SizeLabel, SizeScale: m.SizeScale, InStock: m.InStock,
		StockQty: m.StockQty, Price: m.Price,
	}
}

func fromSku(s catalog.Sku) skuModel {
	return skuModel{
		ID: s.ID, ColorwayID: s.ColorwayID, ProductID: s.ProductID, Size: s.Size,
		SizeLabel: s.SizeLabel, SizeScale: s.SizeScale, InStock: s.InStock,
		StockQty: s.StockQty, Price: s.Price,
	}
}

func orEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func orEmptySwatches(s []catalog.Swatch) []catalog.Swatch {
	if s == nil {
		return []catalog.Swatch{}
	}
	return s
}
