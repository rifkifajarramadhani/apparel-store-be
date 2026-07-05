package size

import "context"

type Size struct {
	ID        string
	ScaleCode string
	Code      string
	Name      string
	SortOrder int
}

type Repository interface {
	List(context.Context) ([]Size, error)
}

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(ctx context.Context) ([]Size, error) { return s.repo.List(ctx) }
