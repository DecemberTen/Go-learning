package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func hello(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{
		"message": "Hello, World!",
	})
}

func getProduct(context *gin.Context) {
	idText := context.Param("id")
	id, err := strconv.ParseInt(idText, 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid product ID",
		})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"id":   id,
		"name": "哈哈",
	})
}

func listProducts(context *gin.Context) {
	keyword := context.Query("keyword")
	limitText := context.DefaultQuery("limit", "20")

	limit, err := strconv.Atoi(limitText)
	if err != nil || limit <= 0 {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid limit",
		})
		return
	}
	context.JSON(http.StatusOK, gin.H{
		"keyword": keyword,
		"limit":   limit,
	})
}

func createProduct(c *gin.Context) {
	var req CreateProductRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"name":        req.Name,
		"price_cents": req.PriceCents,
		"stock":       req.Stock,
		"status":      req.Status,
	})
}

func requestTimer(c *gin.Context) {
	start := time.Now()

	fmt.Println("请求开始")
	c.Next()

	duration := time.Since(start)
	fmt.Printf(
		"请求结束：method=%s path=%s status=%d duration=%v\n",
		c.Request.Method,
		c.FullPath(),
		c.Writer.Status(),
		duration,
	)
}

// func tokenMiddleware(c *gin.Context) {
// 	token := c.GetHeader("Authorization")
// 	if token != "Bearer abc123" {
// 		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
// 			"error": "Invalid token",
// 		})
// 		return
// 	}
// 	c.Set("user_id", int64(123))
// 	c.Next()
// }

func profile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "User ID not found",
		})
		return
	}
	userID, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile",
		"user_id": userID,
	})
}
func updateProduct(c *gin.Context) {
	idText := c.Param("id")
	id, err := strconv.ParseInt(idText, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid product ID",
		})
		return
	}

	var product UpdateProductRequest
	err = c.ShouldBindJSON(&product)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":          id,
		"name":        product.Name,
		"price_cents": product.PriceCents,
		"stock":       product.Stock,
		"status":      product.Status,
	})
}

func deleteProduct(c *gin.Context) {
	idText := c.Param("id")
	id, err := strconv.ParseInt(idText, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid product ID",
		})
		return
	}
	c.Status(http.StatusNoContent)
}

func notFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"error": "Route not found",
	})
}

// methodNotAllowed 用于处理路径存在但 HTTP 方法不支持的请求；参数 c 包含请求和响应信息；返回值为空。
func methodNotAllowed(c *gin.Context) {
	c.JSON(http.StatusMethodNotAllowed, gin.H{
		"error": "Method not allowed",
	})
}

func main() {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:5173",
		},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
		},
	}))

	router.HandleMethodNotAllowed = true
	router.NoMethod(methodNotAllowed)
	router.NoRoute(notFound)

	sqlDB, err := openDatabase(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	log.Println("MySQL连接成功")

	gormDB, err := openGormDatabase(sqlDB)
	if err != nil {
		log.Fatal(err)
	}

	// product, err := queryProductByIDWithGorm(
	// 	context.Background(),
	// 	gormDB,
	// 	1,
	// )
	// if err != nil {
	// 	if errors.Is(err, gorm.ErrRecordNotFound) {
	// 		log.Println("商品不存在")
	// 	} else {
	// 		log.Println(err)
	// 	}
	// } else {
	// 	log.Printf("查询成功: %+v", product)
	// }

	// products, err := queryProductsWithGorm(
	// 	context.Background(),
	// 	gormDB,
	// )
	// if err != nil {
	// 	log.Println(err)
	// } else {
	// 	for _, product := range products {
	// 		log.Printf("商品: %+v", product)
	// 	}
	// }

	// product := Product{
	// 	Name:       "GORM商品",
	// 	PriceCents: 1999,
	// 	Stock:      10,
	// 	Status:     "active",
	// }

	// err = createProductWithGorm(
	// 	context.Background(),
	// 	gormDB,
	// 	&product,
	// )
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// store := newProductStore()
	// handler := newProductHandler(store, sqlDB, gormDB)
	handler := newProductHandler(gormDB)

	handler.RegisterRoutes(router)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET 未设置")
	}

	authHandler := newAuthHandler(gormDB, jwtSecret)
	authHandler.RegisterRoutes(router)

	// router.Use(requestTimer)

	// products, err := queryProducts(context.Background(), db)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// for _, product := range products {
	// 	log.Printf(
	// 		"商品ID=%d，名称=%s，价格=%d，库存=%d，状态=%s",
	// 		product.ID,
	// 		product.Name,
	// 		product.PriceCents,
	// 		product.Stock,
	// 		product.Status,
	// 	)
	// }

	err = router.Run()
	if err != nil {
		fmt.Println(err)
	}
}
