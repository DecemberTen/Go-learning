package main

type CreateProductRequest struct {
	Name       string `json:"name" binding:"required"`
	PriceCents int64  `json:"price_cents" binding:"gt=0"`
	Stock      int    `json:"stock" binding:"gte=0"`
	Status     string `json:"status" binding:"required,oneof=active sold_out disabled"`
}

type UpdateProductRequest struct {
	Name       string `json:"name" binding:"required"`
	PriceCents int64  `json:"price_cents" binding:"gt=0"`
	Stock      int    `json:"stock" binding:"gte=0"`
	Status     string `json:"status" binding:"required,oneof=active sold_out disabled"`
}

type Product struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	PriceCents int64  `json:"price_cents"`
	Stock      int    `json:"stock"`
	Status     string `json:"status"`
}

type SellProductRequest struct {
	Quantity int `json:"quantity" binding:"required,gt=0"`
}

type ListProductsQuery struct {
	Status   string `form:"status" binding:"omitempty,oneof=active sold_out disabled"`
	Page     int    `form:"page,default=1" binding:"min=1"`
	PageSize int    `form:"page_size,default=20" binding:"min=1,max=100"`
}
