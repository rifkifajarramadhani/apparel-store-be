package pagination

const (
	DefaultLimit = 24
	MaxLimit     = 100
)

type CursorPage[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"nextCursor,omitempty"`
}

func NormalizeLimit(limit int) int {
	if limit <= 0 {
		return DefaultLimit
	}

	if limit > MaxLimit {
		return MaxLimit
	}

	return limit
}
