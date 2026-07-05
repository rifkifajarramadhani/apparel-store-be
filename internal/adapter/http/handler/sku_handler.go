package handler

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/pagination"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/sku"
)

type SKUService interface {
	List(context.Context, sku.Query) (pagination.CursorPage[sku.SKU], error)
	SetInventory(context.Context, sku.InventoryAdjustment) error
}

type SKUHandler struct {
	skus   SKUService
	logger *slog.Logger
}

func NewSKUHandler(skus SKUService, logger *slog.Logger) *SKUHandler {
	return &SKUHandler{skus: skus, logger: logger}
}

func (h *SKUHandler) SKUs(c fiber.Ctx) error {
	page, err := h.skus.List(c.Context(), sku.Query{
		ProductID: c.Query("productId"), ColourwayID: c.Query("colourwayId"),
		Currency: c.Query("currency"), Cursor: c.Query("cursor"), Limit: queryLimit(c),
	})
	if err != nil {
		return writeMerchandisingError(c, h.logger, err)
	}

	return c.JSON(toSKUPageResponse(page))
}

func (h *SKUHandler) SetInventory(c fiber.Ctx) error {
	var input sku.InventoryAdjustment
	if err := bindJSON(c, &input); err != nil {
		return writeBindError(c, err)
	}

	input.SKUID = c.Params("id")
	if err := h.skus.SetInventory(c.Context(), input); err != nil {
		return writeMerchandisingError(c, h.logger, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
