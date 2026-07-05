package mysqladapter

import (
	"context"

	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/colourway"
)

func (r *ColourwayRepository) List(ctx context.Context) ([]colourway.Colourway, error) {
	var out []colourway.Colourway
	err := r.db.WithContext(ctx).Table("colourways").Select("public_id id,name,hex_code").Where("archived_at IS NULL").Order("name").Scan(&out).Error
	return out, err
}
