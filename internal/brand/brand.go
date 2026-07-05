package brand

import "context"

type Brand struct {
	ID   string
	Slug string
	Name string
}

type Repository interface {
	List(context.Context) ([]Brand, error)
}

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(ctx context.Context) ([]Brand, error) { return s.repo.List(ctx) }
