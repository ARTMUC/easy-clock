package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const brevoSendURL = "https://api.brevo.com/v3/smtp/email"

// BrevoClient sends transactional emails via the Brevo API.
type BrevoClient struct {
	apiKey      string
	senderEmail string
	senderName  string
	httpClient  *http.Client
}

// NewBrevoClient creates a new Brevo email client.
func NewBrevoClient(apiKey, senderEmail, senderName string) *BrevoClient {
	return &BrevoClient{
		apiKey:      apiKey,
		senderEmail: senderEmail,
		senderName:  senderName,
		httpClient:  &http.Client{},
	}
}

type brevoContact struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type brevoPayload struct {
	Sender      brevoContact   `json:"sender"`
	To          []brevoContact `json:"to"`
	Subject     string         `json:"subject"`
	HTMLContent string         `json:"htmlContent"`
}

// SendVerificationEmail sends an account activation link to the user.
func (c *BrevoClient) SendVerificationEmail(ctx context.Context, toEmail, toName, verifyURL string) error {
	body := brevoPayload{
		Sender: brevoContact{Email: c.senderEmail, Name: c.senderName},
		To:     []brevoContact{{Email: toEmail, Name: toName}},
		Subject: "Activate your account",
		HTMLContent: fmt.Sprintf(`
<div style="font-family:sans-serif;max-width:480px;margin:0 auto">
  <h2 style="color:#6366f1">Starter</h2>
  <p>Hi <strong>%s</strong>,</p>
  <p>Click the button below to activate your account:</p>
  <p style="text-align:center;margin:32px 0">
    <a href="%s"
       style="background:#6366f1;color:#fff;padding:14px 28px;border-radius:12px;text-decoration:none;font-weight:600">
      Activate account
    </a>
  </p>
  <p style="color:#9ca3af;font-size:13px">
    This link expires after use. If you did not create an account, you can safely ignore this email.
  </p>
</div>`, toName, verifyURL),
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("brevo: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, brevoSendURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("brevo: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("brevo: send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("brevo: status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
