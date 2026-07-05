package price

import "time"

type Money struct {
	Currency        string
	Amount          int64
	CompareAtAmount *int64
}

// ActiveAt reports whether a price applies at the supplied instant using a
// half-open interval: [validFrom, validTo).
func ActiveAt(validFrom time.Time, validTo *time.Time, at time.Time) bool {
	return !at.Before(validFrom) && (validTo == nil || at.Before(*validTo))
}
