package handler

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/pagination"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/product"
)

type ProductService interface {
	List(context.Context, product.Query) (pagination.CursorPage[product.Product], error)
	Get(context.Context, string, string) (product.Product, error)
	Create(context.Context, product.Aggregate) error
	Update(context.Context, product.Aggregate) error
	Delete(context.Context, string) error
}

type ProductHandler struct {
	products ProductService
	logger   *slog.Logger
}

func NewProductHandler(products ProductService, logger *slog.Logger) *ProductHandler {
	return &ProductHandler{products: products, logger: logger}
}

func queryLimit(c fiber.Ctx) int { value, _ := strconv.Atoi(c.Query("limit")); return value }

func (h *ProductHandler) Products(c fiber.Ctx) error {
	page, err := h.products.List(c.Context(), product.Query{
		CategorySlug: c.Query("category"), BrandSlug: c.Query("brand"),
		Query: c.Query("q"), Currency: c.Query("currency"),
		Cursor: c.Query("cursor"), Limit: queryLimit(c),
	})
	if err != nil {
		return writeMerchandisingError(c, h.logger, err)
	}

	return c.JSON(toProductPageResponse(page))
}

func (h *ProductHandler) Product(c fiber.Ctx) error {
	item, err := h.products.Get(c.Context(), c.Params("id"), c.Query("currency"))
	if err != nil {
		return writeMerchandisingError(c, h.logger, err)
	}

	return c.JSON(toProductResponse(item))
}

func (h *ProductHandler) Create(c fiber.Ctx) error {
	var input product.Aggregate
	if err := bindJSON(c, &input); err != nil {
		return writeBindError(c, err)
	}

	if err := h.products.Create(c.Context(), input); err != nil {
		return writeMerchandisingError(c, h.logger, err)
	}

	return c.Status(fiber.StatusCreated).JSON(toProductMutationResponse(input))
}

func (h *ProductHandler) Update(c fiber.Ctx) error {
	var input product.Aggregate
	if err := bindJSON(c, &input); err != nil {
		return writeBindError(c, err)
	}

	if input.Product.ID != c.Params("id") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "product id mismatch"})
	}

	if err := h.products.Update(c.Context(), input); err != nil {
		return writeMerchandisingError(c, h.logger, err)
	}

	return c.JSON(toProductMutationResponse(input))
}

func (h *ProductHandler) Delete(c fiber.Ctx) error {
	if err := h.products.Delete(c.Context(), c.Params("id")); err != nil {
		return writeMerchandisingError(c, h.logger, err)
	}

	return c.JSON(fiber.Map{"success": true})
}
