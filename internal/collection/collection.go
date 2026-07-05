package collection

import "context"

type Collection struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type Repository interface {
	List(context.Context) ([]Collection, error)
}

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(ctx context.Context) ([]Collection, error) { return s.repo.List(ctx) }
