package service

import (
	"context"

	"example.com/go-learning/internal/model"
	"example.com/go-learning/internal/repository"
)

var ErrDuplicateSKU = repository.ErrDuplicateSKU
var ErrProductUnavailable = repository.ErrProductUnavailable
var ErrNotFound = repository.ErrNotFound

type ProductService struct {
	store *repository.Store
}

type CreateProductInput struct {
	Name       string
	SKU        string
	PriceCents int64
	Stock      int
	Status     string
}

type UpdateProductInput struct {
	Name       string
	PriceCents int64
	Stock      int
	Status     string
}

type ProductFilter struct {
	Status   string
	Page     int
	PageSize int
}

type ProductList struct {
	Items      []model.Product
	Page       int
	PageSize   int
	Total      int64
	TotalPages int64
}

// NewProductService 创建商品业务 Service。
// 参数：store 为商品数据访问 Store。
// 返回值：返回可处理商品业务的 ProductService。
func NewProductService(store *repository.Store) *ProductService {
	return &ProductService{store: store}
}

// GetProduct 根据 ID 查询商品详情。
// 参数：ctx 控制查询取消和超时；id 为商品 ID。
// 返回值：查询成功返回商品；商品不存在或查询失败时返回错误。
func (service *ProductService) GetProduct(
	ctx context.Context,
	id int64,
) (model.Product, error) {
	return service.store.FindProductByID(ctx, id)
}

// CreateProduct 根据输入创建商品。
// 参数：ctx 控制操作取消和超时；input 为已经通过 HTTP 校验的商品数据。
// 返回值：创建成功返回商品；SKU 重复或数据库操作失败时返回错误。
func (service *ProductService) CreateProduct(
	ctx context.Context,
	input CreateProductInput,
) (model.Product, error) {
	product := model.Product{
		Name:       input.Name,
		SKU:        input.SKU,
		PriceCents: input.PriceCents,
		Stock:      input.Stock,
		Status:     input.Status,
	}

	if err := service.store.CreateProduct(ctx, &product); err != nil {
		return model.Product{}, err
	}

	return product, nil
}

// ListProducts 查询商品列表并计算分页元数据。
// 参数：ctx 控制查询取消和超时；filter 为状态和分页条件。
// 返回值：查询成功返回列表及分页信息；任一数据库查询失败时返回错误。
func (service *ProductService) ListProducts(
	ctx context.Context,
	filter ProductFilter,
) (ProductList, error) {
	products, err := service.store.ListProducts(
		ctx,
		filter.Status,
		filter.Page,
		filter.PageSize,
	)
	if err != nil {
		return ProductList{}, err
	}

	total, err := service.store.CountProducts(ctx, filter.Status)
	if err != nil {
		return ProductList{}, err
	}

	totalPages := (total + int64(filter.PageSize) - 1) /
		int64(filter.PageSize)

	return ProductList{
		Items:      products,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// UpdateProduct 更新指定商品。
// 参数：ctx 控制操作取消和超时；id 为商品 ID；input 为待更新数据。
// 返回值：更新成功返回最新商品；商品不存在或更新失败时返回错误。
func (service *ProductService) UpdateProduct(
	ctx context.Context,
	id int64,
	input UpdateProductInput,
) (model.Product, error) {
	return service.store.UpdateProduct(
		ctx,
		id,
		input.Name,
		input.PriceCents,
		input.Stock,
		input.Status,
	)
}

// DeleteProduct 软删除指定商品。
// 参数：ctx 控制操作取消和超时；id 为商品 ID。
// 返回值：删除成功返回 nil；商品不存在或删除失败时返回错误。
func (service *ProductService) DeleteProduct(
	ctx context.Context,
	id int64,
) error {
	return service.store.DeleteProduct(ctx, id)
}

// SellProduct 扣减指定商品库存。
// 参数：ctx 控制事务取消和超时；id 为商品 ID；quantity 为销售数量。
// 返回值：销售成功返回 nil；商品不可销售或事务失败时返回错误。
func (service *ProductService) SellProduct(
	ctx context.Context,
	id int64,
	quantity int,
) error {
	return service.store.SellProduct(ctx, id, quantity)
}

// AddProductImage 为指定商品添加图片。
// 参数：ctx 控制操作取消和超时；productID 为商品 ID；url 为图片地址。
// 返回值：创建成功返回图片；商品不存在或创建失败时返回错误。
func (service *ProductService) AddProductImage(
	ctx context.Context,
	productID int64,
	url string,
) (model.ProductImage, error) {
	return service.store.AppendProductImage(ctx, productID, url)
}

// DeleteProductImage 删除指定商品图片。
// 参数：ctx 控制操作取消和超时；productID 为商品 ID；imageID 为图片 ID。
// 返回值：删除成功返回 nil；图片不存在或删除失败时返回错误。
func (service *ProductService) DeleteProductImage(
	ctx context.Context,
	productID int64,
	imageID int64,
) error {
	return service.store.DeleteProductImage(ctx, productID, imageID)
}

// DisableProducts 批量禁用商品。
// 参数：ctx 控制操作取消和超时；ids 为待禁用商品 ID。
// 返回值：返回实际更新行数；数据库操作失败时返回错误。
func (service *ProductService) DisableProducts(
	ctx context.Context,
	ids []int64,
) (int64, error) {
	return service.store.DisableProducts(ctx, ids)
}
