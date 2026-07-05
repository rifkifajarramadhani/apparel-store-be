package category

import "context"

type Category struct {
	ID       string
	ParentID *string
	Slug     string
	Name     string
}

type Repository interface {
	List(context.Context) ([]Category, error)
}

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(ctx context.Context) ([]Category, error) { return s.repo.List(ctx) }
