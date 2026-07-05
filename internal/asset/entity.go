package asset

type Asset struct {
	ID          string
	MediaType   string
	URL         string
	AltText     string
	Role        string
	SortOrder   int
	ColourwayID string
	SkuID       string
}

type Write struct {
	URL         string `json:"url"`
	ColourwayID string `json:"colorwayId,omitempty"`
}
