package mysqladapter

// skuRow is the scan target for the SKU list query.
type skuRow struct {
	ID, Code, Barcode, ProductID          string
	ColourwayID, ColourwayName, HexCode   string
	SizeID, ScaleCode, SizeCode, SizeName string
	SortOrder, OnHand, Reserved           int
	Amount                                int64
	CompareAtAmount                       *int64
}

// skuAssetRow is the scan target for hydrateSkuAssets.
type skuAssetRow struct {
	SkuID                                   string
	PublicID, MediaType, URL, AltText, Role string
	SortOrder                               int
}
