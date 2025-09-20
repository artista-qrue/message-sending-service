package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"message-sending-service/internal/application/handlers"
	"message-sending-service/internal/application/middlewares"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Router struct {
	messageHandler   *handlers.MessageHandler
	schedulerHandler *handlers.SchedulerHandler
	logger           *zap.Logger
}

func NewRouter(
	messageHandler *handlers.MessageHandler,
	schedulerHandler *handlers.SchedulerHandler,
	logger *zap.Logger,
) *Router {
	return &Router{
		messageHandler:   messageHandler,
		schedulerHandler: schedulerHandler,
		logger:           logger,
	}
}

func (r *Router) SetupRoutes() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(middlewares.CORSMiddleware())
	router.Use(middlewares.RecoveryMiddleware(r.logger))
	router.Use(middlewares.RequestLoggingMiddleware(r.logger))
	router.GET("/health", r.healthCheck)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	v1 := router.Group("/api/v1")
	{
		messages := v1.Group("/messages")
		{
			messages.POST("", r.messageHandler.CreateMessage)
			messages.GET("/:id", r.messageHandler.GetMessage)
			messages.GET("/sent", r.messageHandler.GetSentMessages)
			messages.GET("/stats", r.messageHandler.GetMessageStats)
			messages.POST("/:id/send", r.messageHandler.SendMessage)
		}

		scheduler := v1.Group("/scheduler")
		{
			scheduler.POST("/start", r.schedulerHandler.StartScheduler)
			scheduler.POST("/stop", r.schedulerHandler.StopScheduler)
			scheduler.GET("/status", r.schedulerHandler.GetSchedulerStatus)
		}
	}

	return router
}

func (r *Router) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "message-sending-service",
		"version": "1.0.0",
	})
}
