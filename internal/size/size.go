package size

import "context"

type Size struct {
	ID        string `json:"id"`
	ScaleCode string `json:"scaleCode"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	SortOrder int    `json:"sortOrder"`
}

type Repository interface {
	List(context.Context) ([]Size, error)
}

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(ctx context.Context) ([]Size, error) { return s.repo.List(ctx) }
