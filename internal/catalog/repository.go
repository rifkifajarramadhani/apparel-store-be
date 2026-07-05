package catalog

import "context"

type Repository interface {
	ListProducts(context.Context, ProductQuery) (CursorPage[Product], error)
	GetProduct(context.Context, string, string) (Product, error)
	ListSkus(context.Context, SkuQuery) (CursorPage[Sku], error)
	ListBrands(context.Context) ([]Brand, error)
	ListCategories(context.Context) ([]Category, error)
	ListCollections(context.Context) ([]Collection, error)
	ListColourways(context.Context) ([]Colourway, error)
	ListSizes(context.Context) ([]Size, error)
	SetInventory(context.Context, InventoryAdjustment) error
	CreateProduct(context.Context, ProductAggregate) error
	UpdateProduct(context.Context, ProductAggregate) error
	DeleteProduct(context.Context, string) error
}
