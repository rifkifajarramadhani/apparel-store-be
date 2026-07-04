package mysqladapter

import (
	"context"
	"errors"
	"time"

	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/catalog"
	"gorm.io/gorm"
)

type CatalogRepository struct{ db *gorm.DB }

func NewCatalogRepository(db *gorm.DB) *CatalogRepository {
	return &CatalogRepository{db: db}
}

type productRow struct {
	ID, PublicID, StyleCode, Slug, Name, Subtitle, Gender, ProductType, Description string
	BrandPublicID, BrandSlug, BrandName                                             string
}

func (r *CatalogRepository) ListProducts(ctx context.Context, q catalog.ProductQuery) (catalog.CursorPage[catalog.Product], error) {
	query := r.db.WithContext(ctx).Table("products p").
		Select("p.id, p.public_id, p.style_code, p.slug, p.name, p.subtitle, p.gender, p.product_type, p.description, b.public_id brand_public_id, b.slug brand_slug, b.name brand_name").
		Joins("JOIN brands b ON b.id = p.brand_id AND b.archived_at IS NULL").
		Where("p.archived_at IS NULL")
	if q.Cursor != "" {
		query = query.Where("p.public_id > ?", q.Cursor)
	}
	if q.BrandSlug != "" {
		query = query.Where("b.slug = ?", q.BrandSlug)
	}
	if q.CategorySlug != "" {
		query = query.Joins("JOIN product_categories pc_filter ON pc_filter.product_id = p.id").Joins("JOIN categories c_filter ON c_filter.id = pc_filter.category_id AND c_filter.archived_at IS NULL").Where("c_filter.slug = ?", q.CategorySlug)
	}
	if q.Query != "" {
		like := "%" + q.Query + "%"
		query = query.Where("p.name LIKE ? OR p.subtitle LIKE ? OR p.style_code LIKE ?", like, like, like)
	}
	var rows []productRow
	if err := query.Order("p.public_id ASC").Limit(q.Limit + 1).Scan(&rows).Error; err != nil {
		return catalog.CursorPage[catalog.Product]{}, err
	}
	hasMore := len(rows) > q.Limit
	if hasMore {
		rows = rows[:q.Limit]
	}
	products, ids := make([]catalog.Product, len(rows)), make([]uint64, len(rows))
	for i, row := range rows {
		products[i] = productFromRow(row)
		ids[i] = parseUintID(row.ID)
	}
	if err := r.hydrateProducts(ctx, products, ids, q.Currency); err != nil {
		return catalog.CursorPage[catalog.Product]{}, err
	}
	page := catalog.CursorPage[catalog.Product]{Items: products}
	if hasMore {
		page.NextCursor = products[len(products)-1].ID
	}
	return page, nil
}

func (r *CatalogRepository) GetProduct(ctx context.Context, id, currency string) (catalog.Product, error) {
	var row productRow
	err := r.db.WithContext(ctx).Table("products p").Select("p.id, p.public_id, p.style_code, p.slug, p.name, p.subtitle, p.gender, p.product_type, p.description, b.public_id brand_public_id, b.slug brand_slug, b.name brand_name").Joins("JOIN brands b ON b.id = p.brand_id AND b.archived_at IS NULL").Where("p.archived_at IS NULL AND (p.public_id = ? OR p.slug = ? OR p.style_code = ?)", id, id, id).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return catalog.Product{}, catalog.ErrNotFound
	}
	if err != nil {
		return catalog.Product{}, err
	}
	products := []catalog.Product{productFromRow(row)}
	if err := r.hydrateProducts(ctx, products, []uint64{parseUintID(row.ID)}, currency); err != nil {
		return catalog.Product{}, err
	}
	return products[0], nil
}

func productFromRow(row productRow) catalog.Product {
	return catalog.Product{ID: row.PublicID, StyleCode: row.StyleCode, Slug: row.Slug, Name: row.Name, Subtitle: row.Subtitle, Gender: row.Gender, ProductType: row.ProductType, Description: row.Description, Brand: catalog.Brand{ID: row.BrandPublicID, Slug: row.BrandSlug, Name: row.BrandName}, Categories: []catalog.Category{}, Colourways: []catalog.Colourway{}, Sizes: []catalog.Size{}, Assets: []catalog.Asset{}}
}

func parseUintID(raw string) uint64 {
	var id uint64
	for _, c := range raw {
		if c >= '0' && c <= '9' {
			id = id*10 + uint64(c-'0')
		}
	}
	return id
}

type categoryLinkRow struct {
	ProductID            uint64
	PublicID, Slug, Name string
	ParentPublicID       *string
}
type colourLinkRow struct {
	ProductID                                   uint64
	PublicID, Slug, Name, ColourFamily, HexCode string
}
type sizeLinkRow struct {
	ProductID                       uint64
	PublicID, ScaleCode, Code, Name string
	SortOrder                       int
}
type assetLinkRow struct {
	ProductID                               uint64
	PublicID, MediaType, URL, AltText, Role string
	SortOrder                               int
}
type priceRangeRow struct {
	ProductID            uint64
	MinAmount, MaxAmount int64
}

func (r *CatalogRepository) hydrateProducts(ctx context.Context, products []catalog.Product, ids []uint64, currency string) error {
	if len(ids) == 0 {
		return nil
	}
	positions := make(map[uint64]int, len(ids))
	for i, id := range ids {
		positions[id] = i
	}
	var categories []categoryLinkRow
	if err := r.db.WithContext(ctx).Raw("SELECT pc.product_id, c.public_id, c.slug, c.name, parent.public_id parent_public_id FROM product_categories pc JOIN categories c ON c.id=pc.category_id AND c.archived_at IS NULL LEFT JOIN categories parent ON parent.id=c.parent_id WHERE pc.product_id IN ? ORDER BY pc.is_primary DESC,c.name", ids).Scan(&categories).Error; err != nil {
		return err
	}
	for _, row := range categories {
		i := positions[row.ProductID]
		products[i].Categories = append(products[i].Categories, catalog.Category{ID: row.PublicID, ParentID: row.ParentPublicID, Slug: row.Slug, Name: row.Name})
	}
	var colours []colourLinkRow
	if err := r.db.WithContext(ctx).Raw("SELECT DISTINCT s.product_id,c.public_id,c.slug,c.name,c.colour_family,c.hex_code FROM skus s JOIN colourways c ON c.id=s.colourway_id AND c.archived_at IS NULL WHERE s.archived_at IS NULL AND s.product_id IN ? ORDER BY c.name", ids).Scan(&colours).Error; err != nil {
		return err
	}
	for _, row := range colours {
		i := positions[row.ProductID]
		products[i].Colourways = append(products[i].Colourways, catalog.Colourway{ID: row.PublicID, Slug: row.Slug, Name: row.Name, ColourFamily: row.ColourFamily, HexCode: row.HexCode})
	}
	var sizes []sizeLinkRow
	if err := r.db.WithContext(ctx).Raw("SELECT DISTINCT s.product_id,sz.public_id,ss.code scale_code,sz.code,sz.name,sz.sort_order FROM skus s JOIN sizes sz ON sz.id=s.size_id AND sz.archived_at IS NULL JOIN size_scales ss ON ss.id=sz.size_scale_id AND ss.archived_at IS NULL WHERE s.archived_at IS NULL AND s.product_id IN ? ORDER BY sz.sort_order,sz.name", ids).Scan(&sizes).Error; err != nil {
		return err
	}
	for _, row := range sizes {
		i := positions[row.ProductID]
		products[i].Sizes = append(products[i].Sizes, catalog.Size{ID: row.PublicID, ScaleCode: row.ScaleCode, Code: row.Code, Name: row.Name, SortOrder: row.SortOrder})
	}
	var assets []assetLinkRow
	if err := r.db.WithContext(ctx).Raw("SELECT a.product_id,a.public_id,a.media_type,a.cdn_url url,COALESCE(a.alt_text,'') alt_text,a.role,a.sort_order FROM assets a WHERE a.archived_at IS NULL AND a.product_id IN ? ORDER BY a.role,a.sort_order", ids).Scan(&assets).Error; err != nil {
		return err
	}
	for _, row := range assets {
		i := positions[row.ProductID]
		products[i].Assets = append(products[i].Assets, catalog.Asset{ID: row.PublicID, MediaType: row.MediaType, URL: row.URL, AltText: row.AltText, Role: row.Role, SortOrder: row.SortOrder})
	}
	var ranges []priceRangeRow
	now := time.Now().UTC()
	priceSQL := `SELECT s.product_id, MIN(COALESCE(sp.amount,pp.amount)) min_amount, MAX(COALESCE(sp.amount,pp.amount)) max_amount FROM skus s LEFT JOIN prices sp ON sp.sku_id=s.id AND sp.currency=? AND sp.archived_at IS NULL AND sp.valid_from<=? AND (sp.valid_to IS NULL OR sp.valid_to>?) LEFT JOIN prices pp ON pp.product_id=s.product_id AND pp.currency=? AND pp.archived_at IS NULL AND pp.valid_from<=? AND (pp.valid_to IS NULL OR pp.valid_to>?) WHERE s.archived_at IS NULL AND s.product_id IN ? GROUP BY s.product_id`
	if err := r.db.WithContext(ctx).Raw(priceSQL, currency, now, now, currency, now, now, ids).Scan(&ranges).Error; err != nil {
		return err
	}
	for _, row := range ranges {
		i := positions[row.ProductID]
		min, max := catalog.Money{Currency: currency, Amount: row.MinAmount}, catalog.Money{Currency: currency, Amount: row.MaxAmount}
		products[i].MinPrice, products[i].MaxPrice = &min, &max
	}
	return nil
}

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
	err := r.db.WithContext(ctx).Table("colourways").Select("public_id id,slug,name,colour_family,hex_code").Where("archived_at IS NULL").Order("name").Scan(&out).Error
	return out, err
}
func (r *CatalogRepository) ListSizes(ctx context.Context) ([]catalog.Size, error) {
	var out []catalog.Size
	err := r.db.WithContext(ctx).Raw("SELECT sz.public_id id,ss.code scale_code,sz.code,sz.name,sz.sort_order FROM sizes sz JOIN size_scales ss ON ss.id=sz.size_scale_id WHERE sz.archived_at IS NULL AND ss.archived_at IS NULL ORDER BY ss.code,sz.sort_order").Scan(&out).Error
	return out, err
}

type skuRow struct {
	ID, Code, Barcode, ProductID                                     string
	ColourwayID, ColourwaySlug, ColourwayName, ColourFamily, HexCode string
	SizeID, ScaleCode, SizeCode, SizeName                            string
	SortOrder, OnHand, Reserved                                      int
	Amount                                                           int64
	CompareAtAmount                                                  *int64
}

func (r *CatalogRepository) ListSkus(ctx context.Context, q catalog.SkuQuery) (catalog.CursorPage[catalog.Sku], error) {
	now := time.Now().UTC()
	sql := `SELECT s.public_id id,s.sku_code code,COALESCE(s.barcode,'') barcode,p.style_code product_id,c.public_id colourway_id,c.slug colourway_slug,c.name colourway_name,COALESCE(c.colour_family,'') colour_family,c.hex_code,sz.public_id size_id,ss.code scale_code,sz.code size_code,sz.name size_name,sz.sort_order,s.on_hand,s.reserved,COALESCE((SELECT sp.amount FROM prices sp WHERE sp.sku_id=s.id AND sp.currency=? AND sp.archived_at IS NULL AND sp.valid_from<=? AND (sp.valid_to IS NULL OR sp.valid_to>?) ORDER BY sp.valid_from DESC LIMIT 1),(SELECT pp.amount FROM prices pp WHERE pp.product_id=s.product_id AND pp.currency=? AND pp.archived_at IS NULL AND pp.valid_from<=? AND (pp.valid_to IS NULL OR pp.valid_to>?) ORDER BY pp.valid_from DESC LIMIT 1),0) amount FROM skus s JOIN products p ON p.id=s.product_id AND p.archived_at IS NULL JOIN colourways c ON c.id=s.colourway_id AND c.archived_at IS NULL JOIN sizes sz ON sz.id=s.size_id AND sz.archived_at IS NULL JOIN size_scales ss ON ss.id=sz.size_scale_id WHERE s.archived_at IS NULL`
	args := []any{q.Currency, now, now, q.Currency, now, now}
	if q.ProductID != "" {
		sql += " AND (p.public_id=? OR p.style_code=?)"
		args = append(args, q.ProductID, q.ProductID)
	}
	if q.ColourwayID != "" {
		sql += " AND c.public_id=?"
		args = append(args, q.ColourwayID)
	}
	if q.Cursor != "" {
		sql += " AND s.public_id>?"
		args = append(args, q.Cursor)
	}
	sql += " ORDER BY s.public_id LIMIT ?"
	args = append(args, q.Limit+1)
	var rows []skuRow
	if err := r.db.WithContext(ctx).Raw(sql, args...).Scan(&rows).Error; err != nil {
		return catalog.CursorPage[catalog.Sku]{}, err
	}
	hasMore := len(rows) > q.Limit
	if hasMore {
		rows = rows[:q.Limit]
	}
	items := make([]catalog.Sku, 0, len(rows))
	for _, row := range rows {
		items = append(items, catalog.Sku{ID: row.ID, Code: row.Code, Barcode: row.Barcode, ProductID: row.ProductID, Colourway: catalog.Colourway{ID: row.ColourwayID, Slug: row.ColourwaySlug, Name: row.ColourwayName, ColourFamily: row.ColourFamily, HexCode: row.HexCode}, Size: catalog.Size{ID: row.SizeID, ScaleCode: row.ScaleCode, Code: row.SizeCode, Name: row.SizeName, SortOrder: row.SortOrder}, Price: catalog.Money{Currency: q.Currency, Amount: row.Amount, CompareAtAmount: row.CompareAtAmount}, OnHand: row.OnHand, Reserved: row.Reserved, Available: row.OnHand - row.Reserved, Assets: []catalog.Asset{}})
	}
	if err := r.hydrateSkuAssets(ctx, items); err != nil {
		return catalog.CursorPage[catalog.Sku]{}, err
	}
	page := catalog.CursorPage[catalog.Sku]{Items: items}
	if hasMore {
		page.NextCursor = items[len(items)-1].ID
	}
	return page, nil
}

func (r *CatalogRepository) SetInventory(ctx context.Context, in catalog.InventoryAdjustment) error {
	result := r.db.WithContext(ctx).Exec("UPDATE skus SET on_hand=?, reserved=? WHERE public_id=? AND archived_at IS NULL", in.OnHand, in.Reserved, in.SkuID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 0 {
		return nil
	}
	var count int64
	if err := r.db.WithContext(ctx).Table("skus").Where("public_id=? AND archived_at IS NULL", in.SkuID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return catalog.ErrNotFound
	}
	return nil
}

type skuAssetRow struct {
	SkuID                                   string
	PublicID, MediaType, URL, AltText, Role string
	SortOrder                               int
}

func (r *CatalogRepository) hydrateSkuAssets(ctx context.Context, skus []catalog.Sku) error {
	if len(skus) == 0 {
		return nil
	}
	positions := make(map[string]int, len(skus))
	ids := make([]string, len(skus))
	for i := range skus {
		ids[i] = skus[i].ID
		positions[skus[i].ID] = i
	}
	var assets []skuAssetRow
	query := "SELECT s.public_id sku_id,a.public_id,a.media_type,a.cdn_url url,COALESCE(a.alt_text,'') alt_text,a.role,a.sort_order FROM assets a JOIN skus s ON s.id=a.sku_id WHERE a.archived_at IS NULL AND s.public_id IN ? ORDER BY a.role,a.sort_order"
	if err := r.db.WithContext(ctx).Raw(query, ids).Scan(&assets).Error; err != nil {
		return err
	}
	for _, row := range assets {
		i := positions[row.SkuID]
		skus[i].Assets = append(skus[i].Assets, catalog.Asset{ID: row.PublicID, MediaType: row.MediaType, URL: row.URL, AltText: row.AltText, Role: row.Role, SortOrder: row.SortOrder})
	}
	return nil
}

var _ catalog.Repository = (*CatalogRepository)(nil)
