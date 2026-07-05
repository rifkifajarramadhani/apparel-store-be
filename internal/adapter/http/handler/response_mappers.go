package handler

import (
	assetdto "github.com/rifkifajarramadhani/golang-clean-architecture/internal/adapter/http/dto/asset"
	branddto "github.com/rifkifajarramadhani/golang-clean-architecture/internal/adapter/http/dto/brand"
	categorydto "github.com/rifkifajarramadhani/golang-clean-architecture/internal/adapter/http/dto/category"
	collectiondto "github.com/rifkifajarramadhani/golang-clean-architecture/internal/adapter/http/dto/collection"
	colourwaydto "github.com/rifkifajarramadhani/golang-clean-architecture/internal/adapter/http/dto/colourway"
	orderdto "github.com/rifkifajarramadhani/golang-clean-architecture/internal/adapter/http/dto/order"
	paginationdto "github.com/rifkifajarramadhani/golang-clean-architecture/internal/adapter/http/dto/pagination"
	pricedto "github.com/rifkifajarramadhani/golang-clean-architecture/internal/adapter/http/dto/price"
	productdto "github.com/rifkifajarramadhani/golang-clean-architecture/internal/adapter/http/dto/product"
	sizedto "github.com/rifkifajarramadhani/golang-clean-architecture/internal/adapter/http/dto/size"
	skudto "github.com/rifkifajarramadhani/golang-clean-architecture/internal/adapter/http/dto/sku"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/asset"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/brand"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/category"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/collection"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/colourway"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/order"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/pagination"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/price"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/product"
	appsize "github.com/rifkifajarramadhani/golang-clean-architecture/internal/size"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/sku"
)

func mapResponses[S any, T any](items []S, mapper func(S) T) []T {
	if items == nil {
		return nil
	}

	responses := make([]T, len(items))
	for i, item := range items {
		responses[i] = mapper(item)
	}

	return responses
}

func toAssetResponse(item asset.Asset) assetdto.AssetResponse {
	return assetdto.AssetResponse{
		ID: item.ID, MediaType: item.MediaType, URL: item.URL, AltText: item.AltText,
		Role: item.Role, SortOrder: item.SortOrder, ColourwayID: item.ColourwayID, SkuID: item.SkuID,
	}
}

func toBrandResponse(item brand.Brand) branddto.BrandResponse {
	return branddto.BrandResponse{ID: item.ID, Slug: item.Slug, Name: item.Name}
}

func toCategoryResponse(item category.Category) categorydto.CategoryResponse {
	return categorydto.CategoryResponse{ID: item.ID, ParentID: item.ParentID, Slug: item.Slug, Name: item.Name}
}

func toCollectionResponse(item collection.Collection) collectiondto.CollectionResponse {
	return collectiondto.CollectionResponse{ID: item.ID, Slug: item.Slug, Name: item.Name}
}

func toColourwayResponse(item colourway.Colourway) colourwaydto.ColourwayResponse {
	return colourwaydto.ColourwayResponse{ID: item.ID, Name: item.Name, HexCode: item.HexCode}
}

func toSizeResponse(item appsize.Size) sizedto.SizeResponse {
	return sizedto.SizeResponse{
		ID: item.ID, ScaleCode: item.ScaleCode, Code: item.Code, Name: item.Name, SortOrder: item.SortOrder,
	}
}

func toMoneyResponse(item price.Money) pricedto.MoneyResponse {
	return pricedto.MoneyResponse{
		Currency: item.Currency, Amount: item.Amount, CompareAtAmount: item.CompareAtAmount,
	}
}

func toMoneyResponsePointer(item *price.Money) *pricedto.MoneyResponse {
	if item == nil {
		return nil
	}

	response := toMoneyResponse(*item)
	return &response
}

func toProductResponse(item product.Product) productdto.ProductResponse {
	return productdto.ProductResponse{
		ID: item.ID, StyleCode: item.StyleCode, Slug: item.Slug, Name: item.Name,
		Subtitle: item.Subtitle, Gender: item.Gender, ProductType: item.ProductType,
		Description: item.Description, Brand: toBrandResponse(item.Brand),
		Categories: mapResponses(item.Categories, toCategoryResponse),
		Colourways: mapResponses(item.Colourways, toColourwayResponse),
		Sizes:      mapResponses(item.Sizes, toSizeResponse),
		Assets:     mapResponses(item.Assets, toAssetResponse),
		MinPrice:   toMoneyResponsePointer(item.MinPrice), MaxPrice: toMoneyResponsePointer(item.MaxPrice),
	}
}

func toProductPageResponse(page pagination.CursorPage[product.Product]) paginationdto.CursorPageResponse[productdto.ProductResponse] {
	return paginationdto.CursorPageResponse[productdto.ProductResponse]{
		Items: mapResponses(page.Items, toProductResponse), NextCursor: page.NextCursor,
	}
}

func toProductMutationResponse(item product.Aggregate) productdto.ProductMutationResponse {
	return productdto.ProductMutationResponse{
		Product: productdto.ProductWriteResponse{
			ID: item.Product.ID, Slug: item.Product.Slug, Name: item.Product.Name,
			Subtitle: item.Product.Subtitle, Brand: item.Product.Brand, Gender: item.Product.Gender,
			ProductType: item.Product.ProductType, Description: item.Product.Description,
			CategorySlug: item.Product.CategorySlug, SizeScale: item.Product.SizeScale,
			BasePrice: item.Product.BasePrice, PublishedAt: item.Product.PublishedAt,
		},
		Colourways: mapResponses(item.Colourways, func(value colourway.Write) productdto.ColourwayWriteResponse {
			return productdto.ColourwayWriteResponse{
				ID: value.ID, Name: value.Name, SwatchHex: value.SwatchHex,
				Price: value.Price, IsDefault: value.IsDefault,
			}
		}),
		SKUs: mapResponses(item.SKUs, func(value sku.Write) skudto.SKUWriteResponse {
			return skudto.SKUWriteResponse{
				ID: value.ID, ColourwayID: value.ColourwayID, Size: value.Size,
				SizeScale: value.SizeScale, StockQty: value.StockQty, Price: value.Price,
			}
		}),
		Images: mapResponses(item.Images, func(value asset.Write) productdto.ImageWriteResponse {
			return productdto.ImageWriteResponse{URL: value.URL, ColourwayID: value.ColourwayID}
		}),
	}
}

func toSKUResponse(item sku.SKU) skudto.SKUResponse {
	return skudto.SKUResponse{
		ID: item.ID, Code: item.Code, Barcode: item.Barcode, ProductID: item.ProductID,
		Colourway: toColourwayResponse(item.Colourway), Size: toSizeResponse(item.Size),
		Price: toMoneyResponse(item.Price), OnHand: item.OnHand, Reserved: item.Reserved,
		Available: item.Available, Assets: mapResponses(item.Assets, toAssetResponse),
	}
}

func toSKUPageResponse(page pagination.CursorPage[sku.SKU]) paginationdto.CursorPageResponse[skudto.SKUResponse] {
	return paginationdto.CursorPageResponse[skudto.SKUResponse]{
		Items: mapResponses(page.Items, toSKUResponse), NextCursor: page.NextCursor,
	}
}

func toOrderResponse(item order.Order) orderdto.OrderResponse {
	return orderdto.OrderResponse{
		ID: item.ID, UserID: item.UserID, Status: item.Status, Total: item.Total, CreatedAt: item.CreatedAt,
		Items: mapResponses(item.Items, func(value order.Item) orderdto.OrderItemResponse {
			return orderdto.OrderItemResponse{
				SkuID: value.SkuID, ProductID: value.ProductID, Name: value.Name,
				Size: value.Size, UnitPrice: value.UnitPrice, Qty: value.Qty,
			}
		}),
	}
}

func toOrderResponses(items []order.Order) []orderdto.OrderResponse {
	return mapResponses(items, toOrderResponse)
}
