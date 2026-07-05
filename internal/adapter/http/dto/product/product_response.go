package dto

import (
	assetdto "github.com/rifkifajarramadhani/golang-clean-architecture/internal/adapter/http/dto/asset"
	branddto "github.com/rifkifajarramadhani/golang-clean-architecture/internal/adapter/http/dto/brand"
	categorydto "github.com/rifkifajarramadhani/golang-clean-architecture/internal/adapter/http/dto/category"
	colourwaydto "github.com/rifkifajarramadhani/golang-clean-architecture/internal/adapter/http/dto/colourway"
	pricedto "github.com/rifkifajarramadhani/golang-clean-architecture/internal/adapter/http/dto/price"
	sizedto "github.com/rifkifajarramadhani/golang-clean-architecture/internal/adapter/http/dto/size"
	skudto "github.com/rifkifajarramadhani/golang-clean-architecture/internal/adapter/http/dto/sku"
)

type ProductResponse struct {
	ID          string                           `json:"id"`
	StyleCode   string                           `json:"styleCode"`
	Slug        string                           `json:"slug"`
	Name        string                           `json:"name"`
	Subtitle    string                           `json:"subtitle"`
	Gender      string                           `json:"gender,omitempty"`
	ProductType string                           `json:"productType,omitempty"`
	Description string                           `json:"description,omitempty"`
	Brand       branddto.BrandResponse           `json:"brand"`
	Categories  []categorydto.CategoryResponse   `json:"categories"`
	Colourways  []colourwaydto.ColourwayResponse `json:"colourways"`
	Sizes       []sizedto.SizeResponse           `json:"sizes"`
	Assets      []assetdto.AssetResponse         `json:"assets"`
	MinPrice    *pricedto.MoneyResponse          `json:"minPrice,omitempty"`
	MaxPrice    *pricedto.MoneyResponse          `json:"maxPrice,omitempty"`
}

type ProductWriteResponse struct {
	ID           string `json:"id"`
	Slug         string `json:"slug"`
	Name         string `json:"name"`
	Subtitle     string `json:"subtitle"`
	Brand        string `json:"brand"`
	Gender       string `json:"gender"`
	ProductType  string `json:"type"`
	Description  string `json:"description"`
	CategorySlug string `json:"categorySlug"`
	SizeScale    string `json:"sizeScale"`
	BasePrice    int64  `json:"basePrice"`
	PublishedAt  string `json:"publishedAt"`
}

type ColourwayWriteResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	SwatchHex string `json:"swatchHex"`
	Price     int64  `json:"price"`
	IsDefault bool   `json:"isDefault"`
}

type ImageWriteResponse struct {
	URL         string `json:"url"`
	ColourwayID string `json:"colorwayId,omitempty"`
}

type ProductMutationResponse struct {
	Product    ProductWriteResponse      `json:"product"`
	Colourways []ColourwayWriteResponse  `json:"colorways"`
	SKUs       []skudto.SKUWriteResponse `json:"skus"`
	Images     []ImageWriteResponse      `json:"images"`
}
