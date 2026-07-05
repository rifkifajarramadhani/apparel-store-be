package mysqladapter

import (
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/catalog"
	"gorm.io/gorm"
)

type CatalogRepository struct{ db *gorm.DB }

func NewCatalogRepository(db *gorm.DB) *CatalogRepository {
	return &CatalogRepository{db: db}
}

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

var _ catalog.Repository = (*CatalogRepository)(nil)
