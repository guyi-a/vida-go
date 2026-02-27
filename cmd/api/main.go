package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"vida-go/internal/api/handler"
	"vida-go/internal/api/middleware"
	"vida-go/internal/api/router"
	"vida-go/internal/config"
	"vida-go/internal/infra/database"
	infraES "vida-go/internal/infra/elasticsearch"
	infraKafka "vida-go/internal/infra/kafka"
	infraMinio "vida-go/internal/infra/minio"
	infraRedis "vida-go/internal/infra/redis"
	"vida-go/internal/model"
	"vida-go/internal/repository"
	"vida-go/internal/service"
	"vida-go/pkg/logger"

	_ "vida-go/api/openapi"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// @title Vida-Go API
// @version 1.0
// @description 视频分享平台 API 服务
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@vida.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host 127.0.0.1:8000
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 输入格式: Bearer {token}

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

	// 自动迁移数据库表
	if err := database.AutoMigrate(
		&model.User{},
		&model.Video{},
		&model.Comment{},
		&model.Favorite{},
		&model.Relation{},
	); err != nil {
		logger.Fatal("Failed to auto migrate", zap.Error(err))
	}

	// 初始化Redis
	if err := infraRedis.Init(&cfg.Redis); err != nil {
		logger.Fatal("Failed to init redis", zap.Error(err))
	}
	defer infraRedis.Close()

	// 初始化MinIO
	if err := infraMinio.Init(&cfg.MinIO); err != nil {
		logger.Fatal("Failed to init minio", zap.Error(err))
	}

	// 初始化Kafka生产者
	if err := infraKafka.InitProducer(&cfg.Kafka); err != nil {
		logger.Fatal("Failed to init kafka producer", zap.Error(err))
	}
	defer infraKafka.CloseProducer()

	// 初始化 Elasticsearch（可选，失败则搜索降级到 DB）
	if err := infraES.Init(&cfg.Elasticsearch); err != nil {
		logger.Warn("Elasticsearch init failed, search will fallback to DB", zap.Error(err))
	} else {
		defer infraES.Close()
		if err := infraES.InitIndexes(); err != nil {
			logger.Warn("Elasticsearch index init failed", zap.Error(err))
		}
	}

	// 设置Gin模式
	gin.SetMode(cfg.App.Mode)

	// 创建Gin路由器（不使用默认中间件）
	r := gin.New()

	// 使用自定义中间件
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())

	// 初始化依赖（Repository -> Service -> Handler）
	db := database.Get()
	userRepo := repository.NewUserRepository(db)
	relationRepo := repository.NewRelationRepository(db)

	videoRepo := repository.NewVideoRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)

	authService := service.NewAuthService(userRepo)
	userService := service.NewUserService(userRepo)
	relationService := service.NewRelationService(relationRepo, userRepo)
	videoService := service.NewVideoService(videoRepo)
	commentService := service.NewCommentService(commentRepo, videoRepo)
	favoriteService := service.NewFavoriteService(favoriteRepo, videoRepo, userRepo)
	searchService := service.NewSearchService(videoRepo)

	// 启动转码结果消费者（后台 goroutine）
	consumerCtx, consumerCancel := context.WithCancel(context.Background())
	defer consumerCancel()

	if topic, ok := cfg.Kafka.Topics["video_uploaded"]; ok {
		resultHandler := func(result *infraKafka.TranscodeResult) error {
			if err := videoService.HandleTranscodeResult(result); err != nil {
				return err
			}
			if result.Status == "published" {
				_ = searchService.SyncVideoToES(result.VideoID)
			}
			return nil
		}
		go infraKafka.StartTranscodeResultConsumer(
			consumerCtx,
			cfg.Kafka.Brokers,
			topic,
			"vida-go-transcode-result",
			resultHandler,
		)
	}

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService, authService)
	relationHandler := handler.NewRelationHandler(relationService)
	videoHandler := handler.NewVideoHandler(videoService)
	commentHandler := handler.NewCommentHandler(commentService)
	favoriteHandler := handler.NewFavoriteHandler(favoriteService)
	searchHandler := handler.NewSearchHandler(searchService)

	// 管理员中间件（需要查数据库获取角色）
	adminMiddleware := middleware.AdminRequired(func(userID int64) (string, error) {
		user, err := userRepo.GetByID(userID)
		if err != nil {
			return "", err
		}
		return user.UserRole, nil
	})

	// 注册基础路由
	r.GET("/healthz", healthCheckHandler)
	r.GET("/", rootHandler)

	// Swagger 文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 注册业务路由
	router.Setup(r, authHandler, userHandler, relationHandler, videoHandler, commentHandler, favoriteHandler, searchHandler, adminMiddleware)

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
	if err := r.Run(addr); err != nil {
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
