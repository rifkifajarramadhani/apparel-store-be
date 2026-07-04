package catalog

import (
	"context"
	"errors"
	"testing"
	"time"
)

type repositoryFake struct {
	inventory InventoryAdjustment
	saved     *ProductAggregate
	deleted   string
	err       error
}

func (f *repositoryFake) ListProducts(context.Context, ProductQuery) (CursorPage[Product], error) {
	return CursorPage[Product]{}, f.err
}
func (f *repositoryFake) GetProduct(context.Context, string, string) (Product, error) {
	return Product{}, f.err
}
func (f *repositoryFake) ListSkus(context.Context, SkuQuery) (CursorPage[Sku], error) {
	return CursorPage[Sku]{}, f.err
}
func (f *repositoryFake) ListBrands(context.Context) ([]Brand, error)           { return nil, f.err }
func (f *repositoryFake) ListCategories(context.Context) ([]Category, error)    { return nil, f.err }
func (f *repositoryFake) ListCollections(context.Context) ([]Collection, error) { return nil, f.err }
func (f *repositoryFake) ListColourways(context.Context) ([]Colourway, error) {
	return nil, f.err
}
func (f *repositoryFake) ListSizes(context.Context) ([]Size, error) { return nil, f.err }
func (f *repositoryFake) SetInventory(_ context.Context, in InventoryAdjustment) error {
	f.inventory = in
	return f.err
}
func (f *repositoryFake) CreateProduct(_ context.Context, in ProductAggregate) error {
	f.saved = &in
	return f.err
}
func (f *repositoryFake) UpdateProduct(_ context.Context, in ProductAggregate) error {
	f.saved = &in
	return f.err
}
func (f *repositoryFake) DeleteProduct(_ context.Context, id string) error {
	f.deleted = id
	return f.err
}

func validAggregate() ProductAggregate {
	return ProductAggregate{
		Product:    ProductWrite{ID: "STYLE-1", Brand: "Acme"},
		Colourways: []ColourwayWrite{{ID: "BLACK", IsDefault: true}},
		Skus:       []SkuWrite{{ID: "SKU-1", ColourwayID: "BLACK"}},
	}
}

func TestCreateProductValidation(t *testing.T) {
	cases := map[string]func(*ProductAggregate){
		"missing product id": func(a *ProductAggregate) { a.Product.ID = "" },
		"missing brand":      func(a *ProductAggregate) { a.Product.Brand = "" },
		"no colourways":      func(a *ProductAggregate) { a.Colourways = nil },
		"no skus":            func(a *ProductAggregate) { a.Skus = nil },
		"no default":         func(a *ProductAggregate) { a.Colourways[0].IsDefault = false },
		"two defaults": func(a *ProductAggregate) {
			a.Colourways = append(a.Colourways, ColourwayWrite{ID: "RED", IsDefault: true})
		},
		"sku unknown colourway": func(a *ProductAggregate) { a.Skus[0].ColourwayID = "GHOST" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			agg := validAggregate()
			mutate(&agg)
			svc := NewService(&repositoryFake{})
			if err := svc.CreateProduct(context.Background(), agg); !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("expected invalid input, got %v", err)
			}
		})
	}
}

func TestCreateProductPassesValidAggregateToRepo(t *testing.T) {
	repo := &repositoryFake{}
	svc := NewService(repo)
	if err := svc.CreateProduct(context.Background(), validAggregate()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.saved == nil || repo.saved.Product.ID != "STYLE-1" {
		t.Fatalf("aggregate not forwarded to repo: %+v", repo.saved)
	}
}

func TestDeleteProductRequiresID(t *testing.T) {
	svc := NewService(&repositoryFake{})
	if err := svc.DeleteProduct(context.Background(), "  "); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func TestSetInventoryRejectsOverReservation(t *testing.T) {
	svc := NewService(&repositoryFake{})
	err := svc.SetInventory(context.Background(), InventoryAdjustment{SkuID: "sku", OnHand: 2, Reserved: 3})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func TestActiveAtUsesHalfOpenInterval(t *testing.T) {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(time.Hour)
	if !ActiveAt(from, &to, from) {
		t.Fatal("valid_from should be inclusive")
	}
	if ActiveAt(from, &to, to) {
		t.Fatal("valid_to should be exclusive")
	}
}
