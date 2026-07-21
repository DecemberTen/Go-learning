package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductHandler struct {
	// store  *ProductStore
	// db     *sql.DB
	gormDB *gorm.DB
}

// func newProductHandler(store *ProductStore, db *sql.DB, gormDB *gorm.DB) *ProductHandler {
func newProductHandler(gormDB *gorm.DB) *ProductHandler {
	return &ProductHandler{
		// store:  store,
		// db:     db,
		gormDB: gormDB,
	}
}

func (handler *ProductHandler) RegisterRoutes(router *gin.Engine) {
	router.GET("/hello", hello)

	productGroup := router.Group("/products")

	// productGroup.Use(tokenMiddleware)

	productGroup.GET("/:id", handler.getProduct)

	productGroup.GET("", handler.listProducts)

	productGroup.POST("", handler.createdProduct)

	productGroup.PUT("/:id", handler.updateProduct)

	productGroup.DELETE("/:id", handler.deleteProduct)

	productGroup.POST("/:id/sell", handler.sellProductHandler)

	productGroup.POST("/:id/images", handler.createProductImage)

	productGroup.DELETE("/:id/images/:image_id", handler.deleteProductImage)

	// router.GET("/profile", tokenMiddleware, profile)
}

func (handler *ProductHandler) getProduct(c *gin.Context) {
	idText := c.Param("id")

	id, err := strconv.ParseInt(idText, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid product ID",
		})
		return
	}

	// product, exists := handler.store.Get(id)
	// if !exists {
	// 	c.JSON(http.StatusNotFound, gin.H{
	// 		"error": "Product not found",
	// 	})
	// 	return
	// }

	// product, err := queryProductByID(c.Request.Context(), handler.db, id)
	product, err := queryProductByIDWithGorm(c.Request.Context(), handler.gormDB, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Product not found",
		})
		return
	}
	if err != nil {
		log.Printf("查询商品失败: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})
		return
	}

	c.JSON(http.StatusOK, product)
}

func (handler *ProductHandler) createdProduct(c *gin.Context) {
	var input CreateProductRequest

	err := c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// createdProduct := handler.store.Create(input)
	// c.JSON(201, createdProduct)

	// product, err := insertProduct(c.Request.Context(), handler.db, input)
	// if err != nil {
	// 	log.Printf("创建失败: %v", err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{
	// 		"error": "Internal Server Error",
	// 	})
	// 	return
	// }

	product := Product{
		Name:       input.Name,
		SKU:        input.SKU,
		PriceCents: input.PriceCents,
		Stock:      input.Stock,
		Status:     input.Status,
	}

	err = createProductWithGorm(c.Request.Context(), handler.gormDB, &product)

	if errors.Is(err, ErrDuplicateSKU) {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Product SKU already exists",
		})
		return
	}

	if err != nil {
		log.Printf("创建商品失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})
		return
	}

	c.JSON(http.StatusCreated, product)
	// c.JSON(201, product)
	// return
}

func (handler *ProductHandler) listProducts(c *gin.Context) {
	// products := handler.store.List()
	// c.JSON(200, products)

	var filter ListProductsQuery

	err := c.ShouldBindQuery(&filter)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// products, err := queryProducts(c.Request.Context(), handler.db, filter)
	products, err := queryProductsWithGorm(c.Request.Context(), handler.gormDB, filter)
	if err != nil {
		log.Printf("查询商品列表失败: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})
		return
	}

	// total, err := countProducts(c.Request.Context(), handler.db, filter)
	total, err := countProductsWithGorm(c.Request.Context(), handler.gormDB, filter)
	if err != nil {
		log.Printf("查询商品总数失败: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})
		return
	}

	totalPages := (total + int64(filter.PageSize) - 1) / int64(filter.PageSize)

	response := ProductListResponse{
		Items:      products,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}
	c.JSON(http.StatusOK, response)
}

func (handler *ProductHandler) updateProduct(c *gin.Context) {
	idText := c.Param("id")
	id, err := strconv.ParseInt(idText, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid product ID",
		})
		return
	}

	var input UpdateProductRequest

	err = c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// product, err := updatedProduct(c.Request.Context(), handler.db, id, input)
	product, err := updateProductWithGorm(c.Request.Context(), handler.gormDB, id, input)

	// if errors.Is(err, sql.ErrNoRows) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Product not found",
		})
		return
	}

	if err != nil {
		log.Printf("更新商品失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})
		return
	}
	c.JSON(http.StatusOK, product)

	// product, exists := handler.store.Update(id, input)
	// if !exists {
	// 	c.JSON(http.StatusNotFound, gin.H{
	// 		"error": "Product not found",
	// 	})
	// 	return
	// }

	// c.JSON(http.StatusOK, product)
}

func (handler *ProductHandler) deleteProduct(c *gin.Context) {
	idText := c.Param("id")
	id, err := strconv.ParseInt(idText, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid product ID",
		})
		return
	}

	// Err := delProductByID(c.Request.Context(), handler.db, id)
	Err := deleteProductWithGorm(c.Request.Context(), handler.gormDB, id)

	if errors.Is(Err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Product not found",
		})
		return
	}
	if Err != nil {
		log.Printf("删除商品失败: %v", Err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})
		return
	}

	c.Status(http.StatusNoContent)
	// exists := handler.store.Delete(id)
	// if !exists {
	// 	c.JSON(http.StatusNotFound, gin.H{
	// 		"error": "Product not found",
	// 	})
	// 	return
	// }
	// c.Status(http.StatusNoContent)
}

func (handler *ProductHandler) sellProductHandler(c *gin.Context) {
	idText := c.Param("id")
	id, err := strconv.ParseInt(idText, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid product ID",
		})
		return
	}
	var input SellProductRequest
	if err = c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// err = sellProduct(c.Request.Context(), handler.db, id, input.Quantity)
	err = sellProductWithGorm(c.Request.Context(), handler.gormDB, id, input.Quantity)

	if errors.Is(err, ErrProductUnavailable) {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Product not available",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})
		return
	}
	c.Status(http.StatusNoContent)
}

// createProductImage 为指定商品创建关联图片。
// 参数：c 为当前 Gin 请求上下文，包含商品 ID 和图片请求体。
// 返回值：无；创建结果写入 HTTP 响应。
func (handler *ProductHandler) createProductImage(c *gin.Context) {
	idText := c.Param("id")

	id, err := strconv.ParseInt(idText, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid product ID",
		})
		return
	}

	var input CreateProductImageRequest

	err = c.ShouldBindJSON(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	img, err := appendProductImageWithGorm(c.Request.Context(), handler.gormDB, id, input.URL)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Product not found",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})
		return
	}
	c.JSON(http.StatusCreated, img)
}

// deleteProductImage 删除指定商品下的一张图片。
// 参数：c 为当前 Gin 请求上下文，包含商品 ID 和图片 ID。
// 返回值：无；删除成功返回 204，失败时写入错误响应。
func (handler *ProductHandler) deleteProductImage(c *gin.Context) {
	productID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid product ID",
		})
		return
	}

	imageID, err := strconv.ParseInt(c.Param("image_id"), 10, 64)
	if err != nil || imageID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid image ID",
		})
		return
	}

	err = deleteProductImageWithGorm(c.Request.Context(), handler.gormDB, productID, imageID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Product image not found",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})
		return
	}
	c.Status(http.StatusNoContent)
}
