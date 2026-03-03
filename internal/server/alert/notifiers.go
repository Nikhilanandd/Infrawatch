package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/nikhilanandd/infrawatch/internal/config"
	"github.com/nikhilanandd/infrawatch/pkg/models"
)

// --- Email Notifier ---

// EmailNotifier sends alert emails via SMTP.
type EmailNotifier struct {
	host     string
	port     int
	username string
	password string
	from     string
	to       []string
}

// NewEmailNotifier creates a new email notifier.
func NewEmailNotifier(cfg *config.ServerConfig) *EmailNotifier {
	return &EmailNotifier{
		host:     cfg.Channels.Email.SMTPHost,
		port:     cfg.Channels.Email.SMTPPort,
		username: cfg.Channels.Email.Username,
		password: cfg.Channels.Email.Password,
		from:     cfg.Channels.Email.From,
		to:       cfg.Channels.Email.To,
	}
}

func (e *EmailNotifier) Name() string { return "email" }

func (e *EmailNotifier) Send(alert models.Alert) error {
	subject := fmt.Sprintf("[InfraWatch %s] %s", strings.ToUpper(alert.Severity), alert.Message)
	body := fmt.Sprintf(
		"Alert: %s\nNode: %s\nSeverity: %s\nValue: %.2f\nFired At: %s\n",
		alert.Message, alert.NodeID, alert.Severity, alert.Value, alert.FiredAt.Format(time.RFC3339),
	)

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		e.from, strings.Join(e.to, ","), subject, body,
	)

	auth := smtp.PlainAuth("", e.username, e.password, e.host)
	addr := fmt.Sprintf("%s:%d", e.host, e.port)
	return smtp.SendMail(addr, auth, e.from, e.to, []byte(msg))
}

// --- Slack Notifier ---

// SlackNotifier sends alerts via Slack webhook.
type SlackNotifier struct {
	webhookURL string
	channel    string
	client     *http.Client
}

// NewSlackNotifier creates a new Slack notifier.
func NewSlackNotifier(cfg *config.ServerConfig) *SlackNotifier {
	return &SlackNotifier{
		webhookURL: cfg.Channels.Slack.WebhookURL,
		channel:    cfg.Channels.Slack.Channel,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *SlackNotifier) Name() string { return "slack" }

func (s *SlackNotifier) Send(alert models.Alert) error {
	icon := ":warning:"
	if alert.Severity == "critical" {
		icon = ":rotating_light:"
	}

	payload := map[string]interface{}{
		"channel": s.channel,
		"text":    fmt.Sprintf("%s *InfraWatch Alert*", icon),
		"attachments": []map[string]interface{}{
			{
				"color": severityColor(alert.Severity),
				"fields": []map[string]string{
					{"title": "Message", "value": alert.Message, "short": "false"},
					{"title": "Node", "value": alert.NodeID, "short": "true"},
					{"title": "Severity", "value": alert.Severity, "short": "true"},
					{"title": "Value", "value": fmt.Sprintf("%.2f", alert.Value), "short": "true"},
					{"title": "Time", "value": alert.FiredAt.Format(time.RFC3339), "short": "true"},
				},
			},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal slack payload: %w", err)
	}

	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("slack webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("slack returned status %d", resp.StatusCode)
	}
	return nil
}

func severityColor(severity string) string {
	switch severity {
	case "critical":
		return "#FF0000"
	case "warning":
		return "#FFA500"
	default:
		return "#36A64F"
	}
}
