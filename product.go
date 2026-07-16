package main

import (
	"errors"
	"time"
)

type Product struct {
	ID         int64         `json:"id"`
	Name       string        `json:"name"`
	PriceCents int64         `json:"price_cents"`
	Stock      int           `json:"stock"`
	Status     ProductStatus `json:"status"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

type ProductStatus string

const (
	ProductStatusActive   ProductStatus = "active"
	ProductStatusSoldOut  ProductStatus = "sold_out"
	ProductStatusDisabled ProductStatus = "disabled"
)

// IsValid 用于判断商品状态是否受系统支持；参数为空；返回状态是否合法。
func (status ProductStatus) IsValid() bool {
	switch status {
	case ProductStatusActive, ProductStatusSoldOut, ProductStatusDisabled:
		return true
	default:
		return false
	}
}

type CreateProductRequest struct {
	Name       string        `json:"name"`
	PriceCents int64         `json:"price_cents"`
	Stock      int           `json:"stock"`
	Status     ProductStatus `json:"status"`
}

type UpdateProductRequest struct {
	Name       string        `json:"name"`
	PriceCents int64         `json:"price_cents"`
	Stock      int           `json:"stock"`
	Status     ProductStatus `json:"status"`
}

// validateProductInput 用于校验商品请求字段；参数依次为名称、价格、库存和状态；返回 nil 表示合法，否则返回具体错误。
func validateProductInput(name string, priceCents int64, stock int, status ProductStatus) error {
	if name == "" {
		return errors.New("Name is required")
	}
	if priceCents <= 0 {
		return errors.New("Price must be greater than 0")
	}
	if stock < 0 {
		return errors.New("Stock cannot be negative")
	}
	if !status.IsValid() {
		return errors.New("Invalid product status")
	}

	return nil
}
