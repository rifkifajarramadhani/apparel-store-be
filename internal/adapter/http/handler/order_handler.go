package handler

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/order"
)

type OrderService interface {
	Create(context.Context, int, []order.Line) (order.Order, error)
	ListByUser(context.Context, int) ([]order.Order, error)
	GetByIDForUser(context.Context, int, int) (order.Order, error)
}

type OrderHandler struct {
	orders OrderService
	logger *slog.Logger
}

func NewOrderHandler(service OrderService, logger *slog.Logger) *OrderHandler {
	return &OrderHandler{orders: service, logger: logger}
}

type createOrderRequest struct {
	Items []struct {
		SkuID string `json:"skuId"`
		Qty   int    `json:"qty"`
	} `json:"items"`
}

func (h *OrderHandler) Create(c fiber.Ctx) error {
	account, ok := currentUser(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req createOrderRequest
	if err := bindJSON(c, &req); err != nil {
		return writeBindError(c, err)
	}
	lines := make([]order.Line, 0, len(req.Items))
	for _, item := range req.Items {
		lines = append(lines, order.Line{SkuID: item.SkuID, Qty: item.Qty})
	}
	created, err := h.orders.Create(c.Context(), account.ID, lines)
	if err != nil {
		return writeOrderError(c, h.logger, err)
	}
	return c.Status(fiber.StatusCreated).JSON(created)
}

func (h *OrderHandler) List(c fiber.Ctx) error {
	account, ok := currentUser(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	orders, err := h.orders.ListByUser(c.Context(), account.ID)
	if err != nil {
		return writeOrderError(c, h.logger, err)
	}
	return c.JSON(orders)
}

func (h *OrderHandler) Get(c fiber.Ctx) error {
	account, ok := currentUser(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid order id"})
	}
	found, err := h.orders.GetByIDForUser(c.Context(), account.ID, id)
	if err != nil {
		return writeOrderError(c, h.logger, err)
	}
	return c.JSON(found)
}

func writeOrderError(c fiber.Ctx, logger *slog.Logger, err error) error {
	switch {
	case errors.Is(err, order.ErrEmptyOrder), errors.Is(err, order.ErrInvalidQty):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, order.ErrOutOfStock):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
	case errors.Is(err, order.ErrNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "order not found"})
	default:
		logger.ErrorContext(c.Context(), "order request failed", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}
}
