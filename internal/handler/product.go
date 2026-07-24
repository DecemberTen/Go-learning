package handler

import (
	"errors"
	"net/http"
	"strconv"

	"example.com/go-learning/internal/response"
	"example.com/go-learning/internal/service"
	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productService *service.ProductService
}

// NewProductHandler 创建商品 HTTP Handler。
// 参数：productService 提供商品业务能力。
// 返回值：返回可注册商品路由的 ProductHandler。
func NewProductHandler(
	productService *service.ProductService,
) *ProductHandler {
	return &ProductHandler{productService: productService}
}

// RegisterRoutes 注册健康检查和商品相关路由。
// 参数：router 为 Gin 路由引擎。
// 返回值：无，路由直接注册到传入引擎。
func (handler *ProductHandler) RegisterRoutes(router *gin.Engine) {
	router.GET("/hello", hello)

	products := router.Group("/products")
	products.GET("/:id", handler.getProduct)
	products.GET("", handler.listProducts)
	products.POST("", handler.createProduct)
	products.PUT("/:id", handler.updateProduct)
	products.DELETE("/:id", handler.deleteProduct)
	products.POST("/:id/sell", handler.sellProduct)
	products.POST("/:id/images", handler.createProductImage)
	products.DELETE(
		"/:id/images/:image_id",
		handler.deleteProductImage,
	)
}

// hello 返回服务可访问的简单消息。
// 参数：context 为当前 Gin 请求上下文。
// 返回值：无，消息直接写入 HTTP 响应。
func hello(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{
		"message": "Hello, World!",
	})
}

// parsePositiveID 将路径参数转换为正整数 ID。
// 参数：value 为路径中的数字文本。
// 返回值：转换成功返回 ID；格式无效或数值不为正数时返回错误。
func parsePositiveID(value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("ID 必须是正整数")
	}
	return id, nil
}

// getProduct 查询并返回单个商品。
// 参数：context 为当前 Gin 请求上下文，包含商品 ID。
// 返回值：无，查询结果直接写入 HTTP 响应。
func (handler *ProductHandler) getProduct(context *gin.Context) {
	id, err := parsePositiveID(context.Param("id"))
	if err != nil {
		response.RespondError(
			context,
			http.StatusBadRequest,
			response.ErrorCodeInvalidRequest,
			"Invalid product ID",
		)
		return
	}

	product, err := handler.productService.GetProduct(
		context.Request.Context(),
		id,
	)
	if err != nil {
		response.HandleError(context, err)
		return
	}

	context.JSON(http.StatusOK, product)
}

// createProduct 校验请求并创建商品。
// 参数：context 为当前 Gin 请求上下文，包含商品请求体。
// 返回值：无，创建结果直接写入 HTTP 响应。
func (handler *ProductHandler) createProduct(context *gin.Context) {
	var input CreateProductRequest
	if err := context.ShouldBindJSON(&input); err != nil {
		response.RespondError(
			context,
			http.StatusBadRequest,
			response.ErrorCodeInvalidRequest,
			"Invalid request body",
		)
		return
	}

	product, err := handler.productService.CreateProduct(
		context.Request.Context(),
		service.CreateProductInput{
			Name:       input.Name,
			SKU:        input.SKU,
			PriceCents: input.PriceCents,
			Stock:      input.Stock,
			Status:     input.Status,
		},
	)
	if err != nil {
		response.HandleError(context, err)
		return
	}

	context.JSON(http.StatusCreated, product)
}

// listProducts 查询商品列表和分页信息。
// 参数：context 为当前 Gin 请求上下文，包含筛选与分页参数。
// 返回值：无，列表结果直接写入 HTTP 响应。
func (handler *ProductHandler) listProducts(context *gin.Context) {
	var query ListProductsQuery
	if err := context.ShouldBindQuery(&query); err != nil {
		response.RespondError(
			context,
			http.StatusBadRequest,
			response.ErrorCodeInvalidRequest,
			"Invalid query parameters",
		)
		return
	}

	result, err := handler.productService.ListProducts(
		context.Request.Context(),
		service.ProductFilter{
			Status:   query.Status,
			Page:     query.Page,
			PageSize: query.PageSize,
		},
	)
	if err != nil {
		response.HandleError(context, err)
		return
	}

	context.JSON(http.StatusOK, ProductListResponse{
		Items:      result.Items,
		Page:       result.Page,
		PageSize:   result.PageSize,
		Total:      result.Total,
		TotalPages: result.TotalPages,
	})
}

// updateProduct 校验请求并更新商品。
// 参数：context 为当前 Gin 请求上下文，包含商品 ID 和请求体。
// 返回值：无，更新结果直接写入 HTTP 响应。
func (handler *ProductHandler) updateProduct(context *gin.Context) {
	id, err := parsePositiveID(context.Param("id"))
	if err != nil {
		response.RespondError(
			context,
			http.StatusBadRequest,
			response.ErrorCodeInvalidRequest,
			"Invalid product ID",
		)
		return
	}

	var input UpdateProductRequest
	if err := context.ShouldBindJSON(&input); err != nil {
		response.RespondError(
			context,
			http.StatusBadRequest,
			response.ErrorCodeInvalidRequest,
			"Invalid request body",
		)
		return
	}

	product, err := handler.productService.UpdateProduct(
		context.Request.Context(),
		id,
		service.UpdateProductInput{
			Name:       input.Name,
			PriceCents: input.PriceCents,
			Stock:      input.Stock,
			Status:     input.Status,
		},
	)
	if err != nil {
		response.HandleError(context, err)
		return
	}

	context.JSON(http.StatusOK, product)
}

// deleteProduct 软删除指定商品。
// 参数：context 为当前 Gin 请求上下文，包含商品 ID。
// 返回值：无，删除成功返回 204，失败时写入错误响应。
func (handler *ProductHandler) deleteProduct(context *gin.Context) {
	id, err := parsePositiveID(context.Param("id"))
	if err != nil {
		response.RespondError(
			context,
			http.StatusBadRequest,
			response.ErrorCodeInvalidRequest,
			"Invalid product ID",
		)
		return
	}

	err = handler.productService.DeleteProduct(
		context.Request.Context(),
		id,
	)
	if err != nil {
		response.HandleError(context, err)
		return
	}

	context.Status(http.StatusNoContent)
}

// sellProduct 校验请求并扣减商品库存。
// 参数：context 为当前 Gin 请求上下文，包含商品 ID 和销售数量。
// 返回值：无，销售成功返回 204，失败时写入错误响应。
func (handler *ProductHandler) sellProduct(context *gin.Context) {
	id, err := parsePositiveID(context.Param("id"))
	if err != nil {
		response.RespondError(
			context,
			http.StatusBadRequest,
			response.ErrorCodeInvalidRequest,
			"Invalid product ID",
		)
		return
	}

	var input SellProductRequest
	if err := context.ShouldBindJSON(&input); err != nil {
		response.RespondError(
			context,
			http.StatusBadRequest,
			response.ErrorCodeInvalidRequest,
			"Invalid request body",
		)
		return
	}

	err = handler.productService.SellProduct(
		context.Request.Context(),
		id,
		input.Quantity,
	)
	if err != nil {
		response.HandleError(context, err)
		return
	}

	context.Status(http.StatusNoContent)
}

// createProductImage 为指定商品创建关联图片。
// 参数：context 为当前 Gin 请求上下文，包含商品 ID 和图片请求体。
// 返回值：无，创建结果直接写入 HTTP 响应。
func (handler *ProductHandler) createProductImage(
	context *gin.Context,
) {
	productID, err := parsePositiveID(context.Param("id"))
	if err != nil {
		response.RespondError(
			context,
			http.StatusBadRequest,
			response.ErrorCodeInvalidRequest,
			"Invalid product ID",
		)
		return
	}

	var input CreateProductImageRequest
	if err := context.ShouldBindJSON(&input); err != nil {
		response.RespondError(
			context,
			http.StatusBadRequest,
			response.ErrorCodeInvalidRequest,
			"Invalid request body",
		)
		return
	}

	image, err := handler.productService.AddProductImage(
		context.Request.Context(),
		productID,
		input.URL,
	)
	if err != nil {
		response.HandleError(context, err)
		return
	}

	context.JSON(http.StatusCreated, image)
}

// deleteProductImage 删除指定商品下的一张图片。
// 参数：context 为当前 Gin 请求上下文，包含商品 ID 和图片 ID。
// 返回值：无，删除成功返回 204，失败时写入错误响应。
func (handler *ProductHandler) deleteProductImage(
	context *gin.Context,
) {
	productID, err := parsePositiveID(context.Param("id"))
	if err != nil {
		response.RespondError(
			context,
			http.StatusBadRequest,
			response.ErrorCodeInvalidRequest,
			"Invalid product ID",
		)
		return
	}

	imageID, err := parsePositiveID(context.Param("image_id"))
	if err != nil {
		response.RespondError(
			context,
			http.StatusBadRequest,
			response.ErrorCodeInvalidRequest,
			"Invalid image ID",
		)
		return
	}

	err = handler.productService.DeleteProductImage(
		context.Request.Context(),
		productID,
		imageID,
	)
	if err != nil {
		response.HandleError(context, err)
		return
	}

	context.Status(http.StatusNoContent)
}
