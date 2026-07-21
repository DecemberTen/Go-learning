package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var ErrDuplicateSKU = errors.New("商品 SKU 已存在")

// filterProductsByStatus 创建商品状态筛选 Scope。
// 参数：status 为商品状态；空字符串表示不添加状态条件。
// 返回值：可以传给 GORM Scopes 的查询函数。
func filterProductsByStatus(status string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if status != "" {
			return db.Where("status = ?", status)
		}
		return db
	}
}

// paginateProducts 创建商品分页 Scope。
// 参数：page 为页码，pageSize 为每页数量。
// 返回值：包含 Limit 和 Offset 的查询函数。
func paginateProducts(
	page int,
	pageSize int,
) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (page - 1) * pageSize

		return db.
			Limit(pageSize).
			Offset(offset)
	}
}

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
		Debug().
		Preload("Images").
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

	// query := db.WithContext(ctx)
	// if filter.Status != "" {
	// 	query = query.Where("status = ?", filter.Status)
	// }
	// offect := (filter.Page - 1) * filter.PageSize

	err := db.WithContext(ctx).
		Scopes(
			filterProductsByStatus(filter.Status),
			paginateProducts(filter.Page, filter.PageSize),
		).
		Order("id asc").
		Preload("Images").
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
	if err == nil {
		return nil
	}

	var mysqlError *mysql.MySQLError
	if errors.As(err, &mysqlError) && mysqlError.Number == 1062 {
		return ErrDuplicateSKU
	}

	return fmt.Errorf("创建商品失败: %w", err)
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
	// query := db.WithContext(ctx).Model(&Product{})

	// if filter.Status != "" {
	// 	query = query.Where("status = ?", filter.Status)
	// }

	err := db.WithContext(ctx).
		Model(&Product{}).
		Scopes(
			filterProductsByStatus(filter.Status),
		).
		Count(&total).
		Error
	if err != nil {
		return 0, fmt.Errorf("统计商品总数失败: %w", err)
	}
	return total, nil
}

func updateProductWithGorm(
	ctx context.Context,
	db *gorm.DB,
	id int64,
	input UpdateProductRequest,
) (Product, error) {
	result := db.WithContext(ctx).Model(&Product{}).Where("id = ?", id).Updates(map[string]any{
		"name":        input.Name,
		"price_cents": input.PriceCents,
		"status":      input.Status,
		"stock":       input.Stock,
	})
	if result.Error != nil {
		return Product{}, fmt.Errorf("更新商品失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return Product{}, gorm.ErrRecordNotFound
	}
	product, err := queryProductByIDWithGorm(ctx, db, id)
	if err != nil {
		return Product{}, err
	}
	return product, nil
}

// deleteProductWithGorm 根据商品 ID 软删除商品。
// 参数 ctx：控制数据库操作取消和超时；db：GORM 数据库对象；id：商品 ID。
// 返回值：删除成功时返回 nil；商品不存在、已删除或数据库执行失败时返回错误。
func deleteProductWithGorm(
	ctx context.Context,
	db *gorm.DB,
	id int64,
) error {
	result := db.WithContext(ctx).Where("id = ?", id).Delete(&Product{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// sellProductWithGorm 原子扣减商品库存，并在库存清零时更新商品状态。
// 参数 ctx：控制事务取消和超时；db：GORM 数据库对象；id：商品 ID；quantity：销售数量。
// 返回值：销售成功时返回 nil；商品不可销售、库存不足或数据库操作失败时返回错误。
func sellProductWithGorm(
	ctx context.Context,
	db *gorm.DB,
	id int64,
	quantity int,
) error {
	if quantity <= 0 {
		return errors.New("销量必须大于0")
	}
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&Product{}).
			Where("id = ? and status = 'active' and stock >= ?", id, quantity).
			Update("stock", gorm.Expr("stock - ?", quantity))

		if result.Error != nil {
			return fmt.Errorf("更新商品库存失败: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return ErrProductUnavailable
		}

		err := tx.Model(&Product{}).
			Where("id = ? and stock = 0", id).
			Update("status", "sold_out").Error
		if err != nil {
			return fmt.Errorf("更新商品状态失败: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("销售商品失败: %w", err)
	}
	return nil
}

// appendProductImageWithGorm 为指定商品添加关联图片。
// 参数：ctx 控制取消和超时，db 为 GORM 对象，productID 为商品 ID，url 为图片地址。
// 返回值：创建成功时返回图片；商品不存在或数据库操作失败时返回错误。
func appendProductImageWithGorm(
	ctx context.Context,
	db *gorm.DB,
	productID int64,
	url string,
) (ProductImage, error) {
	var product Product

	err := db.WithContext(ctx).
		First(&product, productID).
		Error

	if err != nil {
		return ProductImage{}, fmt.Errorf("查询商品失败: %w", err)
	}

	image := ProductImage{
		URL: url,
	}

	err = db.WithContext(ctx).
		Model(&product).
		Association("Images").
		Append(&image)

	if err != nil {
		return ProductImage{}, fmt.Errorf("添加商品图片失败: %w", err)
	}
	return image, nil
}

// deleteProductImageWithGorm 删除指定商品下的一张图片。
// 参数：ctx 控制取消和超时，db 为 GORM 对象，productID 为商品 ID，imageID 为图片 ID。
// 返回值：删除成功返回 nil；图片不存在或数据库操作失败时返回错误。
func deleteProductImageWithGorm(
	ctx context.Context,
	db *gorm.DB,
	productID int64,
	imageID int64,
) error {
	result := db.WithContext(ctx).
		Where("id = ? and product_id = ?", imageID, productID).
		Delete(&ProductImage{})

	if result.Error != nil {
		return fmt.Errorf("删除图片失败 %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func disableProductsWithGorm(
	ctx context.Context,
	db *gorm.DB,
	ids []int64,
) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	result := db.WithContext(ctx).
		Model(&Product{}).Where("id in ?", ids).
		Update("status", "disabled")
	if result.Error != nil {
		return 0, fmt.Errorf("禁用商品失败: %w", result.Error)
	}
	return result.RowsAffected, nil
}
