package dto

import "time"

type OrderItemResponse struct {
	SkuID     string `json:"skuId"`
	ProductID string `json:"productId"`
	Name      string `json:"name"`
	Size      string `json:"size"`
	UnitPrice int    `json:"unitPrice"`
	Qty       int    `json:"qty"`
}

type OrderResponse struct {
	ID        int                 `json:"id"`
	UserID    int                 `json:"userId"`
	Status    string              `json:"status"`
	Total     int                 `json:"total"`
	CreatedAt time.Time           `json:"createdAt"`
	Items     []OrderItemResponse `json:"items"`
}
