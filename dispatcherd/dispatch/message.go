package dispatch

import (
	"context"
	dispatcherdContext "dispatcherd/context"
	"fmt"

	"github.com/google/uuid"
)

type Message struct {
	ID      string
	Title   string
	Message string
	Tags    map[string]string
}

func NewMessage(title string, message string, tags map[string]string) *Message {
	return &Message{
		ID:      uuid.New().String(),
		Title:   title,
		Message: message,
		Tags:    tags,
	}
}

func (m *Message) AnnotateContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, dispatcherdContext.KeyMessageID, m.ID)
}

func (m *Message) String() string {
	tags := "None"
	if m.Tags != nil {
		tags = ""
		for k, v := range m.Tags {
			tags += fmt.Sprintf("%s:%s;", k, v)
		}
	}

	return fmt.Sprintf("id='%s' title='%s' message='%s' tags='%s'", m.ID, m.Title, m.Message, tags)
}
