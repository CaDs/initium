package email

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"net/smtp"
)

//go:embed templates/*.html
var templateFS embed.FS

// SMTPSender implements domain.EmailSender via SMTP.
type SMTPSender struct {
	host    string
	port    int
	appURL  string
	from    string
	tmpl    *template.Template
}

// NewSMTPSender creates a new email sender.
func NewSMTPSender(host string, port int, appURL string) (*SMTPSender, error) {
	tmpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("parsing email templates: %w", err)
	}

	return &SMTPSender{
		host:   host,
		port:   port,
		appURL: appURL,
		from:   "noreply@initium.local",
		tmpl:   tmpl,
	}, nil
}

// SendMagicLink sends a magic link email.
func (s *SMTPSender) SendMagicLink(ctx context.Context, to string, token string) error {
	link := fmt.Sprintf("%s/api/auth/verify?token=%s", s.appURL, token)

	var body bytes.Buffer
	if err := s.tmpl.ExecuteTemplate(&body, "magic_link.html", map[string]string{
		"Link": link,
	}); err != nil {
		return fmt.Errorf("rendering magic link template: %w", err)
	}

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: Sign in to Initium\r\nContent-Type: text/html; charset=utf-8\r\n\r\n%s",
		s.from, to, body.String(),
	)

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	if err := smtp.SendMail(addr, nil, s.from, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("sending email to %s: %w", to, err)
	}

	slog.Info("magic link email sent", "to", to)
	return nil
}
