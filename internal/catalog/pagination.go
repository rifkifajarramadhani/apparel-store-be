package catalog

const (
	DefaultLimit = 24
	MaxLimit     = 100
)

type CursorPage[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"nextCursor,omitempty"`
}
