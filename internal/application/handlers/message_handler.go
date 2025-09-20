package handlers

import (
	"errors"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"message-sending-service/internal/application/dto"
	"message-sending-service/internal/domain/entities"
	"message-sending-service/internal/domain/usecases"
)

// dependency injection yapmak icin struct + construction
type MessageHandler struct {
	messageUseCase usecases.MessageUseCase
	logger         *zap.Logger
}

func NewMessageHandler(messageUseCase usecases.MessageUseCase, logger *zap.Logger) *MessageHandler {
	return &MessageHandler{
		messageUseCase: messageUseCase,
		logger:         logger,
	}
}

func (h *MessageHandler) CreateMessage(c *gin.Context) {
	var req dto.CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("invalid_request", err.Error(), http.StatusBadRequest))
		return
	}

	message, err := h.messageUseCase.CreateMessage(c.Request.Context(), req.Content, req.PhoneNumber)
	if err != nil {
		if errors.Is(err, entities.ErrInvalidMessageContent) || errors.Is(err, entities.ErrMessageTooLong) || errors.Is(err, entities.ErrInvalidPhoneNumber) {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse("validation_error", err.Error(), http.StatusBadRequest))
			return
		}

		h.logger.Error("Failed to create message", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("internal_error", "Failed to create message", http.StatusInternalServerError))
		return
	}

	response := dto.ToMessageResponse(message)
	c.JSON(http.StatusCreated, dto.NewSuccessResponse("Message created successfully", response))
}

func (h *MessageHandler) GetMessage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("invalid_id", "Invalid message ID format", http.StatusBadRequest))
		return
	}

	message, err := h.messageUseCase.GetMessageByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, entities.ErrMessageNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse("not_found", "Message not found", http.StatusNotFound))
			return
		}

		h.logger.Error("Failed to get message", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("internal_error", "Failed to get message", http.StatusInternalServerError))
		return
	}

	response := dto.ToMessageResponse(message)
	c.JSON(http.StatusOK, dto.NewSuccessResponse("Message retrieved successfully", response))
}

func (h *MessageHandler) GetSentMessages(c *gin.Context) {
	var query dto.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("invalid_query", err.Error(), http.StatusBadRequest))
		return
	}

	messages, totalCount, err := h.messageUseCase.GetSentMessages(c.Request.Context(), query.Page, query.Limit)
	if err != nil {
		h.logger.Error("Failed to get sent messages", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("internal_error", "Failed to get sent messages", http.StatusInternalServerError))
		return
	}

	messageResponses := make([]dto.MessageResponse, len(messages))
	for i, message := range messages {
		messageResponses[i] = dto.ToMessageResponse(message)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(query.Limit)))

	response := dto.GetSentMessagesResponse{
		Messages:   messageResponses,
		TotalCount: totalCount,
		Page:       query.Page,
		Limit:      query.Limit,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Sent messages retrieved successfully", response))
}

func (h *MessageHandler) GetMessageStats(c *gin.Context) {
	stats, err := h.messageUseCase.GetMessageStats(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get message stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("internal_error", "Failed to get message statistics", http.StatusInternalServerError))
		return
	}

	response := dto.ToMessageStatsResponse(stats)
	c.JSON(http.StatusOK, dto.NewSuccessResponse("Message statistics retrieved successfully", response))
}

func (h *MessageHandler) SendMessage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("invalid_id", "Invalid message ID format", http.StatusBadRequest))
		return
	}

	message, err := h.messageUseCase.GetMessageByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, entities.ErrMessageNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse("not_found", "Message not found", http.StatusNotFound))
			return
		}

		h.logger.Error("Failed to get message for sending", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("internal_error", "Failed to get message", http.StatusInternalServerError))
		return
	}

	if !message.IsPending() {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("invalid_status", "Message is not in pending status", http.StatusBadRequest))
		return
	}

	if err := h.messageUseCase.SendMessage(c.Request.Context(), message); err != nil {
		h.logger.Error("Failed to send message", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("send_error", "Failed to send message", http.StatusInternalServerError))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Message sent successfully", nil))
}
