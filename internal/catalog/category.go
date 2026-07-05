package catalog

type Category struct {
	ID       string  `json:"id"`
	ParentID *string `json:"parentId"`
	Slug     string  `json:"slug"`
	Name     string  `json:"name"`
}
