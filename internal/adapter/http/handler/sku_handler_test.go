package handler

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/colourway"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/pagination"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/price"
	appsize "github.com/rifkifajarramadhani/golang-clean-architecture/internal/size"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/sku"
)

type skuServiceFake struct {
	page pagination.CursorPage[sku.SKU]
}

func (f *skuServiceFake) List(context.Context, sku.Query) (pagination.CursorPage[sku.SKU], error) {
	return f.page, nil
}

func (*skuServiceFake) SetInventory(context.Context, sku.InventoryAdjustment) error { return nil }

func TestSKUListReturnsDTOContract(t *testing.T) {
	service := &skuServiceFake{page: pagination.CursorPage[sku.SKU]{Items: []sku.SKU{{
		ID: "sku", Code: "STYLE-BLACK-M", ProductID: "STYLE-1",
		Colourway: colourway.Colourway{ID: "black", Name: "Black", HexCode: "#000000"},
		Size:      appsize.Size{ID: "medium", ScaleCode: "alpha", Code: "M", Name: "Medium", SortOrder: 2},
		Price:     price.Money{Currency: "IDR", Amount: 10000}, OnHand: 8, Reserved: 2, Available: 6,
	}}}}
	app := fiber.New()
	app.Get("/skus", NewSKUHandler(service, slog.New(slog.NewTextHandler(io.Discard, nil))).SKUs)

	response, err := app.Test(httptest.NewRequest("GET", "/skus", nil))
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = response.Body.Close() }()
	var body struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}

	if len(body.Items) != 1 || body.Items[0]["productId"] != "STYLE-1" || body.Items[0]["available"] != float64(6) {
		t.Fatalf("sku response = %#v", body)
	}

	if _, leaked := body.Items[0]["ProductID"]; leaked {
		t.Fatalf("domain field name leaked into response: %#v", body.Items[0])
	}
}
