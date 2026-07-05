package price

import (
	"testing"
	"time"
)

func TestActiveAtUsesHalfOpenInterval(t *testing.T) {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(time.Hour)
	if !ActiveAt(from, &to, from) {
		t.Fatal("valid_from should be inclusive")
	}

	if ActiveAt(from, &to, to) {
		t.Fatal("valid_to should be exclusive")
	}
}
