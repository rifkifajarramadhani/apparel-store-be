package order

import "context"

// Repository persists orders. Create must run atomically: resolving each SKU,
// checking and decrementing stock, and inserting the order + items in one tx.
type Repository interface {
	Create(ctx context.Context, userID int, lines []Line) (Order, error)
	ListByUser(ctx context.Context, userID int) ([]Order, error)
	GetByIDForUser(ctx context.Context, userID, orderID int) (Order, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, userID int, lines []Line) (Order, error) {
	if len(lines) == 0 {
		return Order{}, ErrEmptyOrder
	}

	// Collapse duplicate SKUs and validate quantities.
	merged := make(map[string]int)
	order := make([]string, 0, len(lines))
	for _, line := range lines {
		if line.Qty <= 0 {
			return Order{}, ErrInvalidQty
		}
		if _, seen := merged[line.SkuID]; !seen {
			order = append(order, line.SkuID)
		}
		merged[line.SkuID] += line.Qty
	}

	deduped := make([]Line, 0, len(order))
	for _, skuID := range order {
		deduped = append(deduped, Line{SkuID: skuID, Qty: merged[skuID]})
	}

	return s.repo.Create(ctx, userID, deduped)
}

func (s *Service) ListByUser(ctx context.Context, userID int) ([]Order, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) GetByIDForUser(ctx context.Context, userID, orderID int) (Order, error) {
	return s.repo.GetByIDForUser(ctx, userID, orderID)
}
