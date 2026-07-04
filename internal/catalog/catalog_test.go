package catalog

import (
	"context"
	"errors"
	"testing"
	"time"
)

type repositoryFake struct {
	inventory InventoryAdjustment
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
