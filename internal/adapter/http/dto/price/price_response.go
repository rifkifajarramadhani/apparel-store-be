package dto

type MoneyResponse struct {
	Currency        string `json:"currency"`
	Amount          int64  `json:"amount"`
	CompareAtAmount *int64 `json:"compareAtAmount,omitempty"`
}
