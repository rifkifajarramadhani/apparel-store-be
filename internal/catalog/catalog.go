package catalog

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrInsufficientStock = errors.New("insufficient available stock")

const (
	DefaultLimit = 24
	MaxLimit     = 100
)

type Brand struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type Category struct {
	ID       string  `json:"id"`
	ParentID *string `json:"parentId"`
	Slug     string  `json:"slug"`
	Name     string  `json:"name"`
}

type Collection struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type Colourway struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	HexCode string `json:"hexCode"`
}

type Size struct {
	ID        string `json:"id"`
	ScaleCode string `json:"scaleCode"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	SortOrder int    `json:"sortOrder"`
}

type Asset struct {
	ID          string `json:"id"`
	MediaType   string `json:"mediaType"`
	URL         string `json:"url"`
	AltText     string `json:"altText,omitempty"`
	Role        string `json:"role"`
	SortOrder   int    `json:"sortOrder"`
	ColourwayID string `json:"colourwayId,omitempty"`
	SkuID       string `json:"skuId,omitempty"`
}

type Money struct {
	Currency        string `json:"currency"`
	Amount          int64  `json:"amount"`
	CompareAtAmount *int64 `json:"compareAtAmount,omitempty"`
}

type Sku struct {
	ID        string    `json:"id"`
	Code      string    `json:"code"`
	Barcode   string    `json:"barcode,omitempty"`
	ProductID string    `json:"productId"`
	Colourway Colourway `json:"colourway"`
	Size      Size      `json:"size"`
	Price     Money     `json:"price"`
	OnHand    int       `json:"onHand"`
	Reserved  int       `json:"reserved"`
	Available int       `json:"available"`
	Assets    []Asset   `json:"assets"`
}

type Product struct {
	ID          string      `json:"id"`
	StyleCode   string      `json:"styleCode"`
	Slug        string      `json:"slug"`
	Name        string      `json:"name"`
	Subtitle    string      `json:"subtitle"`
	Gender      string      `json:"gender,omitempty"`
	ProductType string      `json:"productType,omitempty"`
	Description string      `json:"description,omitempty"`
	Brand       Brand       `json:"brand"`
	Categories  []Category  `json:"categories"`
	Colourways  []Colourway `json:"colourways"`
	Sizes       []Size      `json:"sizes"`
	Assets      []Asset     `json:"assets"`
	MinPrice    *Money      `json:"minPrice,omitempty"`
	MaxPrice    *Money      `json:"maxPrice,omitempty"`
}

type CursorPage[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"nextCursor,omitempty"`
}

type ProductQuery struct {
	CategorySlug string
	BrandSlug    string
	Query        string
	Currency     string
	Cursor       string
	Limit        int
}

type SkuQuery struct {
	ProductID   string
	ColourwayID string
	Currency    string
	Cursor      string
	Limit       int
}

type InventoryAdjustment struct {
	SkuID    string `json:"skuId"`
	OnHand   int    `json:"onHand"`
	Reserved int    `json:"reserved"`
}

// ProductAggregate is the admin editor's write payload: a product plus its
// colourways and SKUs, referencing dimensions by business key (style code,
// brand name, category/colourway slug, size scale + code).
type ProductAggregate struct {
	Product    ProductWrite
	Colourways []ColourwayWrite
	Skus       []SkuWrite
	Images     []ImageWrite
}

type ProductWrite struct {
	ID           string // style_code
	Slug         string
	Name         string
	Subtitle     string
	Brand        string // brand name
	Gender       string
	ProductType  string
	Description  string
	CategorySlug string
	SizeScale    string
	BasePrice    int64
	PublishedAt  string // YYYY-MM-DD
}

type ColourwayWrite struct {
	ID        string // within-payload correlation key, not persisted
	Name      string
	SwatchHex string
	Price     int64
	IsDefault bool
}

// ImageWrite is one product image; ColourwayID (a business-style colourway
// id), when set, scopes the image to that colourway. Empty means the image
// is shared across all of the product's colourways.
type ImageWrite struct {
	URL         string
	ColourwayID string
}

type SkuWrite struct {
	ID          string // sku_code
	ColourwayID string
	Size        string // size code
	SizeScale   string
	StockQty    int
	Price       int64
}

type Repository interface {
	ListProducts(context.Context, ProductQuery) (CursorPage[Product], error)
	GetProduct(context.Context, string, string) (Product, error)
	ListSkus(context.Context, SkuQuery) (CursorPage[Sku], error)
	ListBrands(context.Context) ([]Brand, error)
	ListCategories(context.Context) ([]Category, error)
	ListCollections(context.Context) ([]Collection, error)
	ListColourways(context.Context) ([]Colourway, error)
	ListSizes(context.Context) ([]Size, error)
	SetInventory(context.Context, InventoryAdjustment) error
	CreateProduct(context.Context, ProductAggregate) error
	UpdateProduct(context.Context, ProductAggregate) error
	DeleteProduct(context.Context, string) error
}

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return DefaultLimit
	}
	if limit > MaxLimit {
		return MaxLimit
	}
	return limit
}

func normalizeCurrency(currency string) (string, error) {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		return "IDR", nil
	}
	if len(currency) != 3 {
		return "", fmt.Errorf("%w: currency must be a three-letter code", ErrInvalidInput)
	}
	return currency, nil
}

func (s *Service) ListProducts(ctx context.Context, q ProductQuery) (CursorPage[Product], error) {
	currency, err := normalizeCurrency(q.Currency)
	if err != nil {
		return CursorPage[Product]{}, err
	}
	q.Currency, q.Limit = currency, normalizeLimit(q.Limit)
	return s.repo.ListProducts(ctx, q)
}

func (s *Service) GetProduct(ctx context.Context, id, currency string) (Product, error) {
	currency, err := normalizeCurrency(currency)
	if err != nil {
		return Product{}, err
	}
	return s.repo.GetProduct(ctx, id, currency)
}

func (s *Service) ListSkus(ctx context.Context, q SkuQuery) (CursorPage[Sku], error) {
	currency, err := normalizeCurrency(q.Currency)
	if err != nil {
		return CursorPage[Sku]{}, err
	}
	q.Currency, q.Limit = currency, normalizeLimit(q.Limit)
	return s.repo.ListSkus(ctx, q)
}

func (s *Service) ListBrands(ctx context.Context) ([]Brand, error) {
	return s.repo.ListBrands(ctx)
}
func (s *Service) ListCategories(ctx context.Context) ([]Category, error) {
	return s.repo.ListCategories(ctx)
}
func (s *Service) ListCollections(ctx context.Context) ([]Collection, error) {
	return s.repo.ListCollections(ctx)
}
func (s *Service) ListColourways(ctx context.Context) ([]Colourway, error) {
	return s.repo.ListColourways(ctx)
}
func (s *Service) ListSizes(ctx context.Context) ([]Size, error) { return s.repo.ListSizes(ctx) }

func (s *Service) SetInventory(ctx context.Context, in InventoryAdjustment) error {
	if in.SkuID == "" || in.OnHand < 0 || in.Reserved < 0 || in.Reserved > in.OnHand {
		return fmt.Errorf("%w: inventory requires a valid sku id and 0 <= reserved <= onHand", ErrInvalidInput)
	}
	return s.repo.SetInventory(ctx, in)
}

func (s *Service) CreateProduct(ctx context.Context, in ProductAggregate) error {
	if err := validateAggregate(in); err != nil {
		return err
	}
	return s.repo.CreateProduct(ctx, in)
}

func (s *Service) UpdateProduct(ctx context.Context, in ProductAggregate) error {
	if err := validateAggregate(in); err != nil {
		return err
	}
	return s.repo.UpdateProduct(ctx, in)
}

func (s *Service) DeleteProduct(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%w: product id is required", ErrInvalidInput)
	}
	return s.repo.DeleteProduct(ctx, id)
}

// validateAggregate enforces the invariants the DB cannot express by itself.
// Field-level shape is already checked by the client; referential integrity
// (unknown brand/category/size) surfaces as ErrInvalidInput from the repo.
func validateAggregate(in ProductAggregate) error {
	switch {
	case strings.TrimSpace(in.Product.ID) == "":
		return fmt.Errorf("%w: product id is required", ErrInvalidInput)
	case strings.TrimSpace(in.Product.Brand) == "":
		return fmt.Errorf("%w: product brand is required", ErrInvalidInput)
	case len(in.Colourways) == 0:
		return fmt.Errorf("%w: at least one colourway is required", ErrInvalidInput)
	case len(in.Skus) == 0:
		return fmt.Errorf("%w: at least one sku is required", ErrInvalidInput)
	}
	defaults := 0
	colourways := make(map[string]struct{}, len(in.Colourways))
	for _, c := range in.Colourways {
		if strings.TrimSpace(c.ID) == "" {
			return fmt.Errorf("%w: every colourway needs an id", ErrInvalidInput)
		}
		colourways[c.ID] = struct{}{}
		if c.IsDefault {
			defaults++
		}
	}
	if defaults != 1 {
		return fmt.Errorf("%w: exactly one colourway must be the default", ErrInvalidInput)
	}
	for _, sku := range in.Skus {
		if _, ok := colourways[sku.ColourwayID]; !ok {
			return fmt.Errorf("%w: sku %q references an unknown colourway", ErrInvalidInput, sku.ID)
		}
	}
	for _, image := range in.Images {
		if image.ColourwayID != "" {
			if _, ok := colourways[image.ColourwayID]; !ok {
				return fmt.Errorf("%w: image %q references an unknown colourway", ErrInvalidInput, image.URL)
			}
		}
	}
	return nil
}

// ActiveAt centralizes half-open price interval semantics: [valid_from, valid_to).
func ActiveAt(from time.Time, to *time.Time, at time.Time) bool {
	return !at.Before(from) && (to == nil || at.Before(*to))
}
