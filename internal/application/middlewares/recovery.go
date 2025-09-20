package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"message-sending-service/internal/application/dto"
)

// proje oldugda coktu gin customer recovery ile panicleri yakaliyoruz ve 500 donuyor burasi genisletirebilir ama case icin basit seviyede yazildi
func RecoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger.Error("Panic occurred",
			zap.Any("panic", recovered),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method))

		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(
			"internal_error",
			"Internal server error occurred",
			http.StatusInternalServerError))
	})
}
