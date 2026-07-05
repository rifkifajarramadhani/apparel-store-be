package price

import "time"

type Money struct {
	Currency        string `json:"currency"`
	Amount          int64  `json:"amount"`
	CompareAtAmount *int64 `json:"compareAtAmount,omitempty"`
}

// ActiveAt reports whether a price applies at the supplied instant using a
// half-open interval: [validFrom, validTo).
func ActiveAt(validFrom time.Time, validTo *time.Time, at time.Time) bool {
	return !at.Before(validFrom) && (validTo == nil || at.Before(*validTo))
}
