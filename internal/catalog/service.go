package catalog

import (
	"context"
	"fmt"
	"strings"
)

const (
	DefaultLimit = 24
	MaxLimit     = 100
)

// Repository is the persistence port. Defined at the consumption site.
type Repository interface {
	ListProducts(ctx context.Context, q ProductQuery) ([]Product, int64, error)
	GetProduct(ctx context.Context, id string) (Product, error)
	ProductExists(ctx context.Context, id string) (bool, error)
	GetAggregate(ctx context.Context, productID string) (Aggregate, error)
	// SaveAggregate replaces the product and all of its colorways/skus in one tx.
	SaveAggregate(ctx context.Context, agg Aggregate) error
	DeleteProduct(ctx context.Context, id string) error

	ListColorways(ctx context.Context, productID string, embedSkus bool) ([]Colorway, error)
	ListSkus(ctx context.Context, productID, colorwayID string) ([]Sku, error)
	GetSku(ctx context.Context, id string) (Sku, error)
	UpdateSku(ctx context.Context, sku Sku) error

	ListCategories(ctx context.Context) ([]Category, error)
	SaveCategory(ctx context.Context, c Category, create bool) error
	DeleteCategory(ctx context.Context, id string) error
	CountCategoryReferences(ctx context.Context, id string) (products, children int64, err error)

	ListCollections(ctx context.Context) ([]Collection, error)
	SaveCollection(ctx context.Context, c Collection, create bool) error
	DeleteCollection(ctx context.Context, id string) error
	CountCollectionReferences(ctx context.Context, id string) (int64, error)

	ListSizeScales(ctx context.Context) ([]SizeScale, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListProducts(ctx context.Context, q ProductQuery) ([]Product, int64, error) {
	if q.Limit > MaxLimit {
		q.Limit = MaxLimit
	}
	return s.repo.ListProducts(ctx, q)
}

func (s *Service) GetProduct(ctx context.Context, id string) (Product, error) {
	return s.repo.GetProduct(ctx, id)
}

func (s *Service) GetAggregate(ctx context.Context, id string) (Aggregate, error) {
	return s.repo.GetAggregate(ctx, id)
}

func (s *Service) ListColorways(ctx context.Context, productID string, embedSkus bool) ([]Colorway, error) {
	return s.repo.ListColorways(ctx, productID, embedSkus)
}

func (s *Service) ListSkus(ctx context.Context, productID, colorwayID string) ([]Sku, error) {
	return s.repo.ListSkus(ctx, productID, colorwayID)
}

func (s *Service) ListCategories(ctx context.Context) ([]Category, error) {
	return s.repo.ListCategories(ctx)
}

func (s *Service) ListCollections(ctx context.Context) ([]Collection, error) {
	return s.repo.ListCollections(ctx)
}

func (s *Service) ListSizeScales(ctx context.Context) ([]SizeScale, error) {
	return s.repo.ListSizeScales(ctx)
}

// UpdateSku edits a single SKU (inventory). The id path must match the body.
func (s *Service) UpdateSku(ctx context.Context, id string, sku Sku) (Sku, error) {
	if id != sku.ID {
		return Sku{}, fmt.Errorf("%w: sku id is immutable", ErrInvalidInput)
	}
	if _, err := s.repo.GetSku(ctx, id); err != nil {
		return Sku{}, err
	}
	if sku.StockQty < 0 {
		return Sku{}, fmt.Errorf("%w: stockQty must not be negative", ErrInvalidInput)
	}
	sku.InStock = sku.StockQty > 0
	if err := s.repo.UpdateSku(ctx, sku); err != nil {
		return Sku{}, err
	}
	return sku, nil
}

// CreateProduct persists a new product aggregate. Fails if the id already exists.
func (s *Service) CreateProduct(ctx context.Context, agg Aggregate) (Aggregate, error) {
	exists, err := s.repo.ProductExists(ctx, agg.Product.ID)
	if err != nil {
		return Aggregate{}, err
	}
	if exists {
		return Aggregate{}, fmt.Errorf("%w: product id already exists", ErrConflict)
	}
	return s.saveAggregate(ctx, agg, nil)
}

// UpdateProduct replaces an existing product aggregate. The id is immutable.
func (s *Service) UpdateProduct(ctx context.Context, id string, agg Aggregate) (Aggregate, error) {
	if id != agg.Product.ID {
		return Aggregate{}, fmt.Errorf("%w: product id is immutable", ErrInvalidInput)
	}
	existing, err := s.repo.GetProduct(ctx, id)
	if err != nil {
		return Aggregate{}, err
	}
	return s.saveAggregate(ctx, agg, &existing)
}

func (s *Service) DeleteProduct(ctx context.Context, id string) error {
	if _, err := s.repo.GetProduct(ctx, id); err != nil {
		return err
	}
	return s.repo.DeleteProduct(ctx, id)
}

func (s *Service) saveAggregate(ctx context.Context, agg Aggregate, existing *Product) (Aggregate, error) {
	derived, err := Derive(agg, existing)
	if err != nil {
		return Aggregate{}, err
	}
	agg.Product = derived
	if err := s.repo.SaveAggregate(ctx, agg); err != nil {
		return Aggregate{}, err
	}
	return agg, nil
}

func (s *Service) SaveCategory(ctx context.Context, c Category, create bool) error {
	return s.repo.SaveCategory(ctx, c, create)
}

func (s *Service) DeleteCategory(ctx context.Context, id string) error {
	products, children, err := s.repo.CountCategoryReferences(ctx, id)
	if err != nil {
		return err
	}
	if products > 0 || children > 0 {
		return fmt.Errorf("%w: category is still referenced", ErrReferenced)
	}
	return s.repo.DeleteCategory(ctx, id)
}

func (s *Service) SaveCollection(ctx context.Context, c Collection, create bool) error {
	return s.repo.SaveCollection(ctx, c, create)
}

func (s *Service) DeleteCollection(ctx context.Context, id string) error {
	products, err := s.repo.CountCollectionReferences(ctx, id)
	if err != nil {
		return err
	}
	if products > 0 {
		return fmt.Errorf("%w: collection is still referenced", ErrReferenced)
	}
	return s.repo.DeleteCollection(ctx, id)
}

// Derive validates a product aggregate and computes the card/summary fields
// off its colorways and skus. Ported from scripts/api-server.mjs deriveProduct.
// existing carries the previously stored product on update (nil on create) so
// non-Sale badges are retained.
func Derive(agg Aggregate, existing *Product) (Product, error) {
	product, colorways, skus := agg.Product, agg.Colorways, agg.Skus
	if product.ID == "" || len(colorways) == 0 || len(skus) == 0 {
		return Product{}, fmt.Errorf("%w: product, at least one colorway, and SKUs are required", ErrInvalidInput)
	}

	var defaults []Colorway
	for _, cw := range colorways {
		if cw.IsDefault {
			defaults = append(defaults, cw)
		}
		if cw.ProductID != product.ID {
			return Product{}, fmt.Errorf("%w: colorways must belong to the product", ErrInvalidInput)
		}
	}
	if len(defaults) != 1 {
		return Product{}, fmt.Errorf("%w: exactly one colorway must be the default", ErrInvalidInput)
	}
	colorwayIDs := make(map[string]bool, len(colorways))
	for _, cw := range colorways {
		colorwayIDs[cw.ID] = true
	}
	for _, sku := range skus {
		if sku.ProductID != product.ID || !colorwayIDs[sku.ColorwayID] {
			return Product{}, fmt.Errorf("%w: SKUs must belong to a submitted colorway", ErrInvalidInput)
		}
	}

	selected := defaults[0]
	minPrice, maxPrice := colorways[0].Price, colorways[0].Price
	anyOnSale := false
	for _, cw := range colorways {
		if cw.Price < minPrice {
			minPrice = cw.Price
		}
		if cw.Price > maxPrice {
			maxPrice = cw.Price
		}
		if cw.OnSale {
			anyOnSale = true
		}
	}

	badges := make([]string, 0)
	if existing != nil {
		for _, b := range existing.Badges {
			if b != "Sale" {
				badges = append(badges, b)
			}
		}
	}
	if anyOnSale {
		badges = append(badges, "Sale")
	}

	product.MinPrice = minPrice
	product.MaxPrice = maxPrice
	product.Badges = badges
	product.ColorwayCount = len(colorways)
	product.ColorFamilies = uniqueStrings(mapColorways(colorways, func(c Colorway) string { return c.ColorFamily }))
	product.Swatches = mapColorways(colorways, func(c Colorway) Swatch {
		return Swatch{StyleColor: c.StyleColor, Hex: c.SwatchHex}
	})
	if len(selected.Images) > 0 {
		product.ThumbnailURL = selected.Images[0]
		if len(selected.Images) > 1 {
			product.HoverImageURL = selected.Images[1]
		} else {
			product.HoverImageURL = selected.Images[0]
		}
	}
	product.DefaultColorwayID = selected.ID
	sizes := make([]string, 0, len(skus))
	for _, sku := range skus {
		sizes = append(sizes, sku.Size)
	}
	product.Sizes = uniqueStrings(sizes)
	return product, nil
}

func mapColorways[T any](colorways []Colorway, fn func(Colorway) T) []T {
	out := make([]T, 0, len(colorways))
	for _, cw := range colorways {
		out = append(out, fn(cw))
	}
	return out
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	out := make([]string, 0, len(values))
	for _, v := range values {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

// NormalizeSort maps the frontend's _sort/_order into a validated column.
func NormalizeSort(sortBy, order string) (string, bool) {
	switch strings.TrimSpace(sortBy) {
	case "publishedAt":
		return "publishedAt", order == "desc"
	case "minPrice":
		return "minPrice", order == "desc"
	default:
		return "", false
	}
}
