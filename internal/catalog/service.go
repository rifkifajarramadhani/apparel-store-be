package catalog

import (
	"context"
	"fmt"
	"strings"
)

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
