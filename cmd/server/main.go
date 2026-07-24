package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	appconfig "example.com/go-learning/internal/config"
	"example.com/go-learning/internal/database"
	"example.com/go-learning/internal/handler"
	applogger "example.com/go-learning/internal/logger"
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
	appLogger := applogger.New()
	slog.SetDefault(appLogger)

	appConfig, err := appconfig.Load()
	if err != nil {
		slog.Error(
			"读取配置失败",
			"error", err,
		)
		return
	}

	sqlDB, err := database.Open(
		context.Background(),
		appConfig.Database,
	)
	if err != nil {
		slog.Error(
			"打开数据库失败",
			"error", err,
		)
		return
	}
	defer sqlDB.Close()

	gormDB, err := database.OpenGORM(sqlDB)
	if err != nil {
		slog.Error(
			"初始化 GORM 失败",
			"error", err,
		)
		return
	}

	store := repository.NewStore(gormDB)
	productService := service.NewProductService(store)
	authService := service.NewAuthService(
		store,
		appConfig.Auth.JWTSecret,
	)

	// gin.Default() 默认包含：
	// gin.Logger()
	// gin.Recovery()
	// 	Gin 自带的 gin.Logger() 输出文本访问日志。我们已经有自己的 JSON 请求日志，所以改用：
	// gin.New()
	// 然后手动注册异常恢复：
	// router.Use(gin.Recovery())
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(middleware.NewRequestLogger())

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

	slog.Info(
		"服务器启动",
		"address", server.Addr,
	)

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error(
				"HTTP Server 运行失败",
				"error", err,
			)
		}
		return
	case <-shutdownContext.Done():
		slog.Info("收到退出信号，开始关闭服务器")
	}

	timeoutContext, cancel := context.WithTimeout(
		context.Background(),
		appConfig.Server.ShutdownTimeout,
	)
	defer cancel()

	if err := server.Shutdown(timeoutContext); err != nil {
		slog.Error(
			"服务器未能正常关闭",
			"error", err,
		)
		return
	}

	slog.Info("服务器已正常关闭")
}
