package mysqladapter

import (
	"context"
	"errors"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/catalog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func isDuplicateKey(err error) bool {
	var mysqlErr *mysqldriver.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

type CatalogRepository struct {
	db *gorm.DB
}

func NewCatalogRepository(db *gorm.DB) *CatalogRepository {
	return &CatalogRepository{db: db}
}

func (r *CatalogRepository) ListProducts(ctx context.Context, q catalog.ProductQuery) ([]catalog.Product, int64, error) {
	tx := r.db.WithContext(ctx).Model(&productModel{})
	if q.CategoryID != "" {
		tx = tx.Where("category_id = ?", q.CategoryID)
	}
	if q.CategorySlug != "" {
		tx = tx.Where("category_slug = ?", q.CategorySlug)
	}
	if q.Gender != "" {
		tx = tx.Where("gender = ?", q.Gender)
	}
	if q.Slug != "" {
		tx = tx.Where("slug = ?", q.Slug)
	}
	if q.MinPrice != nil {
		tx = tx.Where("min_price >= ?", *q.MinPrice)
	}
	if q.MaxPrice != nil {
		tx = tx.Where("max_price <= ?", *q.MaxPrice)
	}
	if q.Q != "" {
		like := "%" + q.Q + "%"
		tx = tx.Where("name LIKE ? OR subtitle LIKE ? OR brand LIKE ? OR type LIKE ?", like, like, like, like)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	switch q.SortBy {
	case "publishedAt":
		tx = tx.Order(orderClause("published_at", q.SortDesc))
	case "minPrice":
		tx = tx.Order(orderClause("min_price", q.SortDesc))
	default:
		tx = tx.Order("published_at DESC")
	}
	if q.Limit > 0 {
		page := q.Page
		if page < 1 {
			page = 1
		}
		tx = tx.Offset((page - 1) * q.Limit).Limit(q.Limit)
	}

	var records []productModel
	if err := tx.Find(&records).Error; err != nil {
		return nil, 0, err
	}
	products := make([]catalog.Product, 0, len(records))
	for _, record := range records {
		products = append(products, toProduct(record))
	}
	return products, total, nil
}

func orderClause(column string, desc bool) string {
	if desc {
		return column + " DESC"
	}
	return column + " ASC"
}

func (r *CatalogRepository) GetProduct(ctx context.Context, id string) (catalog.Product, error) {
	var record productModel
	if err := r.db.WithContext(ctx).First(&record, "id = ?", id).Error; err != nil {
		return catalog.Product{}, mapCatalogNotFound(err)
	}
	return toProduct(record), nil
}

func (r *CatalogRepository) ProductExists(ctx context.Context, id string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&productModel{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *CatalogRepository) GetAggregate(ctx context.Context, productID string) (catalog.Aggregate, error) {
	product, err := r.GetProduct(ctx, productID)
	if err != nil {
		return catalog.Aggregate{}, err
	}
	colorways, err := r.ListColorways(ctx, productID, false)
	if err != nil {
		return catalog.Aggregate{}, err
	}
	skus, err := r.ListSkus(ctx, productID, "")
	if err != nil {
		return catalog.Aggregate{}, err
	}
	return catalog.Aggregate{Product: product, Colorways: colorways, Skus: skus}, nil
}

// SaveAggregate replaces the product row and all of its colorways/skus in one
// transaction, matching the json-server admin route semantics (full replace).
func (r *CatalogRepository) SaveAggregate(ctx context.Context, agg catalog.Aggregate) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("product_id = ?", agg.Product.ID).Delete(&skuModel{}).Error; err != nil {
			return err
		}
		if err := tx.Where("product_id = ?", agg.Product.ID).Delete(&colorwayModel{}).Error; err != nil {
			return err
		}
		model := fromProduct(agg.Product)
		if err := tx.Save(&model).Error; err != nil {
			return err
		}
		for _, cw := range agg.Colorways {
			record := fromColorway(cw)
			if err := tx.Create(&record).Error; err != nil {
				return err
			}
		}
		for _, sku := range agg.Skus {
			record := fromSku(sku)
			if err := tx.Create(&record).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *CatalogRepository) DeleteProduct(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("product_id = ?", id).Delete(&skuModel{}).Error; err != nil {
			return err
		}
		if err := tx.Where("product_id = ?", id).Delete(&colorwayModel{}).Error; err != nil {
			return err
		}
		return tx.Delete(&productModel{}, "id = ?", id).Error
	})
}

func (r *CatalogRepository) ListColorways(ctx context.Context, productID string, embedSkus bool) ([]catalog.Colorway, error) {
	tx := r.db.WithContext(ctx).Model(&colorwayModel{})
	if productID != "" {
		tx = tx.Where("product_id = ?", productID)
	}
	var records []colorwayModel
	if err := tx.Order("is_default DESC, id ASC").Find(&records).Error; err != nil {
		return nil, err
	}
	colorways := make([]catalog.Colorway, 0, len(records))
	for _, record := range records {
		cw := toColorway(record)
		if embedSkus {
			skus, err := r.ListSkus(ctx, "", cw.ID)
			if err != nil {
				return nil, err
			}
			cw.Skus = skus
		}
		colorways = append(colorways, cw)
	}
	return colorways, nil
}

func (r *CatalogRepository) ListSkus(ctx context.Context, productID, colorwayID string) ([]catalog.Sku, error) {
	tx := r.db.WithContext(ctx).Model(&skuModel{})
	if productID != "" {
		tx = tx.Where("product_id = ?", productID)
	}
	if colorwayID != "" {
		tx = tx.Where("colorway_id = ?", colorwayID)
	}
	var records []skuModel
	if err := tx.Order("id ASC").Find(&records).Error; err != nil {
		return nil, err
	}
	skus := make([]catalog.Sku, 0, len(records))
	for _, record := range records {
		skus = append(skus, toSku(record))
	}
	return skus, nil
}

func (r *CatalogRepository) GetSku(ctx context.Context, id string) (catalog.Sku, error) {
	var record skuModel
	if err := r.db.WithContext(ctx).First(&record, "id = ?", id).Error; err != nil {
		return catalog.Sku{}, mapCatalogNotFound(err)
	}
	return toSku(record), nil
}

func (r *CatalogRepository) UpdateSku(ctx context.Context, sku catalog.Sku) error {
	record := fromSku(sku)
	result := r.db.WithContext(ctx).Model(&skuModel{}).Where("id = ?", sku.ID).Updates(map[string]any{
		"colorway_id": record.ColorwayID, "product_id": record.ProductID, "size": record.Size,
		"size_label": record.SizeLabel, "size_scale": record.SizeScale,
		"in_stock": record.InStock, "stock_qty": record.StockQty, "price": record.Price,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return catalog.ErrNotFound
	}
	return nil
}

func (r *CatalogRepository) ListCategories(ctx context.Context) ([]catalog.Category, error) {
	var records []categoryModel
	if err := r.db.WithContext(ctx).Order("level ASC, id ASC").Find(&records).Error; err != nil {
		return nil, err
	}
	categories := make([]catalog.Category, 0, len(records))
	for _, record := range records {
		categories = append(categories, catalog.Category{
			ID: record.ID, Slug: record.Slug, Name: record.Name,
			ParentID: record.ParentID, Gender: record.Gender, Level: record.Level,
		})
	}
	return categories, nil
}

func (r *CatalogRepository) SaveCategory(ctx context.Context, c catalog.Category, create bool) error {
	record := categoryModel{ID: c.ID, Slug: c.Slug, Name: c.Name, ParentID: c.ParentID, Gender: c.Gender, Level: c.Level}
	if create {
		return mapCatalogWriteError(r.db.WithContext(ctx).Create(&record).Error)
	}
	result := r.db.WithContext(ctx).Model(&categoryModel{}).Where("id = ?", c.ID).Updates(map[string]any{
		"slug": c.Slug, "name": c.Name, "parent_id": c.ParentID, "gender": c.Gender, "level": c.Level,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return catalog.ErrNotFound
	}
	return nil
}

func (r *CatalogRepository) DeleteCategory(ctx context.Context, id string) error {
	return requireCatalogAffected(r.db.WithContext(ctx).Delete(&categoryModel{}, "id = ?", id))
}

func (r *CatalogRepository) CountCategoryReferences(ctx context.Context, id string) (int64, int64, error) {
	var products, children int64
	if err := r.db.WithContext(ctx).Model(&productModel{}).Where("category_id = ?", id).Count(&products).Error; err != nil {
		return 0, 0, err
	}
	if err := r.db.WithContext(ctx).Model(&categoryModel{}).Where("parent_id = ?", id).Count(&children).Error; err != nil {
		return 0, 0, err
	}
	return products, children, nil
}

func (r *CatalogRepository) ListCollections(ctx context.Context) ([]catalog.Collection, error) {
	var records []collectionModel
	if err := r.db.WithContext(ctx).Order("id ASC").Find(&records).Error; err != nil {
		return nil, err
	}
	collections := make([]catalog.Collection, 0, len(records))
	for _, record := range records {
		collections = append(collections, catalog.Collection{ID: record.ID, Slug: record.Slug, Name: record.Name})
	}
	return collections, nil
}

func (r *CatalogRepository) SaveCollection(ctx context.Context, c catalog.Collection, create bool) error {
	record := collectionModel{ID: c.ID, Slug: c.Slug, Name: c.Name}
	if create {
		return mapCatalogWriteError(r.db.WithContext(ctx).Create(&record).Error)
	}
	result := r.db.WithContext(ctx).Model(&collectionModel{}).Where("id = ?", c.ID).Updates(map[string]any{
		"slug": c.Slug, "name": c.Name,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return catalog.ErrNotFound
	}
	return nil
}

func (r *CatalogRepository) DeleteCollection(ctx context.Context, id string) error {
	return requireCatalogAffected(r.db.WithContext(ctx).Delete(&collectionModel{}, "id = ?", id))
}

// CountCollectionReferences counts products whose collectionIds JSON array
// contains the id. Uses MySQL JSON_CONTAINS on the JSON column.
func (r *CatalogRepository) CountCollectionReferences(ctx context.Context, id string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&productModel{}).
		Where("JSON_CONTAINS(collection_ids, JSON_QUOTE(?))", id).Count(&count).Error
	return count, err
}

func (r *CatalogRepository) ListSizeScales(ctx context.Context) ([]catalog.SizeScale, error) {
	var records []sizeScaleModel
	if err := r.db.WithContext(ctx).Order("id ASC").Find(&records).Error; err != nil {
		return nil, err
	}
	scales := make([]catalog.SizeScale, 0, len(records))
	for _, record := range records {
		scales = append(scales, catalog.SizeScale{ID: record.ID, Sizes: orEmpty(record.Sizes)})
	}
	return scales, nil
}

// SeedCatalog upserts a full catalog dataset. Idempotent: re-running replaces
// rows by primary key. Used by cmd/seed.
func (r *CatalogRepository) SeedCatalog(
	ctx context.Context,
	products []catalog.Product,
	colorways []catalog.Colorway,
	skus []catalog.Sku,
	categories []catalog.Category,
	collections []catalog.Collection,
	sizeScales []catalog.SizeScale,
) error {
	upsert := clause.OnConflict{UpdateAll: true}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, c := range categories {
			record := categoryModel{ID: c.ID, Slug: c.Slug, Name: c.Name, ParentID: c.ParentID, Gender: c.Gender, Level: c.Level}
			if err := tx.Clauses(upsert).Create(&record).Error; err != nil {
				return err
			}
		}
		for _, c := range collections {
			record := collectionModel{ID: c.ID, Slug: c.Slug, Name: c.Name}
			if err := tx.Clauses(upsert).Create(&record).Error; err != nil {
				return err
			}
		}
		for _, s := range sizeScales {
			record := sizeScaleModel{ID: s.ID, Sizes: orEmpty(s.Sizes)}
			if err := tx.Clauses(upsert).Create(&record).Error; err != nil {
				return err
			}
		}
		for _, p := range products {
			record := fromProduct(p)
			if err := tx.Clauses(upsert).Create(&record).Error; err != nil {
				return err
			}
		}
		for _, cw := range colorways {
			record := fromColorway(cw)
			if err := tx.Clauses(upsert).Create(&record).Error; err != nil {
				return err
			}
		}
		for _, sku := range skus {
			record := fromSku(sku)
			if err := tx.Clauses(upsert).Create(&record).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func mapCatalogNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return catalog.ErrNotFound
	}
	return err
}

func mapCatalogWriteError(err error) error {
	if err == nil {
		return nil
	}
	if isDuplicateKey(err) {
		return catalog.ErrConflict
	}
	return err
}

func requireCatalogAffected(result *gorm.DB) error {
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return catalog.ErrNotFound
	}
	return nil
}

var _ catalog.Repository = (*CatalogRepository)(nil)
