package entities

import "time"

type SchedulerStatus string

const (
	SchedulerStatusStopped SchedulerStatus = "stopped"
	SchedulerStatusRunning SchedulerStatus = "running"
)

type SchedulerInfo struct {
	Status        SchedulerStatus `json:"status"`
	LastRun       *time.Time      `json:"last_run,omitempty"`
	NextRun       *time.Time      `json:"next_run,omitempty"`
	MessagesCount int             `json:"messages_sent_count"`
	Interval      time.Duration   `json:"interval"`
	BatchSize     int             `json:"batch_size"`
}

func (s *SchedulerInfo) IsRunning() bool {
	return s.Status == SchedulerStatusRunning
}

func (s *SchedulerInfo) Start() {
	s.Status = SchedulerStatusRunning
	now := time.Now()
	s.NextRun = &now
}

func (s *SchedulerInfo) Stop() {
	s.Status = SchedulerStatusStopped
	s.NextRun = nil
}
