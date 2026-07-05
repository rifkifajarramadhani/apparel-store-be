package mysqladapter

// productRow is the scan target for the product list/detail queries.
type productRow struct {
	ID, PublicID, StyleCode, Slug, Name, Subtitle, Gender, ProductType, Description string
	BrandPublicID, BrandSlug, BrandName                                             string
}

// The link rows below are scan targets for hydrateProducts, one per child
// dimension attached to a product.
type categoryLinkRow struct {
	ProductID            uint64
	PublicID, Slug, Name string
	ParentPublicID       *string
}
type colourLinkRow struct {
	ProductID               uint64
	PublicID, Name, HexCode string
}
type sizeLinkRow struct {
	ProductID                       uint64
	PublicID, ScaleCode, Code, Name string
	SortOrder                       int
}
type assetLinkRow struct {
	ProductID                               uint64
	PublicID, MediaType, URL, AltText, Role string
	SortOrder                               int
	ColourwayID                             string
	SkuID                                   string
}
type priceRangeRow struct {
	ProductID            uint64
	MinAmount, MaxAmount int64
}
