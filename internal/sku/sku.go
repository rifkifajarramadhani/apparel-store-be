package sku

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/asset"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/colourway"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/pagination"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/price"
	appsize "github.com/rifkifajarramadhani/golang-clean-architecture/internal/size"
)

var (
	ErrNotFound     = errors.New("sku not found")
	ErrInvalidInput = errors.New("invalid sku input")
)

type SKU struct {
	ID        string
	Code      string
	Barcode   string
	ProductID string
	Colourway colourway.Colourway
	Size      appsize.Size
	Price     price.Money
	OnHand    int
	Reserved  int
	Available int
	Assets    []asset.Asset
}

type Query struct {
	ProductID   string
	ColourwayID string
	Currency    string
	Cursor      string
	Limit       int
}

type InventoryAdjustment struct {
	SKUID    string `json:"-"`
	OnHand   int    `json:"onHand"`
	Reserved int    `json:"reserved"`
}

type Write struct {
	ID          string `json:"id"`
	ColourwayID string `json:"colorwayId"`
	Size        string `json:"size"`
	SizeScale   string `json:"sizeScale"`
	StockQty    int    `json:"stockQty"`
	Price       int64  `json:"price"`
}

type Repository interface {
	List(context.Context, Query) (pagination.CursorPage[SKU], error)
	SetInventory(context.Context, InventoryAdjustment) error
}

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(ctx context.Context, q Query) (pagination.CursorPage[SKU], error) {
	currency, err := normalizeCurrency(q.Currency)
	if err != nil {
		return pagination.CursorPage[SKU]{}, err
	}

	q.Currency = currency
	q.Limit = pagination.NormalizeLimit(q.Limit)

	return s.repo.List(ctx, q)
}

func (s *Service) SetInventory(ctx context.Context, in InventoryAdjustment) error {
	if strings.TrimSpace(in.SKUID) == "" || in.OnHand < 0 || in.Reserved < 0 || in.Reserved > in.OnHand {
		return fmt.Errorf("%w: inventory requires a valid sku id and 0 <= reserved <= onHand", ErrInvalidInput)
	}

	return s.repo.SetInventory(ctx, in)
}

func normalizeCurrency(currency string) (string, error) {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		return "IDR", nil
	}

	if len(currency) != 3 {
		return "", fmt.Errorf("%w: currency must be a three-letter code", ErrInvalidInput)
	}

	return currency, nil
}
