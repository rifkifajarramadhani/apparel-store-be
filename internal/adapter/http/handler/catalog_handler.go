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
	CreateProduct(context.Context, catalog.ProductAggregate) error
	UpdateProduct(context.Context, catalog.ProductAggregate) error
	DeleteProduct(context.Context, string) error
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

// productAggregateDTO mirrors the admin editor's ProductAggregateInput JSON.
// bindJSON rejects unknown fields, so every field the client sends is listed.
type productAggregateDTO struct {
	Product struct {
		ID            string   `json:"id"`
		Slug          string   `json:"slug"`
		Name          string   `json:"name"`
		Subtitle      string   `json:"subtitle"`
		Brand         string   `json:"brand"`
		Gender        string   `json:"gender"`
		Type          string   `json:"type"`
		CategoryID    string   `json:"categoryId"`
		CategorySlug  string   `json:"categorySlug"`
		CollectionIDs []string `json:"collectionIds"`
		SizeScale     string   `json:"sizeScale"`
		BasePrice     int64    `json:"basePrice"`
		Description   string   `json:"description"`
		PublishedAt   string   `json:"publishedAt"`
	} `json:"product"`
	Colorways []struct {
		ID         string `json:"id"`
		ProductID  string `json:"productId"`
		StyleColor string `json:"styleColor"`
		Name       string `json:"name"`
		SwatchHex  string `json:"swatchHex"`
		Price      int64  `json:"price"`
		IsDefault  bool   `json:"isDefault"`
		OnSale     bool   `json:"onSale"`
	} `json:"colorways"`
	Skus []struct {
		ID         string `json:"id"`
		ColorwayID string `json:"colorwayId"`
		ProductID  string `json:"productId"`
		Size       string `json:"size"`
		SizeLabel  string `json:"sizeLabel"`
		SizeScale  string `json:"sizeScale"`
		InStock    bool   `json:"inStock"`
		StockQty   int    `json:"stockQty"`
		Price      int64  `json:"price"`
	} `json:"skus"`
	Images []struct {
		URL        string `json:"url"`
		ColorwayID string `json:"colorwayId"`
	} `json:"images"`
}

func (d productAggregateDTO) toDomain() catalog.ProductAggregate {
	agg := catalog.ProductAggregate{
		Product: catalog.ProductWrite{
			ID: d.Product.ID, Slug: d.Product.Slug, Name: d.Product.Name,
			Subtitle: d.Product.Subtitle, Brand: d.Product.Brand, Gender: d.Product.Gender,
			ProductType: d.Product.Type, Description: d.Product.Description,
			CategorySlug: d.Product.CategorySlug, SizeScale: d.Product.SizeScale,
			BasePrice: d.Product.BasePrice, PublishedAt: d.Product.PublishedAt,
		},
		Colourways: make([]catalog.ColourwayWrite, 0, len(d.Colorways)),
		Skus:       make([]catalog.SkuWrite, 0, len(d.Skus)),
		Images:     make([]catalog.ImageWrite, 0, len(d.Images)),
	}

	for _, c := range d.Colorways {
		agg.Colourways = append(agg.Colourways, catalog.ColourwayWrite{
			ID: c.ID, Name: c.Name, SwatchHex: c.SwatchHex,
			Price: c.Price, IsDefault: c.IsDefault,
		})
	}

	for _, s := range d.Skus {
		agg.Skus = append(agg.Skus, catalog.SkuWrite{
			ID: s.ID, ColourwayID: s.ColorwayID, Size: s.Size, SizeScale: s.SizeScale,
			StockQty: s.StockQty, Price: s.Price,
		})
	}

	for _, img := range d.Images {
		agg.Images = append(agg.Images, catalog.ImageWrite{URL: img.URL, ColourwayID: img.ColorwayID})
	}

	return agg
}

func (h *CatalogHandler) CreateProduct(c fiber.Ctx) error {
	var dto productAggregateDTO
	if err := bindJSON(c, &dto); err != nil {
		return writeBindError(c, err)
	}

	if err := h.catalog.CreateProduct(c.Context(), dto.toDomain()); err != nil {
		return writeCatalogError(c, h.logger, err)
	}

	return c.Status(fiber.StatusCreated).JSON(dto)
}

func (h *CatalogHandler) UpdateProduct(c fiber.Ctx) error {
	var dto productAggregateDTO
	if err := bindJSON(c, &dto); err != nil {
		return writeBindError(c, err)
	}

	if dto.Product.ID != c.Params("id") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "product id mismatch"})
	}

	if err := h.catalog.UpdateProduct(c.Context(), dto.toDomain()); err != nil {
		return writeCatalogError(c, h.logger, err)
	}

	return c.JSON(dto)
}

func (h *CatalogHandler) DeleteProduct(c fiber.Ctx) error {
	if err := h.catalog.DeleteProduct(c.Context(), c.Params("id")); err != nil {
		return writeCatalogError(c, h.logger, err)
	}

	return c.JSON(fiber.Map{"success": true})
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
