package mysqladapter

import (
	"context"

	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/category"
)

func (r *CategoryRepository) List(ctx context.Context) ([]category.Category, error) {
	var out []category.Category
	err := r.db.WithContext(ctx).Raw("SELECT c.public_id id,p.public_id parent_id,c.slug,c.name FROM categories c LEFT JOIN categories p ON p.id=c.parent_id WHERE c.archived_at IS NULL ORDER BY c.name").Scan(&out).Error
	return out, err
}
