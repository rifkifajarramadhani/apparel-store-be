package dto

import (
	assetdto "github.com/rifkifajarramadhani/golang-clean-architecture/internal/adapter/http/dto/asset"
	colourwaydto "github.com/rifkifajarramadhani/golang-clean-architecture/internal/adapter/http/dto/colourway"
	pricedto "github.com/rifkifajarramadhani/golang-clean-architecture/internal/adapter/http/dto/price"
	sizedto "github.com/rifkifajarramadhani/golang-clean-architecture/internal/adapter/http/dto/size"
)

type SKUResponse struct {
	ID        string                         `json:"id"`
	Code      string                         `json:"code"`
	Barcode   string                         `json:"barcode,omitempty"`
	ProductID string                         `json:"productId"`
	Colourway colourwaydto.ColourwayResponse `json:"colourway"`
	Size      sizedto.SizeResponse           `json:"size"`
	Price     pricedto.MoneyResponse         `json:"price"`
	OnHand    int                            `json:"onHand"`
	Reserved  int                            `json:"reserved"`
	Available int                            `json:"available"`
	Assets    []assetdto.AssetResponse       `json:"assets"`
}

type SKUWriteResponse struct {
	ID          string `json:"id"`
	ColourwayID string `json:"colorwayId"`
	Size        string `json:"size"`
	SizeScale   string `json:"sizeScale"`
	StockQty    int    `json:"stockQty"`
	Price       int64  `json:"price"`
}
