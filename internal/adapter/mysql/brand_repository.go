package mysqladapter

import (
	"context"

	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/brand"
)

func (r *BrandRepository) List(ctx context.Context) ([]brand.Brand, error) {
	var out []brand.Brand
	err := r.db.WithContext(ctx).Table("brands").Select("public_id id,slug,name").Where("archived_at IS NULL").Order("name").Scan(&out).Error
	return out, err
}
