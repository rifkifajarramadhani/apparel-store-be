package catalog

import (
	"errors"
	"reflect"
	"testing"
)

func baseAggregate() Aggregate {
	return Aggregate{
		Product: Product{ID: "AX-1", Slug: "ax-1", Name: "Tee", Badges: nil},
		Colorways: []Colorway{
			{ID: "AX1-010", ProductID: "AX-1", StyleColor: "AX1-010", ColorFamily: "Black",
				SwatchHex: "#111", Price: 1500, IsDefault: true, OnSale: false,
				Images: []string{"a.jpg", "b.jpg"}},
			{ID: "AX1-100", ProductID: "AX-1", StyleColor: "AX1-100", ColorFamily: "White",
				SwatchHex: "#fff", Price: 1200, IsDefault: false, OnSale: true,
				Images: []string{"c.jpg"}},
		},
		Skus: []Sku{
			{ID: "AX1-010-M", ProductID: "AX-1", ColorwayID: "AX1-010", Size: "M"},
			{ID: "AX1-010-L", ProductID: "AX-1", ColorwayID: "AX1-010", Size: "L"},
			{ID: "AX1-100-M", ProductID: "AX-1", ColorwayID: "AX1-100", Size: "M"},
		},
	}
}

func TestDerive(t *testing.T) {
	got, err := Derive(baseAggregate(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.MinPrice != 1200 || got.MaxPrice != 1500 {
		t.Errorf("price range = %d..%d, want 1200..1500", got.MinPrice, got.MaxPrice)
	}
	if got.ColorwayCount != 2 {
		t.Errorf("colorwayCount = %d, want 2", got.ColorwayCount)
	}
	if got.DefaultColorwayID != "AX1-010" {
		t.Errorf("defaultColorwayId = %q, want AX1-010", got.DefaultColorwayID)
	}
	if got.ThumbnailURL != "a.jpg" || got.HoverImageURL != "b.jpg" {
		t.Errorf("images = %q/%q, want a.jpg/b.jpg", got.ThumbnailURL, got.HoverImageURL)
	}
	if !reflect.DeepEqual(got.Badges, []string{"Sale"}) {
		t.Errorf("badges = %v, want [Sale] (a colorway is on sale)", got.Badges)
	}
	if !reflect.DeepEqual(got.Sizes, []string{"M", "L"}) {
		t.Errorf("sizes = %v, want [M L] (deduped, in order)", got.Sizes)
	}
	if !reflect.DeepEqual(got.ColorFamilies, []string{"Black", "White"}) {
		t.Errorf("colorFamilies = %v, want [Black White]", got.ColorFamilies)
	}
}

func TestDeriveRetainsNonSaleBadges(t *testing.T) {
	agg := baseAggregate()
	agg.Colorways[1].OnSale = false // nothing on sale now
	existing := &Product{Badges: []string{"New", "Sale"}}
	got, err := Derive(agg, existing)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got.Badges, []string{"New"}) {
		t.Errorf("badges = %v, want [New] (Sale dropped, New retained)", got.Badges)
	}
}

func TestDeriveValidation(t *testing.T) {
	tests := map[string]func(a *Aggregate){
		"no default":       func(a *Aggregate) { a.Colorways[0].IsDefault = false },
		"two defaults":     func(a *Aggregate) { a.Colorways[1].IsDefault = true },
		"foreign colorway": func(a *Aggregate) { a.Colorways[0].ProductID = "OTHER" },
		"foreign sku":      func(a *Aggregate) { a.Skus[0].ColorwayID = "GHOST" },
		"no colorways":     func(a *Aggregate) { a.Colorways = nil },
		"no skus":          func(a *Aggregate) { a.Skus = nil },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			agg := baseAggregate()
			mutate(&agg)
			if _, err := Derive(agg, nil); !errors.Is(err, ErrInvalidInput) {
				t.Errorf("expected ErrInvalidInput, got %v", err)
			}
		})
	}
}
