package main

import (
	"fmt"
	"net/http"
	"time"

	"vida-go/internal/api/middleware"
	"vida-go/internal/config"
	"vida-go/internal/infra/database"
	infraRedis "vida-go/internal/infra/redis"
	"vida-go/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 加载配置文件
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// 初始化日志系统
	if err := logger.Init(
		cfg.Log.Level,
		cfg.Log.Format,
		cfg.Log.Output,
		cfg.Log.FilePath,
	); err != nil {
		panic(fmt.Sprintf("Failed to init logger: %v", err))
	}
	defer logger.Sync()

	// 初始化数据库
	if err := database.Init(&cfg.Database); err != nil {
		logger.Fatal("Failed to init database", zap.Error(err))
	}
	defer database.Close()

	// 初始化Redis
	if err := infraRedis.Init(&cfg.Redis); err != nil {
		logger.Fatal("Failed to init redis", zap.Error(err))
	}
	defer infraRedis.Close()

	// 设置Gin模式
	gin.SetMode(cfg.App.Mode)

	// 创建Gin路由器（不使用默认中间件）
	router := gin.New()

	// 使用自定义中间件
	router.Use(middleware.Recovery())
	router.Use(middleware.Logger())

	// 注册路由
	router.GET("/healthz", healthCheckHandler)
	router.GET("/", rootHandler)

	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.App.Port)
	logger.Info("Starting application",
		zap.String("name", cfg.App.Name),
		zap.String("version", cfg.App.Version),
		zap.String("mode", cfg.App.Mode),
		zap.String("addr", addr),
	)
	logger.Info("Configuration loaded",
		zap.String("database", fmt.Sprintf("%s@%s:%d/%s", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)),
		zap.String("redis", cfg.Redis.Addr()),
		zap.String("minio", cfg.MinIO.Endpoint),
		zap.String("agent", cfg.Agent.URL),
	)

	// 启动HTTP服务器
	logger.Info("Server listening", zap.String("addr", addr))
	if err := router.Run(addr); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

// healthCheckHandler 健康检查接口
func healthCheckHandler(c *gin.Context) {
	cfg := config.Get()
	
	logger.Debug("Health check requested", zap.String("ip", c.ClientIP()))
	
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"message":   "Service is healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   cfg.App.Name,
		"version":   cfg.App.Version,
		"mode":      cfg.App.Mode,
	})
}

// rootHandler 根路径处理器
func rootHandler(c *gin.Context) {
	cfg := config.Get()
	
	logger.Info("Root endpoint accessed", zap.String("ip", c.ClientIP()))
	
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Welcome to %s API", cfg.App.Name),
		"project": cfg.App.Name,
		"version": cfg.App.Version,
		"mode":    cfg.App.Mode,
		"docs":    fmt.Sprintf("http://localhost:%d/healthz", cfg.App.Port),
	})
}
