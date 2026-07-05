package product

import (
	"context"
	"errors"
	"testing"

	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/colourway"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/pagination"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/sku"
)

type repositoryFake struct {
	query   Query
	saved   *Aggregate
	deleted string
	err     error
}

func (f *repositoryFake) List(_ context.Context, query Query) (pagination.CursorPage[Product], error) {
	f.query = query
	return pagination.CursorPage[Product]{}, f.err
}

func (f *repositoryFake) Get(context.Context, string, string) (Product, error) {
	return Product{}, f.err
}

func (f *repositoryFake) Create(_ context.Context, aggregate Aggregate) error {
	f.saved = &aggregate
	return f.err
}

func (f *repositoryFake) Update(_ context.Context, aggregate Aggregate) error {
	f.saved = &aggregate
	return f.err
}

func (f *repositoryFake) Delete(_ context.Context, id string) error {
	f.deleted = id
	return f.err
}

func validAggregate() Aggregate {
	return Aggregate{
		Product:    Write{ID: "STYLE-1", Brand: "Acme"},
		Colourways: []colourway.Write{{ID: "BLACK", IsDefault: true}},
		SKUs:       []sku.Write{{ID: "SKU-1", ColourwayID: "BLACK"}},
	}
}

func TestCreateValidatesAggregate(t *testing.T) {
	cases := map[string]func(*Aggregate){
		"missing product id": func(a *Aggregate) { a.Product.ID = "" },
		"missing brand":      func(a *Aggregate) { a.Product.Brand = "" },
		"no colourways":      func(a *Aggregate) { a.Colourways = nil },
		"no skus":            func(a *Aggregate) { a.SKUs = nil },
		"no default":         func(a *Aggregate) { a.Colourways[0].IsDefault = false },
		"two defaults": func(a *Aggregate) {
			a.Colourways = append(a.Colourways, colourway.Write{ID: "RED", IsDefault: true})
		},
		"unknown sku colourway": func(a *Aggregate) { a.SKUs[0].ColourwayID = "GHOST" },
	}

	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			aggregate := validAggregate()
			mutate(&aggregate)
			if err := NewService(&repositoryFake{}).Create(context.Background(), aggregate); !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("expected invalid input, got %v", err)
			}
		})
	}
}

func TestListNormalizesCurrencyAndLimit(t *testing.T) {
	repo := &repositoryFake{}
	if _, err := NewService(repo).List(context.Background(), Query{Currency: "idr", Limit: 1000}); err != nil {
		t.Fatal(err)
	}

	if repo.query.Currency != "IDR" || repo.query.Limit != pagination.MaxLimit {
		t.Fatalf("query was not normalized: %+v", repo.query)
	}
}

func TestDeleteRequiresID(t *testing.T) {
	if err := NewService(&repositoryFake{}).Delete(context.Background(), "  "); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}
