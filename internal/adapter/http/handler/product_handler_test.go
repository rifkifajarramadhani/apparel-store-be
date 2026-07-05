package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/pagination"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/product"
)

type productServiceFake struct {
	created product.Aggregate
	page    pagination.CursorPage[product.Product]
	item    product.Product
}

func (f *productServiceFake) List(context.Context, product.Query) (pagination.CursorPage[product.Product], error) {
	return f.page, nil
}

func (f *productServiceFake) Get(context.Context, string, string) (product.Product, error) {
	return f.item, nil
}

func TestProductReadHandlersReturnDTOContract(t *testing.T) {
	item := product.Product{ID: "product", StyleCode: "STYLE-1", Slug: "shirt", Name: "Shirt", Subtitle: "Everyday"}
	service := &productServiceFake{
		page: pagination.CursorPage[product.Product]{Items: []product.Product{item}, NextCursor: "next"},
		item: item,
	}
	app := fiber.New()
	handler := NewProductHandler(service, slog.New(slog.NewTextHandler(io.Discard, nil)))
	app.Get("/products", handler.Products)
	app.Get("/products/:id", handler.Product)

	tests := []struct {
		name string
		path string
		list bool
	}{
		{name: "list", path: "/products", list: true},
		{name: "detail", path: "/products/product"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response, err := app.Test(httptest.NewRequest("GET", test.path, nil))
			if err != nil {
				t.Fatal(err)
			}

			defer func() { _ = response.Body.Close() }()
			var body map[string]any
			if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}

			payload := body
			if test.list {
				items, ok := body["items"].([]any)
				if !ok || len(items) != 1 || body["nextCursor"] != "next" {
					t.Fatalf("list response = %#v", body)
				}

				payload = items[0].(map[string]any)
			}

			if payload["styleCode"] != "STYLE-1" || payload["name"] != "Shirt" {
				t.Fatalf("product response = %#v", payload)
			}

			if _, leaked := payload["StyleCode"]; leaked {
				t.Fatalf("domain field name leaked into response: %#v", payload)
			}
		})
	}
}

func (f *productServiceFake) Create(_ context.Context, aggregate product.Aggregate) error {
	f.created = aggregate
	return nil
}

func (*productServiceFake) Update(context.Context, product.Aggregate) error { return nil }
func (*productServiceFake) Delete(context.Context, string) error            { return nil }

func TestProductCreateUsesCleanAggregateContract(t *testing.T) {
	service := &productServiceFake{}
	app := fiber.New()
	app.Post("/products", NewProductHandler(service, slog.New(slog.NewTextHandler(io.Discard, nil))).Create)

	clean := `{"product":{"id":"STYLE-1","brand":"Acme"},"colorways":[{"id":"BLACK","name":"Black","swatchHex":"#000000","isDefault":true}],"skus":[{"id":"SKU-1","colorwayId":"BLACK","size":"M","sizeScale":"alpha"}],"images":[]}`
	request := httptest.NewRequest("POST", "/products", bytes.NewBufferString(clean))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != fiber.StatusCreated {
		t.Fatalf("status = %d, want %d", response.StatusCode, fiber.StatusCreated)
	}

	if service.created.Product.ID != "STYLE-1" || len(service.created.SKUs) != 1 {
		t.Fatalf("aggregate was not decoded: %+v", service.created)
	}
}

func TestProductCreateRejectsRemovedCompatibilityFields(t *testing.T) {
	app := fiber.New()
	app.Post("/products", NewProductHandler(&productServiceFake{}, slog.New(slog.NewTextHandler(io.Discard, nil))).Create)

	legacy := `{"product":{"id":"STYLE-1","brand":"Acme","categoryId":"legacy"},"colorways":[],"skus":[],"images":[]}`
	request := httptest.NewRequest("POST", "/products", bytes.NewBufferString(legacy))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.StatusCode, fiber.StatusBadRequest)
	}
}
