package userapplication

import "context"

// EmailSender is the port for sending transactional emails.
// The infrastructure layer provides the concrete adapter (e.g. Brevo).
type EmailSender interface {
	SendVerificationEmail(ctx context.Context, toEmail, toName, verifyURL string) error
}
