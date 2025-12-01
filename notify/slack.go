package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"cdrgenerator/config"
)

// SlackNotifier sends notifications to Slack
type SlackNotifier struct {
	config     *config.SlackConfig
	instanceID string
	logger     *slog.Logger
	client     *http.Client
}

// SlackMessage represents a Slack webhook message
type SlackMessage struct {
	Text        string            `json:"text,omitempty"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

// SlackAttachment represents a Slack message attachment
type SlackAttachment struct {
	Color      string        `json:"color,omitempty"`
	Title      string        `json:"title,omitempty"`
	Text       string        `json:"text,omitempty"`
	Fields     []SlackField  `json:"fields,omitempty"`
	Footer     string        `json:"footer,omitempty"`
	FooterIcon string        `json:"footer_icon,omitempty"`
	Timestamp  int64         `json:"ts,omitempty"`
}

// SlackField represents a field in a Slack attachment
type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// NewSlackNotifier creates a new Slack notifier
func NewSlackNotifier(cfg *config.SlackConfig, instanceID string, logger *slog.Logger) *SlackNotifier {
	return &SlackNotifier{
		config:     cfg,
		instanceID: instanceID,
		logger:     logger,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// IsEnabled returns true if Slack notifications are configured
func (s *SlackNotifier) IsEnabled() bool {
	return s.config.WebhookURL != ""
}

// NotifyStartup sends a startup notification
func (s *SlackNotifier) NotifyStartup(channels int) error {
	if !s.IsEnabled() || !s.config.NotifyStartup {
		return nil
	}

	msg := SlackMessage{
		Attachments: []SlackAttachment{
			{
				Color: "good",
				Title: "CDRGenerator Started",
				Fields: []SlackField{
					{Title: "Instance", Value: s.instanceID, Short: true},
					{Title: "Channels", Value: fmt.Sprintf("%d", channels), Short: true},
				},
				Footer:    "CDRGenerator",
				Timestamp: time.Now().Unix(),
			},
		},
	}

	return s.send(msg)
}

// NotifyShutdown sends a shutdown notification
func (s *SlackNotifier) NotifyShutdown(recordsSent int64, uptime time.Duration) error {
	if !s.IsEnabled() || !s.config.NotifyShutdown {
		return nil
	}

	msg := SlackMessage{
		Attachments: []SlackAttachment{
			{
				Color: "warning",
				Title: "CDRGenerator Stopped",
				Fields: []SlackField{
					{Title: "Instance", Value: s.instanceID, Short: true},
					{Title: "Uptime", Value: formatDuration(uptime), Short: true},
					{Title: "Records Sent", Value: fmt.Sprintf("%d", recordsSent), Short: true},
				},
				Footer:    "CDRGenerator",
				Timestamp: time.Now().Unix(),
			},
		},
	}

	return s.send(msg)
}

// NotifyError sends an error notification
func (s *SlackNotifier) NotifyError(device string, err error) error {
	if !s.IsEnabled() || !s.config.NotifyErrors {
		return nil
	}

	msg := SlackMessage{
		Attachments: []SlackAttachment{
			{
				Color: "danger",
				Title: "CDRGenerator Error",
				Fields: []SlackField{
					{Title: "Instance", Value: s.instanceID, Short: true},
					{Title: "Device", Value: device, Short: true},
					{Title: "Error", Value: err.Error(), Short: false},
				},
				Footer:    "CDRGenerator",
				Timestamp: time.Now().Unix(),
			},
		},
	}

	return s.send(msg)
}

func (s *SlackNotifier) send(msg SlackMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack message: %w", err)
	}

	req, err := http.NewRequest("POST", s.config.WebhookURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Slack message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack returned non-OK status: %d", resp.StatusCode)
	}

	s.logger.Debug("Slack notification sent")
	return nil
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
