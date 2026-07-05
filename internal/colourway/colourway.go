package colourway

import "context"

type Colourway struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	HexCode string `json:"hexCode"`
}

type Write struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	SwatchHex string `json:"swatchHex"`
	Price     int64  `json:"price"`
	IsDefault bool   `json:"isDefault"`
}

type Repository interface {
	List(context.Context) ([]Colourway, error)
}

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(ctx context.Context) ([]Colourway, error) { return s.repo.List(ctx) }
