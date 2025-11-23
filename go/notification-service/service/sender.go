package service

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type MailGunMailer struct {
	apiKey string
	domain string
	client *http.Client
}
type EmailRequest struct {
	To      string   `json:"to"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
	Tags    []string `json:"tags"`
}
type Notifier interface {
	SendEmail(ctx context.Context, req EmailRequest) error
}

func NewMailGunMailer(apiKey, domain string) *MailGunMailer {
	return &MailGunMailer{
		apiKey: apiKey,
		domain: domain,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}
func (m *MailGunMailer) SendEmail(ctx context.Context, req EmailRequest) error {
	form := url.Values{}
	form.Set("from", fmt.Sprintf("Notification Service <%s>", "no-reply@"+m.domain))
	form.Set("to", req.To)
	form.Set("subject", req.Subject)
	form.Set("text", req.Body)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("https://api.mailgun.net/v3/%s/messages", m.domain),
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return err
	}

	httpReq.SetBasicAuth("api", m.apiKey)
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := m.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed to send email: status=%d", resp.StatusCode)
	}

	return nil
}
