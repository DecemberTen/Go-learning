package repository

import (
	"context"
	"errors"
	"fmt"

	"example.com/go-learning/internal/model"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var ErrDuplicateSKU = errors.New("商品 SKU 已存在")
var ErrProductUnavailable = errors.New("商品不存在、未上架或库存不足")

// filterProductsByStatus 创建商品状态筛选 Scope。
// 参数：status 为商品状态，空字符串表示不添加筛选条件。
// 返回值：返回可以传给 GORM Scopes 的查询函数。
func filterProductsByStatus(status string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if status != "" {
			return db.Where("status = ?", status)
		}
		return db
	}
}

// paginateProducts 创建商品分页 Scope。
// 参数：page 为页码；pageSize 为每页数量。
// 返回值：返回包含 Limit 和 Offset 的 GORM 查询函数。
func paginateProducts(page int, pageSize int) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Limit(pageSize).Offset((page - 1) * pageSize)
	}
}

// FindProductByID 使用商品 ID 查询商品及其图片。
// 参数：ctx 控制查询取消和超时；id 为商品 ID。
// 返回值：查询成功返回商品；商品不存在或数据库查询失败时返回错误。
func (store *Store) FindProductByID(
	ctx context.Context,
	id int64,
) (model.Product, error) {
	var product model.Product

	err := store.db.WithContext(ctx).
		Preload("Images").
		First(&product, id).
		Error
	if err != nil {
		return model.Product{}, fmt.Errorf("查询商品失败: %w", err)
	}

	return product, nil
}

// ListProducts 根据状态和分页参数查询商品列表。
// 参数：ctx 控制查询取消和超时；status 为状态筛选；page 和 pageSize 控制分页。
// 返回值：查询成功返回商品列表；数据库查询失败时返回错误。
func (store *Store) ListProducts(
	ctx context.Context,
	status string,
	page int,
	pageSize int,
) ([]model.Product, error) {
	products := make([]model.Product, 0)

	err := store.db.WithContext(ctx).
		Scopes(
			filterProductsByStatus(status),
			paginateProducts(page, pageSize),
		).
		Order("id asc").
		Preload("Images").
		Find(&products).
		Error
	if err != nil {
		return nil, fmt.Errorf("查询商品列表失败: %w", err)
	}

	return products, nil
}

// CountProducts 根据状态筛选统计商品总数。
// 参数：ctx 控制查询取消和超时；status 为商品状态筛选。
// 返回值：查询成功返回商品总数；数据库查询失败时返回错误。
func (store *Store) CountProducts(
	ctx context.Context,
	status string,
) (int64, error) {
	var total int64

	err := store.db.WithContext(ctx).
		Model(&model.Product{}).
		Scopes(filterProductsByStatus(status)).
		Count(&total).
		Error
	if err != nil {
		return 0, fmt.Errorf("统计商品总数失败: %w", err)
	}

	return total, nil
}

// CreateProduct 将商品保存到数据库。
// 参数：ctx 控制创建操作取消和超时；product 为待创建商品。
// 返回值：创建成功返回 nil；SKU 重复或数据库操作失败时返回错误。
func (store *Store) CreateProduct(
	ctx context.Context,
	product *model.Product,
) error {
	err := store.db.WithContext(ctx).Create(product).Error
	if err == nil {
		return nil
	}

	var mysqlError *mysql.MySQLError
	if errors.As(err, &mysqlError) && mysqlError.Number == 1062 {
		return ErrDuplicateSKU
	}

	return fmt.Errorf("创建商品失败: %w", err)
}

// UpdateProduct 更新指定商品并返回最新数据。
// 参数：ctx 控制操作取消和超时；id 为商品 ID；其余参数为待更新字段。
// 返回值：更新成功返回商品；商品不存在或数据库操作失败时返回错误。
func (store *Store) UpdateProduct(
	ctx context.Context,
	id int64,
	name string,
	priceCents int64,
	stock int,
	status string,
) (model.Product, error) {
	result := store.db.WithContext(ctx).
		Model(&model.Product{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"name":        name,
			"price_cents": priceCents,
			"stock":       stock,
			"status":      status,
		})

	if result.Error != nil {
		return model.Product{}, fmt.Errorf("更新商品失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return model.Product{}, gorm.ErrRecordNotFound
	}

	return store.FindProductByID(ctx, id)
}

// DeleteProduct 根据商品 ID 软删除商品。
// 参数：ctx 控制操作取消和超时；id 为商品 ID。
// 返回值：删除成功返回 nil；商品不存在或数据库操作失败时返回错误。
func (store *Store) DeleteProduct(ctx context.Context, id int64) error {
	result := store.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&model.Product{})

	if result.Error != nil {
		return fmt.Errorf("删除商品失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// SellProduct 原子扣减商品库存，并在库存清零时将状态改为 sold_out。
// 参数：ctx 控制事务取消和超时；id 为商品 ID；quantity 为销售数量。
// 返回值：销售成功返回 nil；商品不可销售、库存不足或事务失败时返回错误。
func (store *Store) SellProduct(
	ctx context.Context,
	id int64,
	quantity int,
) error {
	if quantity <= 0 {
		return errors.New("销量必须大于 0")
	}

	err := store.Transaction(ctx, func(transactionStore *Store) error {
		result := transactionStore.db.
			Model(&model.Product{}).
			Where(
				"id = ? AND status = 'active' AND stock >= ?",
				id,
				quantity,
			).
			Update("stock", gorm.Expr("stock - ?", quantity))
		if result.Error != nil {
			return fmt.Errorf("更新商品库存失败: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return ErrProductUnavailable
		}

		err := transactionStore.db.
			Model(&model.Product{}).
			Where("id = ? AND stock = 0", id).
			Update("status", "sold_out").
			Error
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

// AppendProductImage 为指定商品添加关联图片。
// 参数：ctx 控制操作取消和超时；productID 为商品 ID；url 为图片地址。
// 返回值：创建成功返回图片；商品不存在或数据库操作失败时返回错误。
func (store *Store) AppendProductImage(
	ctx context.Context,
	productID int64,
	url string,
) (model.ProductImage, error) {
	var product model.Product

	err := store.db.WithContext(ctx).First(&product, productID).Error
	if err != nil {
		return model.ProductImage{}, fmt.Errorf("查询商品失败: %w", err)
	}

	image := model.ProductImage{URL: url}
	err = store.db.WithContext(ctx).
		Model(&product).
		Association("Images").
		Append(&image)
	if err != nil {
		return model.ProductImage{}, fmt.Errorf("添加商品图片失败: %w", err)
	}

	return image, nil
}

// DeleteProductImage 删除指定商品下的一张图片。
// 参数：ctx 控制操作取消和超时；productID 为商品 ID；imageID 为图片 ID。
// 返回值：删除成功返回 nil；图片不存在或数据库操作失败时返回错误。
func (store *Store) DeleteProductImage(
	ctx context.Context,
	productID int64,
	imageID int64,
) error {
	result := store.db.WithContext(ctx).
		Where("id = ? AND product_id = ?", imageID, productID).
		Delete(&model.ProductImage{})

	if result.Error != nil {
		return fmt.Errorf("删除商品图片失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// DisableProducts 批量禁用指定商品。
// 参数：ctx 控制操作取消和超时；ids 为待禁用商品 ID。
// 返回值：返回实际更新行数；数据库操作失败时返回错误。
func (store *Store) DisableProducts(
	ctx context.Context,
	ids []int64,
) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	result := store.db.WithContext(ctx).
		Model(&model.Product{}).
		Where("id IN ?", ids).
		Update("status", "disabled")
	if result.Error != nil {
		return 0, fmt.Errorf("禁用商品失败: %w", result.Error)
	}

	return result.RowsAffected, nil
}
