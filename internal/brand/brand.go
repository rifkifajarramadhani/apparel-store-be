package brand

import "context"

type Brand struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type Repository interface {
	List(context.Context) ([]Brand, error)
}

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(ctx context.Context) ([]Brand, error) { return s.repo.List(ctx) }
