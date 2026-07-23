package model

import (
	"time"

	"gorm.io/gorm"
)

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

type ProductImage struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	ProductID int64     `gorm:"not null;index" json:"product_id"`
	URL       string    `gorm:"size:500;not null" json:"url"`
	CreatedAt time.Time `json:"created_at"`
}
