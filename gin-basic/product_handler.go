package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	store *ProductStore
	db    *sql.DB
}

func newProductHandler(store *ProductStore, db *sql.DB) *ProductHandler {
	return &ProductHandler{
		store: store,
		db:    db,
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

	product, err := queryProductByID(c.Request.Context(), handler.db, id)
	if errors.Is(err, sql.ErrNoRows) {
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

	product, err := insertProduct(c.Request.Context(), handler.db, input)
	if err != nil {
		log.Printf("创建失败: %v", err)
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

	products, err := queryProducts(c.Request.Context(), handler.db, filter)
	if err != nil {
		log.Printf("查询商品列表失败: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})
		return
	}

	total, err := countProducts(c.Request.Context(), handler.db, filter)
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

	product, err := updatedProduct(c.Request.Context(), handler.db, id, input)

	if errors.Is(err, sql.ErrNoRows) {
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

	Err := delProductByID(c.Request.Context(), handler.db, id)

	if errors.Is(Err, sql.ErrNoRows) {
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

	err = sellProduct(c.Request.Context(), handler.db, id, input.Quantity)

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
