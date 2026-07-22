package main

import (
	"time"

	"gorm.io/gorm"
)

type CreateProductRequest struct {
	Name       string `json:"name" binding:"required"`
	SKU        string `json:"sku" binding:"required"`
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
	ID int64 `json:"id"`

	Name string `gorm:"size:255;not null" json:"name"`

	SKU string `gorm:"size:64;uniqueIndex;not null" json:"sku"`

	PriceCents int64 `gorm:"not null;check:chk_products_price_positive,price_cents > 0" json:"price_cents"`

	Stock int `gorm:"not null;check:chk_products_stock_nonnegative,stock >= 0" json:"stock"`

	Status string `gorm:"size:32;not null;index;check:chk_products_status,status IN ('active','sold_out','disabled')" json:"status"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Images []ProductImage `gorm:"foreignKey:ProductID" json:"images,omitempty"`
}

type SellProductRequest struct {
	Quantity int `json:"quantity" binding:"required,gt=0"`
}

type ListProductsQuery struct {
	Status   string `form:"status" binding:"omitempty,oneof=active sold_out disabled"`
	Page     int    `form:"page,default=1" binding:"min=1"`
	PageSize int    `form:"page_size,default=20" binding:"min=1,max=100"`
}

type ProductListResponse struct {
	Items      []Product `json:"items"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	Total      int64     `json:"total"`
	TotalPages int64     `json:"total_pages"`
}

type ProductImage struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	ProductID int64     `gorm:"not null;index" json:"product_id"`
	URL       string    `gorm:"size:500;not null" json:"url"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateProductImageRequest struct {
	URL string `json:"url" binding:"required,url"`
}


