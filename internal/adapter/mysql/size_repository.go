package mysqladapter

import (
	"context"

	appsize "github.com/rifkifajarramadhani/golang-clean-architecture/internal/size"
)

func (r *SizeRepository) List(ctx context.Context) ([]appsize.Size, error) {
	var out []appsize.Size
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			sz.public_id AS id,
			ss.code AS scale_code,
			sz.code,
			sz.name,
			sz.sort_order
		FROM sizes AS sz
		JOIN size_scales AS ss ON ss.id = sz.size_scale_id
		WHERE
			sz.archived_at IS NULL
			AND ss.archived_at IS NULL
		ORDER BY
			ss.code,
			sz.sort_order
	`).Scan(&out).Error
	return out, err
}
