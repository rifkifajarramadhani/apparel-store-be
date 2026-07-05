package handler

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/asset"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/brand"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/category"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/collection"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/colourway"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/order"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/pagination"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/price"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/product"
	appsize "github.com/rifkifajarramadhani/golang-clean-architecture/internal/size"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/sku"
)

func TestLookupResponseMappers(t *testing.T) {
	parentID := "parent"
	tests := []struct {
		name string
		got  any
		want any
	}{
		{name: "brand", got: toBrandResponse(brand.Brand{ID: "brand", Slug: "axis", Name: "Axis"}), want: map[string]any{"id": "brand", "slug": "axis", "name": "Axis"}},
		{name: "category", got: toCategoryResponse(category.Category{ID: "category", ParentID: &parentID, Slug: "tops", Name: "Tops"}), want: map[string]any{"id": "category", "parentId": "parent", "slug": "tops", "name": "Tops"}},
		{name: "collection", got: toCollectionResponse(collection.Collection{ID: "collection", Slug: "new", Name: "New"}), want: map[string]any{"id": "collection", "slug": "new", "name": "New"}},
		{name: "colourway", got: toColourwayResponse(colourway.Colourway{ID: "colour", Name: "Black", HexCode: "#000000"}), want: map[string]any{"id": "colour", "name": "Black", "hexCode": "#000000"}},
		{name: "size", got: toSizeResponse(appsize.Size{ID: "size", ScaleCode: "alpha", Code: "M", Name: "Medium", SortOrder: 2}), want: map[string]any{"id": "size", "scaleCode": "alpha", "code": "M", "name": "Medium", "sortOrder": float64(2)}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			encoded, err := json.Marshal(test.got)
			if err != nil {
				t.Fatal(err)
			}

			var got map[string]any
			if err := json.Unmarshal(encoded, &got); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("response = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestProductResponseMapperPreservesNestedContract(t *testing.T) {
	compareAt := int64(12000)
	minimum := price.Money{Currency: "IDR", Amount: 10000, CompareAtAmount: &compareAt}
	item := product.Product{
		ID: "product", StyleCode: "STYLE-1", Slug: "shirt", Name: "Shirt", Subtitle: "Everyday",
		Gender: "men", ProductType: "apparel", Description: "Cotton",
		Brand:      brand.Brand{ID: "brand", Slug: "axis", Name: "Axis"},
		Categories: []category.Category{{ID: "category", Slug: "tops", Name: "Tops"}},
		Colourways: []colourway.Colourway{{ID: "colour", Name: "Black", HexCode: "#000000"}},
		Sizes:      []appsize.Size{{ID: "size", ScaleCode: "alpha", Code: "M", Name: "Medium", SortOrder: 2}},
		Assets:     []asset.Asset{{ID: "asset", MediaType: "image", URL: "https://example.com/a.png", Role: "primary", SortOrder: 1, ColourwayID: "colour"}},
		MinPrice:   &minimum,
	}

	response := toProductPageResponse(pagination.CursorPage[product.Product]{Items: []product.Product{item}, NextCursor: "next"})
	if len(response.Items) != 1 || response.NextCursor != "next" {
		t.Fatalf("page mapping failed: %+v", response)
	}

	got := response.Items[0]
	if got.Brand.Name != "Axis" || got.Categories[0].Slug != "tops" || got.Colourways[0].HexCode != "#000000" {
		t.Fatalf("nested product mapping failed: %+v", got)
	}

	if got.MinPrice == nil || got.MinPrice.CompareAtAmount == nil || *got.MinPrice.CompareAtAmount != compareAt || got.MaxPrice != nil {
		t.Fatalf("price mapping failed: %+v", got)
	}
}

func TestProductMutationResponseIsWireCompatible(t *testing.T) {
	aggregate := product.Aggregate{
		Product: product.Write{
			ID: "STYLE-1", Slug: "shirt", Name: "Shirt", Subtitle: "Everyday", Brand: "Axis",
			Gender: "men", ProductType: "apparel", Description: "Cotton", CategorySlug: "tops",
			SizeScale: "alpha", BasePrice: 10000, PublishedAt: "2026-07-05",
		},
		Colourways: []colourway.Write{{ID: "black", Name: "Black", SwatchHex: "#000000", Price: 10000, IsDefault: true}},
		SKUs:       []sku.Write{{ID: "sku", ColourwayID: "black", Size: "M", SizeScale: "alpha", StockQty: 5, Price: 10000}},
		Images:     []asset.Write{{URL: "https://example.com/a.png", ColourwayID: "black"}},
	}

	want, err := json.Marshal(aggregate)
	if err != nil {
		t.Fatal(err)
	}

	got, err := json.Marshal(toProductMutationResponse(aggregate))
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != string(want) {
		t.Fatalf("mutation response changed\ngot:  %s\nwant: %s", got, want)
	}
}

func TestSKUAndOrderResponseMappers(t *testing.T) {
	createdAt := time.Date(2026, 7, 5, 12, 30, 0, 0, time.UTC)
	skuResponse := toSKUPageResponse(pagination.CursorPage[sku.SKU]{
		Items: []sku.SKU{{
			ID: "sku", Code: "STYLE-BLACK-M", Barcode: "123", ProductID: "STYLE-1",
			Colourway: colourway.Colourway{ID: "black", Name: "Black", HexCode: "#000000"},
			Size:      appsize.Size{ID: "medium", ScaleCode: "alpha", Code: "M", Name: "Medium", SortOrder: 2},
			Price:     price.Money{Currency: "IDR", Amount: 10000}, OnHand: 8, Reserved: 2, Available: 6,
			Assets: []asset.Asset{},
		}},
	})
	if len(skuResponse.Items) != 1 || skuResponse.Items[0].Price.Amount != 10000 || skuResponse.Items[0].Available != 6 {
		t.Fatalf("sku mapping failed: %+v", skuResponse)
	}

	orderResponse := toOrderResponse(order.Order{
		ID: 7, UserID: 3, Status: order.StatusConfirmed, Total: 20000, CreatedAt: createdAt,
		Items: []order.Item{{SkuID: "sku", ProductID: "STYLE-1", Name: "Shirt", Size: "M", UnitPrice: 10000, Qty: 2}},
	})
	if orderResponse.CreatedAt != createdAt || len(orderResponse.Items) != 1 || orderResponse.Items[0].SkuID != "sku" {
		t.Fatalf("order mapping failed: %+v", orderResponse)
	}
}

func TestMapResponsesPreservesNilAndEmptySlices(t *testing.T) {
	mapper := func(value int) int { return value }
	if mapResponses[int, int](nil, mapper) != nil {
		t.Fatal("nil input must remain nil")
	}

	empty := mapResponses([]int{}, mapper)
	if empty == nil || len(empty) != 0 {
		t.Fatalf("empty input must remain a non-nil empty slice: %#v", empty)
	}
}
