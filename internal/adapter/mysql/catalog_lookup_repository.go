package mysqladapter

import (
	"context"

	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/catalog"
)

func (r *CatalogRepository) ListBrands(ctx context.Context) ([]catalog.Brand, error) {
	var out []catalog.Brand
	err := r.db.WithContext(ctx).Table("brands").Select("public_id id,slug,name").Where("archived_at IS NULL").Order("name").Scan(&out).Error
	return out, err
}
func (r *CatalogRepository) ListCategories(ctx context.Context) ([]catalog.Category, error) {
	var out []catalog.Category
	err := r.db.WithContext(ctx).Raw("SELECT c.public_id id,p.public_id parent_id,c.slug,c.name FROM categories c LEFT JOIN categories p ON p.id=c.parent_id WHERE c.archived_at IS NULL ORDER BY c.name").Scan(&out).Error
	return out, err
}
func (r *CatalogRepository) ListCollections(ctx context.Context) ([]catalog.Collection, error) {
	var out []catalog.Collection
	err := r.db.WithContext(ctx).Table("collections").Select("public_id id,slug,name").Where("archived_at IS NULL").Order("name").Scan(&out).Error
	return out, err
}
func (r *CatalogRepository) ListColourways(ctx context.Context) ([]catalog.Colourway, error) {
	var out []catalog.Colourway
	err := r.db.WithContext(ctx).Table("colourways").Select("public_id id,name,hex_code").Where("archived_at IS NULL").Order("name").Scan(&out).Error
	return out, err
}
func (r *CatalogRepository) ListSizes(ctx context.Context) ([]catalog.Size, error) {
	var out []catalog.Size
	err := r.db.WithContext(ctx).Raw("SELECT sz.public_id id,ss.code scale_code,sz.code,sz.name,sz.sort_order FROM sizes sz JOIN size_scales ss ON ss.id=sz.size_scale_id WHERE sz.archived_at IS NULL AND ss.archived_at IS NULL ORDER BY ss.code,sz.sort_order").Scan(&out).Error
	return out, err
}
