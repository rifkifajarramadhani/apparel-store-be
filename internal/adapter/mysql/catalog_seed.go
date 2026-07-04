package mysqladapter

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/catalog"
	"gorm.io/gorm"
)

func seedPublicID(prefix, value string) string {
	sum := sha256.Sum256([]byte(value))
	return prefix + strings.ToUpper(fmt.Sprintf("%x", sum[:12]))
}

// SeedCatalog replaces catalog fixture data in one transaction. It targets only
// the normalized schema and is safe to rerun on a clean development database.
func (r *CatalogRepository) SeedCatalog(ctx context.Context, products []catalog.SeedProduct, colourways []catalog.SeedColourway, skus []catalog.SeedSKU, categories []catalog.SeedCategory, collections []catalog.SeedCollection, scales []catalog.SeedSizeScale) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, table := range []string{"sku_assets", "product_assets", "inventory_balances", "prices", "skus", "product_collections", "product_categories", "products", "assets", "colourways", "sizes", "size_scales", "collections", "categories", "brands", "inventory_locations"} {
			if err := tx.Exec("DELETE FROM " + table).Error; err != nil {
				return err
			}
		}
		brandIDs := map[string]uint64{}
		for _, p := range products {
			key := strings.TrimSpace(p.Brand)
			if _, ok := brandIDs[key]; ok {
				continue
			}
			slug := strings.ToLower(strings.ReplaceAll(key, " ", "-"))
			if err := tx.Exec("INSERT INTO brands(public_id,slug,name) VALUES(?,?,?)", seedPublicID("BR", key), slug, key).Error; err != nil {
				return err
			}
			id, err := seedRowID(tx, "brands", "slug", slug)
			if err != nil {
				return err
			}
			brandIDs[key] = id
		}
		categoryIDs := map[string]uint64{}
		for _, c := range categories {
			if err := tx.Exec("INSERT INTO categories(public_id,slug,name,gender) VALUES(?,?,?,?)", seedPublicID("CA", c.ID), c.Slug, c.Name, c.Gender).Error; err != nil {
				return err
			}
			id, err := seedRowID(tx, "categories", "slug", c.Slug)
			if err != nil {
				return err
			}
			categoryIDs[c.ID] = id
		}
		for _, c := range categories {
			if c.ParentID != nil {
				if err := tx.Exec("UPDATE categories SET parent_id=? WHERE id=?", categoryIDs[*c.ParentID], categoryIDs[c.ID]).Error; err != nil {
					return err
				}
			}
		}
		collectionIDs := map[string]uint64{}
		for _, c := range collections {
			if err := tx.Exec("INSERT INTO collections(public_id,slug,name) VALUES(?,?,?)", seedPublicID("CL", c.ID), c.Slug, c.Name).Error; err != nil {
				return err
			}
			id, err := seedRowID(tx, "collections", "slug", c.Slug)
			if err != nil {
				return err
			}
			collectionIDs[c.ID] = id
		}
		scaleIDs, sizeIDs := map[string]uint64{}, map[string]uint64{}
		for _, scale := range scales {
			if err := tx.Exec("INSERT INTO size_scales(public_id,code,name) VALUES(?,?,?)", seedPublicID("SC", scale.ID), scale.ID, scale.ID).Error; err != nil {
				return err
			}
			id, err := seedRowID(tx, "size_scales", "code", scale.ID)
			if err != nil {
				return err
			}
			scaleIDs[scale.ID] = id
			for order, size := range scale.Sizes {
				key := scale.ID + ":" + size
				if err := tx.Exec("INSERT INTO sizes(public_id,size_scale_id,code,name,sort_order) VALUES(?,?,?,?,?)", seedPublicID("SZ", key), scaleIDs[scale.ID], size, size, order).Error; err != nil {
					return err
				}
				var sizeID uint64
				if err := tx.Raw("SELECT id FROM sizes WHERE size_scale_id=? AND code=?", scaleIDs[scale.ID], size).Scan(&sizeID).Error; err != nil {
					return err
				}
				sizeIDs[key] = sizeID
			}
		}
		colourIDs := map[string]uint64{}
		for _, colour := range colourways {
			slug := strings.ToLower(strings.ReplaceAll(colour.ID, " ", "-"))
			if err := tx.Exec("INSERT INTO colourways(public_id,slug,name,colour_family,hex_code) VALUES(?,?,?,?,?)", seedPublicID("CO", colour.ID), slug, colour.Name, colour.ColorFamily, colour.SwatchHex).Error; err != nil {
				return err
			}
			id, err := seedRowID(tx, "colourways", "slug", slug)
			if err != nil {
				return err
			}
			colourIDs[colour.ID] = id
		}
		productIDs := map[string]uint64{}
		for _, p := range products {
			published, _ := time.Parse("2006-01-02", p.PublishedAt)
			if err := tx.Exec("INSERT INTO products(public_id,style_code,slug,brand_id,name,subtitle,gender,product_type,description,published_at) VALUES(?,?,?,?,?,?,?,?,?,?)", seedPublicID("PR", p.ID), p.ID, p.Slug, brandIDs[strings.TrimSpace(p.Brand)], p.Name, p.Subtitle, p.Gender, p.Type, p.Description, published).Error; err != nil {
				return err
			}
			id, err := seedRowID(tx, "products", "style_code", p.ID)
			if err != nil {
				return err
			}
			productIDs[p.ID] = id
			if categoryIDs[p.CategoryID] != 0 {
				if err := tx.Exec("INSERT INTO product_categories(product_id,category_id,is_primary) VALUES(?,?,TRUE)", productIDs[p.ID], categoryIDs[p.CategoryID]).Error; err != nil {
					return err
				}
			}
			for _, collectionID := range p.CollectionIDs {
				if collectionIDs[collectionID] != 0 {
					if err := tx.Exec("INSERT INTO product_collections(product_id,collection_id) VALUES(?,?)", productIDs[p.ID], collectionIDs[collectionID]).Error; err != nil {
						return err
					}
				}
			}
			if err := tx.Exec("INSERT INTO prices(public_id,product_id,currency,amount,valid_from) VALUES(?,?,'IDR',?,'1970-01-01')", seedPublicID("PP", p.ID), productIDs[p.ID], p.BasePrice).Error; err != nil {
				return err
			}
		}
		if err := tx.Exec("INSERT INTO inventory_locations(public_id,code,name) VALUES('IL000000000000000000000001','default','Default')").Error; err != nil {
			return err
		}
		var locationID uint64
		if err := tx.Raw("SELECT id FROM inventory_locations WHERE code='default'").Scan(&locationID).Error; err != nil {
			return err
		}
		skuIDs := map[string]uint64{}
		for _, sku := range skus {
			if err := tx.Exec("INSERT INTO skus(public_id,sku_code,product_id,colourway_id,size_id) VALUES(?,?,?,?,?)", seedPublicID("SK", sku.ID), sku.ID, productIDs[sku.ProductID], colourIDs[sku.ColourwayID], sizeIDs[sku.SizeScale+":"+sku.Size]).Error; err != nil {
				return err
			}
			var skuID uint64
			if err := tx.Raw("SELECT id FROM skus WHERE sku_code=?", sku.ID).Scan(&skuID).Error; err != nil {
				return err
			}
			skuIDs[sku.ID] = skuID
			if err := tx.Exec("INSERT INTO inventory_balances(sku_id,location_id,on_hand,reserved) VALUES(?,?,?,0)", skuID, locationID, sku.StockQty).Error; err != nil {
				return err
			}
			if sku.Price > 0 {
				if err := tx.Exec("INSERT INTO prices(public_id,sku_id,currency,amount,valid_from) VALUES(?,?,'IDR',?,'1970-01-01')", seedPublicID("SP", sku.ID), skuID, sku.Price).Error; err != nil {
					return err
				}
			}
		}
		for _, colour := range colourways {
			for order, imageURL := range colour.Images {
				if imageURL == "" {
					continue
				}
				if err := tx.Exec("INSERT IGNORE INTO assets(public_id,media_type,url,alt_text) VALUES(?,'image',?,?)", seedPublicID("AS", imageURL), imageURL, colour.Name).Error; err != nil {
					return err
				}
				assetID, err := seedRowID(tx, "assets", "url", imageURL)
				if err != nil {
					return err
				}
				if err := tx.Exec("INSERT IGNORE INTO product_assets(product_id,asset_id,role,sort_order) VALUES(?,?,'product_image',?)", productIDs[colour.ProductID], assetID, order).Error; err != nil {
					return err
				}
				for _, sku := range skus {
					if sku.ColourwayID == colour.ID {
						if err := tx.Exec("INSERT IGNORE INTO sku_assets(sku_id,asset_id,role,sort_order) VALUES(?,?,'product_image',?)", skuIDs[sku.ID], assetID, order).Error; err != nil {
							return err
						}
					}
				}
			}
		}
		return nil
	})
}

func seedRowID(tx *gorm.DB, table, column string, value any) (uint64, error) {
	var id uint64
	err := tx.Table(table).Select("id").Where(column+" = ?", value).Scan(&id).Error
	return id, err
}
