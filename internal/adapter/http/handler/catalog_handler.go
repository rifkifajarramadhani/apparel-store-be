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
	ListProducts(context.Context, catalog.ProductQuery) ([]catalog.Product, int64, error)
	GetProduct(context.Context, string) (catalog.Product, error)
	GetAggregate(context.Context, string) (catalog.Aggregate, error)
	CreateProduct(context.Context, catalog.Aggregate) (catalog.Aggregate, error)
	UpdateProduct(context.Context, string, catalog.Aggregate) (catalog.Aggregate, error)
	DeleteProduct(context.Context, string) error
	ListColorways(context.Context, string, bool) ([]catalog.Colorway, error)
	ListSkus(context.Context, string, string) ([]catalog.Sku, error)
	UpdateSku(context.Context, string, catalog.Sku) (catalog.Sku, error)
	ListCategories(context.Context) ([]catalog.Category, error)
	SaveCategory(context.Context, catalog.Category, bool) error
	DeleteCategory(context.Context, string) error
	ListCollections(context.Context) ([]catalog.Collection, error)
	SaveCollection(context.Context, catalog.Collection, bool) error
	DeleteCollection(context.Context, string) error
	ListSizeScales(context.Context) ([]catalog.SizeScale, error)
}

type CatalogHandler struct {
	catalog CatalogService
	logger  *slog.Logger
}

func NewCatalogHandler(service CatalogService, logger *slog.Logger) *CatalogHandler {
	return &CatalogHandler{catalog: service, logger: logger}
}

func (h *CatalogHandler) GetProducts(c fiber.Ctx) error {
	q := catalog.ProductQuery{
		CategoryID:   c.Query("categoryId"),
		CategorySlug: c.Query("categorySlug"),
		Gender:       c.Query("gender"),
		Slug:         c.Query("slug"),
		Q:            c.Query("q"),
		SortBy:       c.Query("_sort"),
		SortDesc:     c.Query("_order") == "desc",
	}
	if v, ok := optionalInt(c.Query("minPrice_gte")); ok {
		q.MinPrice = &v
	}
	if v, ok := optionalInt(c.Query("maxPrice_lte")); ok {
		q.MaxPrice = &v
	}
	if v, ok := optionalInt(c.Query("_page")); ok {
		q.Page = v
	}
	if v, ok := optionalInt(c.Query("_limit")); ok {
		q.Limit = v
	}
	products, total, err := h.catalog.ListProducts(c.Context(), q)
	if err != nil {
		return writeCatalogError(c, h.logger, err)
	}
	c.Set("X-Total-Count", strconv.FormatInt(total, 10))
	return c.JSON(products)
}

func (h *CatalogHandler) GetProduct(c fiber.Ctx) error {
	product, err := h.catalog.GetProduct(c.Context(), c.Params("id"))
	if err != nil {
		return writeCatalogError(c, h.logger, err)
	}
	return c.JSON(product)
}

func (h *CatalogHandler) GetColorways(c fiber.Ctx) error {
	embed := c.Query("_embed") == "skus"
	colorways, err := h.catalog.ListColorways(c.Context(), c.Query("productId"), embed)
	if err != nil {
		return writeCatalogError(c, h.logger, err)
	}
	return c.JSON(colorways)
}

func (h *CatalogHandler) GetSkus(c fiber.Ctx) error {
	skus, err := h.catalog.ListSkus(c.Context(), c.Query("productId"), c.Query("colorwayId"))
	if err != nil {
		return writeCatalogError(c, h.logger, err)
	}
	return c.JSON(skus)
}

func (h *CatalogHandler) GetCategories(c fiber.Ctx) error {
	categories, err := h.catalog.ListCategories(c.Context())
	if err != nil {
		return writeCatalogError(c, h.logger, err)
	}
	return c.JSON(categories)
}

func (h *CatalogHandler) GetCollections(c fiber.Ctx) error {
	collections, err := h.catalog.ListCollections(c.Context())
	if err != nil {
		return writeCatalogError(c, h.logger, err)
	}
	return c.JSON(collections)
}

func (h *CatalogHandler) GetSizeScales(c fiber.Ctx) error {
	scales, err := h.catalog.ListSizeScales(c.Context())
	if err != nil {
		return writeCatalogError(c, h.logger, err)
	}
	return c.JSON(scales)
}

func (h *CatalogHandler) UpdateSku(c fiber.Ctx) error {
	var sku catalog.Sku
	if err := bindJSON(c, &sku); err != nil {
		return writeBindError(c, err)
	}
	updated, err := h.catalog.UpdateSku(c.Context(), c.Params("id"), sku)
	if err != nil {
		return writeCatalogError(c, h.logger, err)
	}
	return c.JSON(updated)
}

func (h *CatalogHandler) GetProductAggregate(c fiber.Ctx) error {
	agg, err := h.catalog.GetAggregate(c.Context(), c.Params("id"))
	if err != nil {
		return writeCatalogError(c, h.logger, err)
	}
	return c.JSON(agg)
}

func (h *CatalogHandler) CreateProductAggregate(c fiber.Ctx) error {
	var agg catalog.Aggregate
	if err := bindJSON(c, &agg); err != nil {
		return writeBindError(c, err)
	}
	created, err := h.catalog.CreateProduct(c.Context(), agg)
	if err != nil {
		return writeCatalogError(c, h.logger, err)
	}
	return c.Status(fiber.StatusCreated).JSON(created)
}

func (h *CatalogHandler) UpdateProductAggregate(c fiber.Ctx) error {
	var agg catalog.Aggregate
	if err := bindJSON(c, &agg); err != nil {
		return writeBindError(c, err)
	}
	updated, err := h.catalog.UpdateProduct(c.Context(), c.Params("id"), agg)
	if err != nil {
		return writeCatalogError(c, h.logger, err)
	}
	return c.JSON(updated)
}

func (h *CatalogHandler) DeleteProduct(c fiber.Ctx) error {
	if err := h.catalog.DeleteProduct(c.Context(), c.Params("id")); err != nil {
		return writeCatalogError(c, h.logger, err)
	}
	return c.JSON(fiber.Map{"success": true})
}

func (h *CatalogHandler) CreateCategory(c fiber.Ctx) error {
	var category catalog.Category
	if err := bindJSON(c, &category); err != nil {
		return writeBindError(c, err)
	}
	if err := h.catalog.SaveCategory(c.Context(), category, true); err != nil {
		return writeCatalogError(c, h.logger, err)
	}
	return c.Status(fiber.StatusCreated).JSON(category)
}

func (h *CatalogHandler) UpdateCategory(c fiber.Ctx) error {
	var category catalog.Category
	if err := bindJSON(c, &category); err != nil {
		return writeBindError(c, err)
	}
	category.ID = c.Params("id")
	if err := h.catalog.SaveCategory(c.Context(), category, false); err != nil {
		return writeCatalogError(c, h.logger, err)
	}
	return c.JSON(category)
}

func (h *CatalogHandler) DeleteCategory(c fiber.Ctx) error {
	if err := h.catalog.DeleteCategory(c.Context(), c.Params("id")); err != nil {
		return writeCatalogError(c, h.logger, err)
	}
	return c.JSON(fiber.Map{"success": true})
}

func (h *CatalogHandler) CreateCollection(c fiber.Ctx) error {
	var collection catalog.Collection
	if err := bindJSON(c, &collection); err != nil {
		return writeBindError(c, err)
	}
	if err := h.catalog.SaveCollection(c.Context(), collection, true); err != nil {
		return writeCatalogError(c, h.logger, err)
	}
	return c.Status(fiber.StatusCreated).JSON(collection)
}

func (h *CatalogHandler) UpdateCollection(c fiber.Ctx) error {
	var collection catalog.Collection
	if err := bindJSON(c, &collection); err != nil {
		return writeBindError(c, err)
	}
	collection.ID = c.Params("id")
	if err := h.catalog.SaveCollection(c.Context(), collection, false); err != nil {
		return writeCatalogError(c, h.logger, err)
	}
	return c.JSON(collection)
}

func (h *CatalogHandler) DeleteCollection(c fiber.Ctx) error {
	if err := h.catalog.DeleteCollection(c.Context(), c.Params("id")); err != nil {
		return writeCatalogError(c, h.logger, err)
	}
	return c.JSON(fiber.Map{"success": true})
}

func optionalInt(raw string) (int, bool) {
	if raw == "" {
		return 0, false
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	return value, true
}

func writeCatalogError(c fiber.Ctx, logger *slog.Logger, err error) error {
	switch {
	case errors.Is(err, catalog.ErrInvalidInput):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, catalog.ErrNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
	case errors.Is(err, catalog.ErrConflict):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, catalog.ErrReferenced):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	default:
		logger.ErrorContext(c.Context(), "catalog request failed", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
}
