package catalog

type Product struct {
	ID          string      `json:"id"`
	StyleCode   string      `json:"styleCode"`
	Slug        string      `json:"slug"`
	Name        string      `json:"name"`
	Subtitle    string      `json:"subtitle"`
	Gender      string      `json:"gender,omitempty"`
	ProductType string      `json:"productType,omitempty"`
	Description string      `json:"description,omitempty"`
	Brand       Brand       `json:"brand"`
	Categories  []Category  `json:"categories"`
	Colourways  []Colourway `json:"colourways"`
	Sizes       []Size      `json:"sizes"`
	Assets      []Asset     `json:"assets"`
	MinPrice    *Money      `json:"minPrice,omitempty"`
	MaxPrice    *Money      `json:"maxPrice,omitempty"`
}

type ProductQuery struct {
	CategorySlug string
	BrandSlug    string
	Query        string
	Currency     string
	Cursor       string
	Limit        int
}
