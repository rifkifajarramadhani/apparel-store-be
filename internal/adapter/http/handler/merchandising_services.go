package handler

// MerchandisingServices groups independently owned domain services for HTTP
// wiring. It is a dependency bundle, not a domain service or repository.
type MerchandisingServices struct {
	Products    ProductService
	SKUs        SKUService
	Brands      BrandService
	Categories  CategoryService
	Collections CollectionService
	Colourways  ColourwayService
	Sizes       SizeService
}
