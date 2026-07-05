package dto

type SizeResponse struct {
	ID        string `json:"id"`
	ScaleCode string `json:"scaleCode"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	SortOrder int    `json:"sortOrder"`
}
