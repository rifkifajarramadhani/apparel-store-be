package mysqladapter

import (
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/brand"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/category"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/collection"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/colourway"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/product"
	appsize "github.com/rifkifajarramadhani/golang-clean-architecture/internal/size"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/sku"
	"gorm.io/gorm"
)

type ProductRepository struct{ db *gorm.DB }
type SKURepository struct{ db *gorm.DB }
type BrandRepository struct{ db *gorm.DB }
type CategoryRepository struct{ db *gorm.DB }
type CollectionRepository struct{ db *gorm.DB }
type ColourwayRepository struct{ db *gorm.DB }
type SizeRepository struct{ db *gorm.DB }
type CatalogSeeder struct{ db *gorm.DB }

func NewProductRepository(db *gorm.DB) *ProductRepository { return &ProductRepository{db: db} }
func NewSKURepository(db *gorm.DB) *SKURepository         { return &SKURepository{db: db} }
func NewBrandRepository(db *gorm.DB) *BrandRepository     { return &BrandRepository{db: db} }
func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}
func NewCollectionRepository(db *gorm.DB) *CollectionRepository {
	return &CollectionRepository{db: db}
}
func NewColourwayRepository(db *gorm.DB) *ColourwayRepository {
	return &ColourwayRepository{db: db}
}
func NewSizeRepository(db *gorm.DB) *SizeRepository { return &SizeRepository{db: db} }
func NewCatalogSeeder(db *gorm.DB) *CatalogSeeder   { return &CatalogSeeder{db: db} }

// parseUintID strips the internal numeric id out of a scanned string column.
func parseUintID(raw string) uint64 {
	var id uint64
	for _, c := range raw {
		if c >= '0' && c <= '9' {
			id = id*10 + uint64(c-'0')
		}
	}

	return id
}

var (
	_ product.Repository    = (*ProductRepository)(nil)
	_ sku.Repository        = (*SKURepository)(nil)
	_ brand.Repository      = (*BrandRepository)(nil)
	_ category.Repository   = (*CategoryRepository)(nil)
	_ collection.Repository = (*CollectionRepository)(nil)
	_ colourway.Repository  = (*ColourwayRepository)(nil)
	_ appsize.Repository    = (*SizeRepository)(nil)
)
