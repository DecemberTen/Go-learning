package handler

import (
	"time"

	"example.com/go-learning/internal/model"
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

type SellProductRequest struct {
	Quantity int `json:"quantity" binding:"required,gt=0"`
}

type ListProductsQuery struct {
	Status   string `form:"status" binding:"omitempty,oneof=active sold_out disabled"`
	Page     int    `form:"page,default=1" binding:"min=1"`
	PageSize int    `form:"page_size,default=20" binding:"min=1,max=100"`
}

type ProductListResponse struct {
	Items      []model.Product `json:"items"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	Total      int64           `json:"total"`
	TotalPages int64           `json:"total_pages"`
}

type CreateProductImageRequest struct {
	URL string `json:"url" binding:"required,url"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type UserResponse struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
	User         UserResponse `json:"user"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}
