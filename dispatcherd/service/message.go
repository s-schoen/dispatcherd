package service

import (
	"context"
	"fmt"
	"log/slog"
)

type Message struct {
	Title   string
	Message string
	Tags    map[string]string
}

type MessageService interface {
	QueueMessage(ctx context.Context, message Message) error
}

type messageService struct {
	logger *slog.Logger
}

func NewMessageService(logger *slog.Logger) MessageService {
	return &messageService{
		logger: logger,
	}
}

func (s *messageService) QueueMessage(ctx context.Context, message Message) error {
	tags := "None"
	if message.Tags != nil {
		tags = ""
		for k, v := range message.Tags {
			tags += fmt.Sprintf("%s:%s;", k, v)
		}
	}
	s.logger.DebugContext(ctx, fmt.Sprintf("received message: title='%s' message='%s' tags='%s'",
		message.Title, message.Message, tags))

	return nil
}
