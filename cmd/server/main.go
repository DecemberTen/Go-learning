package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	appconfig "example.com/go-learning/internal/config"
	"example.com/go-learning/internal/database"
	"example.com/go-learning/internal/handler"
	"example.com/go-learning/internal/middleware"
	"example.com/go-learning/internal/repository"
	"example.com/go-learning/internal/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// main 读取配置、组装应用依赖，并管理 HTTP Server 的启动与优雅关闭。
// 参数：无。
// 返回值：无；初始化或运行失败时记录错误并结束程序。
func main() {
	appConfig, err := appconfig.Load()
	if err != nil {
		log.Printf("读取配置失败: %v", err)
		return
	}

	sqlDB, err := database.Open(
		context.Background(),
		appConfig.Database,
	)
	if err != nil {
		log.Printf("打开数据库失败: %v", err)
		return
	}
	defer sqlDB.Close()

	gormDB, err := database.OpenGORM(sqlDB)
	if err != nil {
		log.Printf("初始化 GORM 失败: %v", err)
		return
	}

	store := repository.NewStore(gormDB)
	productService := service.NewProductService(store)
	authService := service.NewAuthService(
		store,
		appConfig.Auth.JWTSecret,
	)

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
	router.NoMethod(handler.MethodNotAllowed)
	router.NoRoute(handler.NotFound)

	productHandler := handler.NewProductHandler(productService)
	productHandler.RegisterRoutes(router)

	authHandler := handler.NewAuthHandler(
		authService,
		middleware.NewAuth(authService),
		middleware.NewAdmin(authService),
	)
	authHandler.RegisterRoutes(router)

	server := &http.Server{
		Addr:              appConfig.Server.Address,
		Handler:           router,
		ReadHeaderTimeout: appConfig.Server.ReadHeaderTimeout,
		ReadTimeout:       appConfig.Server.ReadTimeout,
		WriteTimeout:      appConfig.Server.WriteTimeout,
		IdleTimeout:       appConfig.Server.IdleTimeout,
	}

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	shutdownContext, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	log.Printf("服务器启动，监听地址：%s", server.Addr)

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP Server 运行失败：%v", err)
		}
		return
	case <-shutdownContext.Done():
		log.Println("收到退出信号，开始关闭服务器")
	}

	timeoutContext, cancel := context.WithTimeout(
		context.Background(),
		appConfig.Server.ShutdownTimeout,
	)
	defer cancel()

	if err := server.Shutdown(timeoutContext); err != nil {
		log.Printf("服务器未能正常关闭：%v", err)
		return
	}

	log.Println("服务器已正常关闭")
}
