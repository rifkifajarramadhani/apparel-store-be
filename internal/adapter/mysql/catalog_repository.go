package mysqladapter

import (
	"context"
	"errors"
	"fmt"
	"strings"
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
	ColourwayID                             string
	SkuID                                   string
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
	assetSQL := `SELECT COALESCE(a.product_id, sk.product_id) product_id, a.public_id, a.media_type, a.cdn_url url,
		COALESCE(a.alt_text,'') alt_text, a.role, a.sort_order,
		COALESCE(c.public_id,'') colourway_id, COALESCE(sk.public_id,'') sku_id
	FROM assets a
	LEFT JOIN colourways c ON c.id=a.colourway_id AND c.archived_at IS NULL
	LEFT JOIN skus sk ON sk.id=a.sku_id AND sk.archived_at IS NULL
	WHERE a.archived_at IS NULL AND (a.product_id IN ? OR sk.product_id IN ?)
	ORDER BY a.role,a.sort_order`
	if err := r.db.WithContext(ctx).Raw(assetSQL, ids, ids).Scan(&assets).Error; err != nil {
		return err
	}
	for _, row := range assets {
		i := positions[row.ProductID]
		products[i].Assets = append(products[i].Assets, catalog.Asset{ID: row.PublicID, MediaType: row.MediaType, URL: row.URL, AltText: row.AltText, Role: row.Role, SortOrder: row.SortOrder, ColourwayID: row.ColourwayID, SkuID: row.SkuID})
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

func (r *CatalogRepository) CreateProduct(ctx context.Context, in catalog.ProductAggregate) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		exists, err := productExists(tx, in.Product.ID)
		if err != nil {
			return err
		}
		if exists {
			return catalog.ErrConflict
		}
		return saveAggregate(tx, in)
	})
}

func (r *CatalogRepository) UpdateProduct(ctx context.Context, in catalog.ProductAggregate) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		exists, err := productExists(tx, in.Product.ID)
		if err != nil {
			return err
		}
		if !exists {
			return catalog.ErrNotFound
		}
		return saveAggregate(tx, in)
	})
}

func (r *CatalogRepository) DeleteProduct(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Exec("UPDATE products SET archived_at=NOW(6) WHERE (style_code=? OR public_id=? OR slug=?) AND archived_at IS NULL", id, id, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return catalog.ErrNotFound
	}
	return nil
}

func productExists(tx *gorm.DB, styleCode string) (bool, error) {
	var count int64
	err := tx.Table("products").Where("style_code = ?", styleCode).Count(&count).Error
	return count > 0, err
}

// saveAggregate upserts a product and its children in one transaction, mirroring
// the seeder's per-product loop. Products and SKUs are upserted by business code
// (preserving public_id and existing order/inventory references); the product's
// assets, prices, product_categories are delete-and-reinserted to reflect edits.
// SKU on_hand is left untouched on update — inventory is owned by SetInventory.
func saveAggregate(tx *gorm.DB, in catalog.ProductAggregate) error {
	brandID, err := resolveBrand(tx, in.Product.Brand)
	if err != nil {
		return err
	}
	categoryID, err := resolveCategory(tx, in.Product.CategorySlug)
	if err != nil {
		return err
	}
	published, _ := time.Parse("2006-01-02", in.Product.PublishedAt)

	if err := tx.Exec(`INSERT INTO products(public_id,style_code,slug,brand_id,name,subtitle,gender,product_type,description,published_at)
		VALUES(?,?,?,?,?,?,?,?,?,?)
		ON DUPLICATE KEY UPDATE slug=VALUES(slug),brand_id=VALUES(brand_id),name=VALUES(name),subtitle=VALUES(subtitle),gender=VALUES(gender),product_type=VALUES(product_type),description=VALUES(description),published_at=VALUES(published_at),archived_at=NULL`,
		seedPublicID("PR", in.Product.ID), in.Product.ID, in.Product.Slug, brandID, in.Product.Name, in.Product.Subtitle, in.Product.Gender, in.Product.ProductType, in.Product.Description, published).Error; err != nil {
		return err
	}
	productID, err := seedRowID(tx, "products", "style_code", in.Product.ID)
	if err != nil {
		return err
	}

	colourIDs := make(map[string]uint64, len(in.Colourways))
	for _, c := range in.Colourways {
		id, err := resolveColourway(tx, c)
		if err != nil {
			return err
		}
		colourIDs[c.ID] = id
	}

	if err := tx.Exec("DELETE FROM product_categories WHERE product_id=?", productID).Error; err != nil {
		return err
	}
	if categoryID != 0 {
		if err := tx.Exec("INSERT INTO product_categories(product_id,category_id,is_primary) VALUES(?,?,TRUE)", productID, categoryID).Error; err != nil {
			return err
		}
	}

	// Prices are fully rewritten from the payload; drop the product's own price
	// and any of its SKUs' prices before reinserting.
	if err := tx.Exec("DELETE FROM prices WHERE product_id=? OR sku_id IN (SELECT id FROM (SELECT id FROM skus WHERE product_id=?) t)", productID, productID).Error; err != nil {
		return err
	}
	if err := tx.Exec("INSERT INTO prices(public_id,product_id,currency,amount,valid_from) VALUES(?,?,'IDR',?,'1970-01-01')", seedPublicID("PP", in.Product.ID), productID, in.Product.BasePrice).Error; err != nil {
		return err
	}

	codes := make([]string, 0, len(in.Skus))
	for _, sku := range in.Skus {
		sizeID, err := resolveSize(tx, sku.SizeScale, sku.Size)
		if err != nil {
			return err
		}
		if err := tx.Exec(`INSERT INTO skus(public_id,sku_code,product_id,colourway_id,size_id,on_hand)
			VALUES(?,?,?,?,?,?)
			ON DUPLICATE KEY UPDATE product_id=VALUES(product_id),colourway_id=VALUES(colourway_id),size_id=VALUES(size_id),archived_at=NULL`,
			seedPublicID("SK", sku.ID), sku.ID, productID, colourIDs[sku.ColourwayID], sizeID, sku.StockQty).Error; err != nil {
			return err
		}
		skuID, err := seedRowID(tx, "skus", "sku_code", sku.ID)
		if err != nil {
			return err
		}
		if sku.Price > 0 {
			if err := tx.Exec("INSERT INTO prices(public_id,sku_id,currency,amount,valid_from) VALUES(?,?,'IDR',?,'1970-01-01')", seedPublicID("SP", sku.ID), skuID, sku.Price).Error; err != nil {
				return err
			}
		}
		codes = append(codes, sku.ID)
	}
	// Archive SKUs the editor removed from the product.
	if err := tx.Exec("UPDATE skus SET archived_at=NOW(6) WHERE product_id=? AND sku_code NOT IN ? AND archived_at IS NULL", productID, codes).Error; err != nil {
		return err
	}

	if err := tx.Exec("DELETE FROM assets WHERE product_id=? OR sku_id IN (SELECT id FROM (SELECT id FROM skus WHERE product_id=?) t)", productID, productID).Error; err != nil {
		return err
	}
	seen := make(map[string]struct{})
	for order, image := range in.Images {
		if image.URL == "" {
			continue
		}
		if _, dup := seen[image.URL]; dup {
			continue
		}
		seen[image.URL] = struct{}{}
		var colourwayID any
		if image.ColourwayID != "" {
			if id, ok := colourIDs[image.ColourwayID]; ok {
				colourwayID = id
			}
		}
		if err := tx.Exec("INSERT INTO assets(public_id,product_id,colourway_id,media_type,storage_provider,cdn_url,alt_text,role,sort_order) VALUES(?,?,?,'image','external',?,?,'product_image',?)", seedPublicID("AS", image.URL), productID, colourwayID, image.URL, "", order).Error; err != nil {
			return err
		}
	}
	return nil
}

func resolveBrand(tx *gorm.DB, name string) (uint64, error) {
	name = strings.TrimSpace(name)
	slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	id, err := seedRowID(tx, "brands", "slug", slug)
	if err != nil {
		return 0, err
	}
	if id != 0 {
		return id, nil
	}
	if err := tx.Exec("INSERT INTO brands(public_id,slug,name) VALUES(?,?,?)", seedPublicID("BR", name), slug, name).Error; err != nil {
		return 0, err
	}
	return seedRowID(tx, "brands", "slug", slug)
}

func resolveCategory(tx *gorm.DB, slug string) (uint64, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return 0, nil
	}
	id, err := seedRowID(tx, "categories", "slug", slug)
	if err != nil {
		return 0, err
	}
	if id == 0 {
		return 0, fmt.Errorf("%w: unknown category %q", catalog.ErrInvalidInput, slug)
	}
	return id, nil
}

func resolveSize(tx *gorm.DB, scaleCode, code string) (uint64, error) {
	var id uint64
	if err := tx.Raw("SELECT sz.id FROM sizes sz JOIN size_scales ss ON ss.id=sz.size_scale_id WHERE ss.code=? AND sz.code=?", scaleCode, code).Scan(&id).Error; err != nil {
		return 0, err
	}
	if id == 0 {
		return 0, fmt.Errorf("%w: unknown size %q in scale %q", catalog.ErrInvalidInput, code, scaleCode)
	}
	return id, nil
}

func resolveColourway(tx *gorm.DB, c catalog.ColourwayWrite) (uint64, error) {
	slug := strings.ToLower(strings.TrimSpace(c.ID))
	id, err := seedRowID(tx, "colourways", "slug", slug)
	if err != nil {
		return 0, err
	}
	if id != 0 {
		if err := tx.Exec("UPDATE colourways SET name=?,colour_family=?,hex_code=?,archived_at=NULL WHERE id=?", c.Name, c.ColorFamily, c.SwatchHex, id).Error; err != nil {
			return 0, err
		}
		return id, nil
	}
	if err := tx.Exec("INSERT INTO colourways(public_id,slug,name,colour_family,hex_code) VALUES(?,?,?,?,?)", seedPublicID("CO", c.ID), slug, c.Name, c.ColorFamily, c.SwatchHex).Error; err != nil {
		return 0, err
	}
	return seedRowID(tx, "colourways", "slug", slug)
}

var _ catalog.Repository = (*CatalogRepository)(nil)
