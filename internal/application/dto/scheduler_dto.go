package dto

import (
	"time"

	"message-sending-service/internal/domain/entities"
)

type SchedulerStatusResponse struct {
	Status        string     `json:"status" example:"running"`
	LastRun       *time.Time `json:"last_run,omitempty" example:"2023-01-01T12:00:00Z"`
	NextRun       *time.Time `json:"next_run,omitempty" example:"2023-01-01T12:02:00Z"`
	MessagesCount int        `json:"messages_sent_count" example:"150"`
	Interval      string     `json:"interval" example:"2m0s"`
	BatchSize     int        `json:"batch_size" example:"2"`
}

type StartSchedulerResponse struct {
	Message string `json:"message" example:"Scheduler started successfully"`
	Status  string `json:"status" example:"running"`
}

type StopSchedulerResponse struct {
	Message string `json:"message" example:"Scheduler stopped successfully"`
	Status  string `json:"status" example:"stopped"`
}

func ToSchedulerStatusResponse(info *entities.SchedulerInfo) SchedulerStatusResponse {
	return SchedulerStatusResponse{
		Status:        string(info.Status),
		LastRun:       info.LastRun,
		NextRun:       info.NextRun,
		MessagesCount: info.MessagesCount,
		Interval:      info.Interval.String(),
		BatchSize:     info.BatchSize,
	}
}
