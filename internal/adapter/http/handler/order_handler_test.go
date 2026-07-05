package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/order"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/user"
)

type orderServiceFake struct {
	item order.Order
}

func (f orderServiceFake) Create(context.Context, int, []order.Line) (order.Order, error) {
	return f.item, nil
}

func (f orderServiceFake) ListByUser(context.Context, int) ([]order.Order, error) {
	return []order.Order{f.item}, nil
}

func (f orderServiceFake) GetByIDForUser(context.Context, int, int) (order.Order, error) {
	return f.item, nil
}

func TestOrderHandlersReturnDTOContract(t *testing.T) {
	item := order.Order{
		ID: 7, UserID: 3, Status: order.StatusConfirmed, Total: 20000,
		CreatedAt: time.Date(2026, 7, 5, 12, 30, 0, 0, time.UTC),
		Items:     []order.Item{{SkuID: "sku", ProductID: "STYLE-1", Name: "Shirt", Size: "M", UnitPrice: 10000, Qty: 2}},
	}
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals("auth_user", &user.User{ID: 3})
		return c.Next()
	})
	handler := NewOrderHandler(orderServiceFake{item: item}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	app.Post("/orders", handler.Create)
	app.Get("/orders", handler.List)
	app.Get("/orders/:id", handler.Get)

	tests := []struct {
		name   string
		method string
		path   string
		body   string
		list   bool
	}{
		{name: "create", method: "POST", path: "/orders", body: `{"items":[{"skuId":"sku","qty":2}]}`},
		{name: "list", method: "GET", path: "/orders", list: true},
		{name: "detail", method: "GET", path: "/orders/7"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, bytes.NewBufferString(test.body))
			if test.body != "" {
				request.Header.Set("Content-Type", "application/json")
			}

			response, err := app.Test(request)
			if err != nil {
				t.Fatal(err)
			}

			defer func() { _ = response.Body.Close() }()
			var payload map[string]any
			if test.list {
				var body []map[string]any
				if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}

				if len(body) != 1 {
					t.Fatalf("order list = %#v", body)
				}

				payload = body[0]
			} else if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}

			if payload["userId"] != float64(3) || payload["createdAt"] != "2026-07-05T12:30:00Z" {
				t.Fatalf("order response = %#v", payload)
			}

			items, ok := payload["items"].([]any)
			if !ok || len(items) != 1 || items[0].(map[string]any)["skuId"] != "sku" {
				t.Fatalf("order items = %#v", payload["items"])
			}
		})
	}
}
