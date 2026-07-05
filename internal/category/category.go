package category

import "context"

type Category struct {
	ID       string  `json:"id"`
	ParentID *string `json:"parentId"`
	Slug     string  `json:"slug"`
	Name     string  `json:"name"`
}

type Repository interface {
	List(context.Context) ([]Category, error)
}

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(ctx context.Context) ([]Category, error) { return s.repo.List(ctx) }
