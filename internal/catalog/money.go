package catalog

import "time"

type Money struct {
	Currency        string `json:"currency"`
	Amount          int64  `json:"amount"`
	CompareAtAmount *int64 `json:"compareAtAmount,omitempty"`
}

// ActiveAt centralizes half-open price interval semantics: [valid_from, valid_to).
func ActiveAt(from time.Time, to *time.Time, at time.Time) bool {
	return !at.Before(from) && (to == nil || at.Before(*to))
}
