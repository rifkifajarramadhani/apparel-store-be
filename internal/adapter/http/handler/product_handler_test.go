package handler

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/pagination"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/product"
)

type productServiceFake struct{ created product.Aggregate }

func (*productServiceFake) List(context.Context, product.Query) (pagination.CursorPage[product.Product], error) {
	return pagination.CursorPage[product.Product]{}, nil
}

func (*productServiceFake) Get(context.Context, string, string) (product.Product, error) {
	return product.Product{}, nil
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
