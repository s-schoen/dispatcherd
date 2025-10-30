package dispatch

import (
	"context"
	"dispatcherd/logging"
	"fmt"
	"log/slog"
	"math"

	"github.com/wneessen/go-mail"
)

type mailConfig struct {
	to         string
	smtpServer string
	username   string
	password   string
	tls        bool
	port       int
}

type MailDispatcher struct {
	logger *slog.Logger
	config mailConfig
}

func (m *MailDispatcher) Dispatch(ctx context.Context, msg *Message) error {
	m.logger.DebugContext(ctx, fmt.Sprintf("sending mail to=%s, from=%s, via=%s:%d, tls=%t",
		m.config.to, m.config.username, m.config.smtpServer, m.config.port, m.config.tls))

	message := mail.NewMsg()
	if err := message.From(m.config.username); err != nil {
		return err
	}

	if err := message.To(m.config.to); err != nil {
		return err
	}

	message.Subject(msg.Title)
	message.SetBodyString(mail.TypeTextPlain, msg.Message)

	options := []mail.Option{
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(m.config.username),
		mail.WithPassword(m.config.password),
		mail.WithPort(m.config.port),
	}

	if m.config.tls {
		options = append(options, mail.WithSSL())
	}

	client, err := mail.NewClient(m.config.smtpServer, options...)

	if err != nil {
		return err
	}

	if err := client.DialAndSendWithContext(ctx, message); err != nil {
		return err
	}

	m.logger.DebugContext(ctx, "sent mail to "+m.config.to)

	return nil
}

func (m *MailDispatcher) ConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"to":         "required,email",
		"smtpServer": "required,hostname",
		"smtpPort":   "required",
		"username":   "required,email",
		"password":   "required",
		"tls":        "required,boolean",
	}
}

func (m *MailDispatcher) SetConfig(config map[string]interface{}) {
	m.config = mailConfig{
		to:         config["to"].(string),
		smtpServer: config["smtpServer"].(string),
		port:       int(math.Round(config["smtpPort"].(float64))),
		username:   config["username"].(string),
		password:   config["password"].(string),
		tls:        config["tls"].(bool),
	}
}

func NewMailDispatcher() *MailDispatcher {
	return &MailDispatcher{
		logger: logging.GetLogger(logging.MessageProcessing),
	}
}
