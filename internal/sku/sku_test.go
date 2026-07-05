package sku

import (
	"context"
	"errors"
	"testing"

	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/pagination"
)

type repositoryFake struct {
	query     Query
	inventory InventoryAdjustment
}

func (f *repositoryFake) List(_ context.Context, query Query) (pagination.CursorPage[SKU], error) {
	f.query = query
	return pagination.CursorPage[SKU]{}, nil
}

func (f *repositoryFake) SetInventory(_ context.Context, adjustment InventoryAdjustment) error {
	f.inventory = adjustment
	return nil
}

func TestListNormalizesDefaults(t *testing.T) {
	repo := &repositoryFake{}
	if _, err := NewService(repo).List(context.Background(), Query{}); err != nil {
		t.Fatal(err)
	}

	if repo.query.Currency != "IDR" || repo.query.Limit != pagination.DefaultLimit {
		t.Fatalf("query was not normalized: %+v", repo.query)
	}
}

func TestSetInventoryValidation(t *testing.T) {
	cases := []InventoryAdjustment{
		{},
		{SKUID: "sku", OnHand: -1},
		{SKUID: "sku", OnHand: 2, Reserved: 3},
	}

	for _, input := range cases {
		if err := NewService(&repositoryFake{}).SetInventory(context.Background(), input); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("expected invalid input for %+v, got %v", input, err)
		}
	}
}

func TestSetInventoryForwardsValidAdjustment(t *testing.T) {
	repo := &repositoryFake{}
	input := InventoryAdjustment{SKUID: "sku", OnHand: 3, Reserved: 1}
	if err := NewService(repo).SetInventory(context.Background(), input); err != nil {
		t.Fatal(err)
	}

	if repo.inventory != input {
		t.Fatalf("adjustment was not forwarded: %+v", repo.inventory)
	}
}
