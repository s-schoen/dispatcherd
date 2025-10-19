package handler

import (
	"dispatcherd/dispatch"
	"dispatcherd/service"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type dispatchPostBody struct {
	Title   string            `json:"title" validate:"required"`
	Message string            `json:"message" validate:"required"`
	Tags    map[string]string `json:"tags"`
}

type dispatchPostResponse struct {
	MessageID string `json:"messageId"`
}

type MessageHandler struct {
	logger     *slog.Logger
	validate   *validator.Validate
	messageSvc service.MessageService
}

func NewDispatchHandler(logger *slog.Logger, msgSvc service.MessageService) *MessageHandler {
	return &MessageHandler{
		logger:     logger,
		validate:   validator.New(validator.WithRequiredStructEnabled()),
		messageSvc: msgSvc,
	}
}

func (h *MessageHandler) HandlePost(w http.ResponseWriter, r *http.Request) error {
	var body dispatchPostBody
	if err := ParseAndValidateBody(&body, r, h.validate); err != nil {
		var apiErr APIError
		if errors.As(err, &apiErr) {
			return apiErr
		}
		return OtherError(err)
	}

	message := dispatch.NewMessage(body.Title, body.Message, body.Tags)

	if err := h.messageSvc.QueueMessage(r.Context(), message); err != nil {
		var apiErr APIError
		if errors.As(err, &apiErr) {
			return apiErr
		}
		return OtherError(err)
	}

	return respondOne(w, r, dispatchPostResponse{
		MessageID: message.ID,
	})
}
