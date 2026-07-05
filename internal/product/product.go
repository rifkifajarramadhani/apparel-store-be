package product

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/asset"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/brand"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/category"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/colourway"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/pagination"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/price"
	appsize "github.com/rifkifajarramadhani/golang-clean-architecture/internal/size"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/sku"
)

var (
	ErrNotFound     = errors.New("product not found")
	ErrInvalidInput = errors.New("invalid product input")
	ErrConflict     = errors.New("product already exists")
	ErrReferenced   = errors.New("product is still referenced")
)

type Product struct {
	ID          string                `json:"id"`
	StyleCode   string                `json:"styleCode"`
	Slug        string                `json:"slug"`
	Name        string                `json:"name"`
	Subtitle    string                `json:"subtitle"`
	Gender      string                `json:"gender,omitempty"`
	ProductType string                `json:"productType,omitempty"`
	Description string                `json:"description,omitempty"`
	Brand       brand.Brand           `json:"brand"`
	Categories  []category.Category   `json:"categories"`
	Colourways  []colourway.Colourway `json:"colourways"`
	Sizes       []appsize.Size        `json:"sizes"`
	Assets      []asset.Asset         `json:"assets"`
	MinPrice    *price.Money          `json:"minPrice,omitempty"`
	MaxPrice    *price.Money          `json:"maxPrice,omitempty"`
}

type Query struct {
	CategorySlug string
	BrandSlug    string
	Query        string
	Currency     string
	Cursor       string
	Limit        int
}

type Aggregate struct {
	Product    Write             `json:"product"`
	Colourways []colourway.Write `json:"colorways"`
	SKUs       []sku.Write       `json:"skus"`
	Images     []asset.Write     `json:"images"`
}

type Write struct {
	ID           string `json:"id"`
	Slug         string `json:"slug"`
	Name         string `json:"name"`
	Subtitle     string `json:"subtitle"`
	Brand        string `json:"brand"`
	Gender       string `json:"gender"`
	ProductType  string `json:"type"`
	Description  string `json:"description"`
	CategorySlug string `json:"categorySlug"`
	SizeScale    string `json:"sizeScale"`
	BasePrice    int64  `json:"basePrice"`
	PublishedAt  string `json:"publishedAt"`
}

type Repository interface {
	List(context.Context, Query) (pagination.CursorPage[Product], error)
	Get(context.Context, string, string) (Product, error)
	Create(context.Context, Aggregate) error
	Update(context.Context, Aggregate) error
	Delete(context.Context, string) error
}

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(ctx context.Context, q Query) (pagination.CursorPage[Product], error) {
	currency, err := normalizeCurrency(q.Currency)
	if err != nil {
		return pagination.CursorPage[Product]{}, err
	}

	q.Currency = currency
	q.Limit = pagination.NormalizeLimit(q.Limit)

	return s.repo.List(ctx, q)
}

func (s *Service) Get(ctx context.Context, id, currency string) (Product, error) {
	currency, err := normalizeCurrency(currency)
	if err != nil {
		return Product{}, err
	}

	return s.repo.Get(ctx, id, currency)
}

func (s *Service) Create(ctx context.Context, in Aggregate) error {
	if err := validateAggregate(in); err != nil {
		return err
	}

	return s.repo.Create(ctx, in)
}

func (s *Service) Update(ctx context.Context, in Aggregate) error {
	if err := validateAggregate(in); err != nil {
		return err
	}

	return s.repo.Update(ctx, in)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%w: product id is required", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
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

func validateAggregate(in Aggregate) error {
	switch {
	case strings.TrimSpace(in.Product.ID) == "":
		return fmt.Errorf("%w: product id is required", ErrInvalidInput)
	case strings.TrimSpace(in.Product.Brand) == "":
		return fmt.Errorf("%w: product brand is required", ErrInvalidInput)
	case len(in.Colourways) == 0:
		return fmt.Errorf("%w: at least one colourway is required", ErrInvalidInput)
	case len(in.SKUs) == 0:
		return fmt.Errorf("%w: at least one sku is required", ErrInvalidInput)
	}

	defaults := 0
	colourways := make(map[string]struct{}, len(in.Colourways))
	for _, item := range in.Colourways {
		if strings.TrimSpace(item.ID) == "" {
			return fmt.Errorf("%w: every colourway needs an id", ErrInvalidInput)
		}

		colourways[item.ID] = struct{}{}
		if item.IsDefault {
			defaults++
		}
	}

	if defaults != 1 {
		return fmt.Errorf("%w: exactly one colourway must be the default", ErrInvalidInput)
	}

	for _, item := range in.SKUs {
		if _, ok := colourways[item.ColourwayID]; !ok {
			return fmt.Errorf("%w: sku %q references an unknown colourway", ErrInvalidInput, item.ID)
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
