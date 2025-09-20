package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	_ "message-sending-service/docs"
	"message-sending-service/internal/application/handlers"
	"message-sending-service/internal/application/usecases"
	"message-sending-service/internal/domain/repositories"
	domainUsecases "message-sending-service/internal/domain/usecases"
	"message-sending-service/internal/infrastructure/config"
	"message-sending-service/internal/infrastructure/database"
	"message-sending-service/internal/infrastructure/external"
	infraRedis "message-sending-service/internal/infrastructure/redis"
	httpPresentation "message-sending-service/internal/presentation/http"
	"message-sending-service/pkg/logger"
)

// @title Message Sending Service API
// @version 1.0
// @description Automatic message sending system with scheduler functionality
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @schemes http https
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	zapLogger, err := logger.NewLogger(cfg.Logger.Level)
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer zapLogger.Sync()

	zapLogger.Info("Starting Message Sending Service",
		zap.String("version", "1.0.0"),
		zap.String("server_addr", cfg.GetServerAddr()))

	db, err := database.NewPostgreSQLConnection(cfg)
	if err != nil {
		zapLogger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	if err := database.CreateTables(db); err != nil {
		zapLogger.Fatal("Failed to create database tables", zap.Error(err))
	}

	redisClient := infraRedis.NewRedisClient(cfg)
	defer redisClient.Close()

	ctx := context.Background()
	if err := infraRedis.TestConnection(ctx, redisClient); err != nil {
		zapLogger.Warn("Failed to connect to Redis, caching will be disabled", zap.Error(err))
		redisClient = nil
	}

	app := initializeApp(cfg, db, redisClient, zapLogger)

	router := app.router.SetupRoutes()
	server := &http.Server{
		Addr:    cfg.GetServerAddr(),
		Handler: router,
	}

	go func() {
		zapLogger.Info("Starting HTTP server", zap.String("addr", cfg.GetServerAddr()))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	if err := app.schedulerUseCase.StartScheduler(ctx); err != nil {
		zapLogger.Error("Failed to auto-start scheduler", zap.Error(err))
	} else {
		zapLogger.Info("Scheduler auto-started successfully")
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zapLogger.Info("Shutting down server...")

	if err := app.schedulerUseCase.StopScheduler(ctx); err != nil {
		zapLogger.Error("Failed to stop scheduler", zap.Error(err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		zapLogger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	zapLogger.Info("Server exited")
}

type App struct {
	messageUseCase   domainUsecases.MessageUseCase
	schedulerUseCase domainUsecases.SchedulerUseCase
	router           *httpPresentation.Router
}

func initializeApp(cfg *config.Config, db *sql.DB, redisClient *redis.Client, logger *zap.Logger) *App {
	messageRepo := database.NewMessageRepository(db)

	var cacheRepo repositories.CacheRepository
	if redisClient != nil {
		cacheRepo = infraRedis.NewCacheRepository(redisClient)
	}

	apiClient := external.NewMessageAPIClient(cfg)

	messageUseCase := usecases.NewMessageUseCase(messageRepo, cacheRepo, apiClient, logger)
	schedulerUseCase := usecases.NewSchedulerUseCase(messageUseCase, cacheRepo, cfg, logger)

	messageHandler := handlers.NewMessageHandler(messageUseCase, logger)
	schedulerHandler := handlers.NewSchedulerHandler(schedulerUseCase, logger)

	router := httpPresentation.NewRouter(messageHandler, schedulerHandler, logger)

	return &App{
		messageUseCase:   messageUseCase,
		schedulerUseCase: schedulerUseCase,
		router:           router,
	}
}
