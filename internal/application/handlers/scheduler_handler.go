package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"message-sending-service/internal/application/dto"
	"message-sending-service/internal/domain/entities"
	"message-sending-service/internal/domain/usecases"
)

type SchedulerHandler struct {
	schedulerUseCase usecases.SchedulerUseCase
	logger           *zap.Logger
}

func NewSchedulerHandler(schedulerUseCase usecases.SchedulerUseCase, logger *zap.Logger) *SchedulerHandler {
	return &SchedulerHandler{
		schedulerUseCase: schedulerUseCase,
		logger:           logger,
	}
}

func (h *SchedulerHandler) StartScheduler(c *gin.Context) {
	err := h.schedulerUseCase.StartScheduler(c.Request.Context())
	if err != nil {
		if errors.Is(err, entities.ErrSchedulerAlreadyRunning) {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse("already_running", "Scheduler is already running", http.StatusBadRequest))
			return
		}

		h.logger.Error("Failed to start scheduler", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("internal_error", "Failed to start scheduler", http.StatusInternalServerError))
		return
	}

	response := dto.StartSchedulerResponse{
		Message: "Scheduler started successfully",
		Status:  "running",
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Scheduler started successfully", response))
}

func (h *SchedulerHandler) StopScheduler(c *gin.Context) {
	err := h.schedulerUseCase.StopScheduler(c.Request.Context())
	if err != nil {
		if errors.Is(err, entities.ErrSchedulerNotRunning) {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse("not_running", "Scheduler is not running", http.StatusBadRequest))
			return
		}

		h.logger.Error("Failed to stop scheduler", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("internal_error", "Failed to stop scheduler", http.StatusInternalServerError))
		return
	}

	response := dto.StopSchedulerResponse{
		Message: "Scheduler stopped successfully",
		Status:  "stopped",
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Scheduler stopped successfully", response))
}

func (h *SchedulerHandler) GetSchedulerStatus(c *gin.Context) {
	status, err := h.schedulerUseCase.GetSchedulerStatus(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get scheduler status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("internal_error", "Failed to get scheduler status", http.StatusInternalServerError))
		return
	}

	response := dto.ToSchedulerStatusResponse(status)
	c.JSON(http.StatusOK, dto.NewSuccessResponse("Scheduler status retrieved successfully", response))
}
