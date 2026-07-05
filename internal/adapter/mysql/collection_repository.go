package mysqladapter

import (
	"context"

	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/collection"
)

func (r *CollectionRepository) List(ctx context.Context) ([]collection.Collection, error) {
	var out []collection.Collection
	err := r.db.WithContext(ctx).Table("collections").Select("public_id id,slug,name").Where("archived_at IS NULL").Order("name").Scan(&out).Error
	return out, err
}
