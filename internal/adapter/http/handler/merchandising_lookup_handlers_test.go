package handler

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/brand"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/category"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/collection"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/colourway"
	appsize "github.com/rifkifajarramadhani/golang-clean-architecture/internal/size"
)

type brandServiceFake struct{ items []brand.Brand }

func (f brandServiceFake) List(context.Context) ([]brand.Brand, error) { return f.items, nil }

type categoryServiceFake struct{ items []category.Category }

func (f categoryServiceFake) List(context.Context) ([]category.Category, error) { return f.items, nil }

type collectionServiceFake struct{ items []collection.Collection }

func (f collectionServiceFake) List(context.Context) ([]collection.Collection, error) {
	return f.items, nil
}

type colourwayServiceFake struct{ items []colourway.Colourway }

func (f colourwayServiceFake) List(context.Context) ([]colourway.Colourway, error) {
	return f.items, nil
}

type sizeServiceFake struct{ items []appsize.Size }

func (f sizeServiceFake) List(context.Context) ([]appsize.Size, error) { return f.items, nil }

func TestLookupHandlersReturnDTOContracts(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tests := []struct {
		name      string
		handler   fiber.Handler
		field     string
		wantValue any
	}{
		{name: "brands", handler: NewBrandHandler(brandServiceFake{[]brand.Brand{{ID: "brand", Slug: "axis", Name: "Axis"}}}, logger).List, field: "slug", wantValue: "axis"},
		{name: "categories", handler: NewCategoryHandler(categoryServiceFake{[]category.Category{{ID: "category", Slug: "tops", Name: "Tops"}}}, logger).List, field: "name", wantValue: "Tops"},
		{name: "collections", handler: NewCollectionHandler(collectionServiceFake{[]collection.Collection{{ID: "collection", Slug: "new", Name: "New"}}}, logger).List, field: "slug", wantValue: "new"},
		{name: "colourways", handler: NewColourwayHandler(colourwayServiceFake{[]colourway.Colourway{{ID: "black", Name: "Black", HexCode: "#000000"}}}, logger).List, field: "hexCode", wantValue: "#000000"},
		{name: "sizes", handler: NewSizeHandler(sizeServiceFake{[]appsize.Size{{ID: "medium", ScaleCode: "alpha", Code: "M", Name: "Medium", SortOrder: 2}}}, logger).List, field: "sortOrder", wantValue: float64(2)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/lookup", test.handler)
			response, err := app.Test(httptest.NewRequest("GET", "/lookup", nil))
			if err != nil {
				t.Fatal(err)
			}

			defer func() { _ = response.Body.Close() }()
			var body []map[string]any
			if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}

			if len(body) != 1 || body[0][test.field] != test.wantValue {
				t.Fatalf("lookup response = %#v", body)
			}
		})
	}
}
