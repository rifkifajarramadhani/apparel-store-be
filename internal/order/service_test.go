package order

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type fakeRepo struct {
	gotLines []Line
	err      error
}

func (f *fakeRepo) Create(_ context.Context, _ int, lines []Line) (Order, error) {
	f.gotLines = lines
	if f.err != nil {
		return Order{}, f.err
	}
	return Order{ID: 1, Items: []Item{}}, nil
}
func (f *fakeRepo) ListByUser(context.Context, int) ([]Order, error)        { return nil, nil }
func (f *fakeRepo) GetByIDForUser(context.Context, int, int) (Order, error) { return Order{}, nil }

func TestCreateRejectsEmpty(t *testing.T) {
	svc := NewService(&fakeRepo{})
	if _, err := svc.Create(context.Background(), 1, nil); !errors.Is(err, ErrEmptyOrder) {
		t.Fatalf("expected ErrEmptyOrder, got %v", err)
	}
}

func TestCreateRejectsNonPositiveQty(t *testing.T) {
	svc := NewService(&fakeRepo{})
	if _, err := svc.Create(context.Background(), 1, []Line{{SkuID: "a", Qty: 0}}); !errors.Is(err, ErrInvalidQty) {
		t.Fatalf("expected ErrInvalidQty, got %v", err)
	}
}

func TestCreateMergesDuplicateSkus(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)
	_, err := svc.Create(context.Background(), 1, []Line{
		{SkuID: "a", Qty: 1}, {SkuID: "b", Qty: 2}, {SkuID: "a", Qty: 3},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []Line{{SkuID: "a", Qty: 4}, {SkuID: "b", Qty: 2}}
	if !reflect.DeepEqual(repo.gotLines, want) {
		t.Errorf("lines = %v, want %v", repo.gotLines, want)
	}
}

func TestCreatePropagatesOutOfStock(t *testing.T) {
	svc := NewService(&fakeRepo{err: ErrOutOfStock})
	if _, err := svc.Create(context.Background(), 1, []Line{{SkuID: "a", Qty: 1}}); !errors.Is(err, ErrOutOfStock) {
		t.Fatalf("expected ErrOutOfStock, got %v", err)
	}
}
