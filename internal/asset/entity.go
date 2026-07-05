package asset

type Asset struct {
	ID          string `json:"id"`
	MediaType   string `json:"mediaType"`
	URL         string `json:"url"`
	AltText     string `json:"altText,omitempty"`
	Role        string `json:"role"`
	SortOrder   int    `json:"sortOrder"`
	ColourwayID string `json:"colourwayId,omitempty"`
	SkuID       string `json:"skuId,omitempty"`
}

type Write struct {
	URL         string `json:"url"`
	ColourwayID string `json:"colorwayId,omitempty"`
}
