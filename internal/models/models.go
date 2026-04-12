package models

// Product matches the OpenAPI Product schema (id, name, price, category).
type Product struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Category string  `json:"category"`
	Price    float64 `json:"price"`
}

// OrderReq is the POST /order body.
type OrderReq struct {
	CouponCode *string       `json:"couponCode,omitempty"`
	Items      []OrderItemIn `json:"items"`
}

// OrderItemIn is one line in the request.
type OrderItemIn struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

// Order is the successful order response (OpenAPI + helpful totals).
type Order struct {
	ID         string          `json:"id"`
	Items      []OrderItemOut  `json:"items"`
	Products   []Product       `json:"products"`
	CouponCode *string         `json:"couponCode,omitempty"`
	Subtotal   float64         `json:"subtotal"`
	Discount   float64         `json:"discount"`
	Total      float64         `json:"total"`
}

// OrderItemOut matches Order.items in OpenAPI.
type OrderItemOut struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}
