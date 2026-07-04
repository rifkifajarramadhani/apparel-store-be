package handler

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/catalog"
)

type CatalogService interface {
	ListProducts(context.Context, catalog.ProductQuery) (catalog.CursorPage[catalog.Product], error)
	GetProduct(context.Context, string, string) (catalog.Product, error)
	ListSkus(context.Context, catalog.SkuQuery) (catalog.CursorPage[catalog.Sku], error)
	ListBrands(context.Context) ([]catalog.Brand, error)
	ListCategories(context.Context) ([]catalog.Category, error)
	ListCollections(context.Context) ([]catalog.Collection, error)
	ListColourways(context.Context) ([]catalog.Colourway, error)
	ListSizes(context.Context) ([]catalog.Size, error)
	SetInventory(context.Context, catalog.InventoryAdjustment) error
}

type CatalogHandler struct {
	catalog CatalogService
	logger  *slog.Logger
}

func NewCatalogHandler(service CatalogService, logger *slog.Logger) *CatalogHandler {
	return &CatalogHandler{catalog: service, logger: logger}
}

func queryLimit(c fiber.Ctx) int { value, _ := strconv.Atoi(c.Query("limit")); return value }
func (h *CatalogHandler) Products(c fiber.Ctx) error {
	page, err := h.catalog.ListProducts(c.Context(), catalog.ProductQuery{CategorySlug: c.Query("category"), BrandSlug: c.Query("brand"), Query: c.Query("q"), Currency: c.Query("currency"), Cursor: c.Query("cursor"), Limit: queryLimit(c)})
	if err != nil {
		return writeCatalogError(c, h.logger, err)
	}
	return c.JSON(page)
}
func (h *CatalogHandler) Product(c fiber.Ctx) error {
	item, err := h.catalog.GetProduct(c.Context(), c.Params("id"), c.Query("currency"))
	if err != nil {
		return writeCatalogError(c, h.logger, err)
	}
	return c.JSON(item)
}
func (h *CatalogHandler) Skus(c fiber.Ctx) error {
	page, err := h.catalog.ListSkus(c.Context(), catalog.SkuQuery{ProductID: c.Query("productId"), ColourwayID: c.Query("colourwayId"), Currency: c.Query("currency"), Cursor: c.Query("cursor"), Limit: queryLimit(c)})
	if err != nil {
		return writeCatalogError(c, h.logger, err)
	}
	return c.JSON(page)
}
func (h *CatalogHandler) Brands(c fiber.Ctx) error {
	items, err := h.catalog.ListBrands(c.Context())
	if err != nil {
		return writeCatalogError(c, h.logger, err)
	}
	return c.JSON(items)
}
func (h *CatalogHandler) Categories(c fiber.Ctx) error {
	items, err := h.catalog.ListCategories(c.Context())
	if err != nil {
		return writeCatalogError(c, h.logger, err)
	}
	return c.JSON(items)
}
func (h *CatalogHandler) Collections(c fiber.Ctx) error {
	items, err := h.catalog.ListCollections(c.Context())
	if err != nil {
		return writeCatalogError(c, h.logger, err)
	}
	return c.JSON(items)
}
func (h *CatalogHandler) Colourways(c fiber.Ctx) error {
	items, err := h.catalog.ListColourways(c.Context())
	if err != nil {
		return writeCatalogError(c, h.logger, err)
	}
	return c.JSON(items)
}
func (h *CatalogHandler) Sizes(c fiber.Ctx) error {
	items, err := h.catalog.ListSizes(c.Context())
	if err != nil {
		return writeCatalogError(c, h.logger, err)
	}
	return c.JSON(items)
}
func (h *CatalogHandler) SetInventory(c fiber.Ctx) error {
	var input catalog.InventoryAdjustment
	if err := bindJSON(c, &input); err != nil {
		return writeBindError(c, err)
	}
	if err := h.catalog.SetInventory(c.Context(), input); err != nil {
		return writeCatalogError(c, h.logger, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func writeCatalogError(c fiber.Ctx, logger *slog.Logger, err error) error {
	switch {
	case errors.Is(err, catalog.ErrInvalidInput):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, catalog.ErrNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "catalog resource not found"})
	case errors.Is(err, catalog.ErrConflict), errors.Is(err, catalog.ErrReferenced), errors.Is(err, catalog.ErrInsufficientStock):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	default:
		logger.ErrorContext(c.Context(), "catalog request failed", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
}
