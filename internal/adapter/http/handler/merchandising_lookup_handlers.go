package handler

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/brand"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/category"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/collection"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/colourway"
	appsize "github.com/rifkifajarramadhani/golang-clean-architecture/internal/size"
)

type BrandService interface {
	List(context.Context) ([]brand.Brand, error)
}
type CategoryService interface {
	List(context.Context) ([]category.Category, error)
}
type CollectionService interface {
	List(context.Context) ([]collection.Collection, error)
}
type ColourwayService interface {
	List(context.Context) ([]colourway.Colourway, error)
}
type SizeService interface {
	List(context.Context) ([]appsize.Size, error)
}

type BrandHandler struct {
	service BrandService
	logger  *slog.Logger
}

func NewBrandHandler(service BrandService, logger *slog.Logger) *BrandHandler {
	return &BrandHandler{service: service, logger: logger}
}

func (h *BrandHandler) List(c fiber.Ctx) error {
	items, err := h.service.List(c.Context())
	if err != nil {
		return writeMerchandisingError(c, h.logger, err)
	}

	return c.JSON(items)
}

type CategoryHandler struct {
	service CategoryService
	logger  *slog.Logger
}

func NewCategoryHandler(service CategoryService, logger *slog.Logger) *CategoryHandler {
	return &CategoryHandler{service: service, logger: logger}
}

func (h *CategoryHandler) List(c fiber.Ctx) error {
	items, err := h.service.List(c.Context())
	if err != nil {
		return writeMerchandisingError(c, h.logger, err)
	}

	return c.JSON(items)
}

type CollectionHandler struct {
	service CollectionService
	logger  *slog.Logger
}

func NewCollectionHandler(service CollectionService, logger *slog.Logger) *CollectionHandler {
	return &CollectionHandler{service: service, logger: logger}
}

func (h *CollectionHandler) List(c fiber.Ctx) error {
	items, err := h.service.List(c.Context())
	if err != nil {
		return writeMerchandisingError(c, h.logger, err)
	}

	return c.JSON(items)
}

type ColourwayHandler struct {
	service ColourwayService
	logger  *slog.Logger
}

func NewColourwayHandler(service ColourwayService, logger *slog.Logger) *ColourwayHandler {
	return &ColourwayHandler{service: service, logger: logger}
}

func (h *ColourwayHandler) List(c fiber.Ctx) error {
	items, err := h.service.List(c.Context())
	if err != nil {
		return writeMerchandisingError(c, h.logger, err)
	}

	return c.JSON(items)
}

type SizeHandler struct {
	service SizeService
	logger  *slog.Logger
}

func NewSizeHandler(service SizeService, logger *slog.Logger) *SizeHandler {
	return &SizeHandler{service: service, logger: logger}
}

func (h *SizeHandler) List(c fiber.Ctx) error {
	items, err := h.service.List(c.Context())
	if err != nil {
		return writeMerchandisingError(c, h.logger, err)
	}

	return c.JSON(items)
}
