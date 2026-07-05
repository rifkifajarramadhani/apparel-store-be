package handler

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/product"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/sku"
)

func writeMerchandisingError(c fiber.Ctx, logger *slog.Logger, err error) error {
	switch {
	case errors.Is(err, product.ErrInvalidInput), errors.Is(err, sku.ErrInvalidInput):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, product.ErrNotFound), errors.Is(err, sku.ErrNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "resource not found"})
	case errors.Is(err, product.ErrConflict), errors.Is(err, product.ErrReferenced):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	default:
		logger.ErrorContext(c.Context(), "domain request failed", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
}
