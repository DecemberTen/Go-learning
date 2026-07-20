package main

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// queryProductByIDWithGorm 使用商品 ID 查询单个商品。
// 参数 ctx：控制数据库查询的取消和超时；db：GORM 数据库对象；id：商品 ID。
// 返回值：查询成功时返回商品；查询失败或商品不存在时返回错误。
func queryProductByIDWithGorm(
	ctx context.Context,
	db *gorm.DB,
	id int64,
) (Product, error) {
	var product Product

	err := db.
		WithContext(ctx).
		First(&product, id).
		Error

	if err != nil {
		return Product{}, fmt.Errorf("查询商品失败: %w", err)
	}

	return product, nil
}

func queryProductsWithGorm(
	ctx context.Context,
	db *gorm.DB,
	filter ListProductsQuery,
) ([]Product, error) {
	products := make([]Product, 0)

	// err := db.WithContext(ctx).Find(&products).Error
	// if err != nil {
	// 	return nil, fmt.Errorf("查询商品失败: %w", err)
	// }

	query := db.WithContext(ctx)
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	offect := (filter.Page - 1) * filter.PageSize

	err := query.
		Order("id asc").
		Limit(filter.PageSize).
		Offset(offect).
		Find(&products).
		Error

	if err != nil {
		return nil, fmt.Errorf("查询商品失败: %w", err)
	}

	return products, nil
}

func createProductWithGorm(
	ctx context.Context,
	db *gorm.DB,
	product *Product,
) error {
	err := db.WithContext(ctx).Create(product).Error
	if err != nil {
		return fmt.Errorf("创建商品失败: %w", err)
	}
	return nil
}

// countProductsWithGorm 根据筛选条件统计商品总数。
// 参数 ctx：控制查询取消和超时；db：GORM 数据库对象；filter：商品筛选条件。
// 返回值：查询成功时返回商品总数；查询失败时返回错误。
func countProductsWithGorm(
	ctx context.Context,
	db *gorm.DB,
	filter ListProductsQuery,
) (int64, error) {
	var total int64
	query := db.WithContext(ctx).Model(&Product{})

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return 0, fmt.Errorf("统计商品总数失败: %w", err)
	}
	return total, nil
}
