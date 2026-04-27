package email

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"net"
	"net/smtp"
	"net/url"
)

//go:embed templates/*.html
var templateFS embed.FS

// SMTPSender implements domain.EmailSender via SMTP.
type SMTPSender struct {
	host          string
	port          int
	appURL        string
	appDeepScheme string
	from          string
	tmpl          *template.Template
}

// NewSMTPSender creates a new email sender.
func NewSMTPSender(host string, port int, from string, appURL string, appDeepScheme string) (*SMTPSender, error) {
	tmpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("parsing email templates: %w", err)
	}

	return &SMTPSender{
		host:          host,
		port:          port,
		appURL:        appURL,
		appDeepScheme: appDeepScheme,
		from:          from,
		tmpl:          tmpl,
	}, nil
}

// SendMagicLink sends a magic link email with both web and app deep links.
func (s *SMTPSender) SendMagicLink(ctx context.Context, to string, token string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("sending email: %w", err)
	}

	var body bytes.Buffer
	if err := s.tmpl.ExecuteTemplate(&body, "magic_link.html", s.magicLinkData(token)); err != nil {
		return fmt.Errorf("rendering magic link template: %w", err)
	}

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: Sign in to Initium\r\nContent-Type: text/html; charset=utf-8\r\n\r\n%s",
		s.from, to, body.String(),
	)

	if err := s.sendMail(ctx, to, []byte(msg)); err != nil {
		return fmt.Errorf("sending email: %w", err)
	}

	slog.Info("magic link email sent")
	return nil
}

func (s *SMTPSender) magicLinkData(token string) map[string]string {
	escapedToken := url.QueryEscape(token)
	data := map[string]string{
		"Link": fmt.Sprintf("%s/api/auth/verify?token=%s", s.appURL, escapedToken),
	}
	if s.appDeepScheme != "" {
		data["AppLink"] = fmt.Sprintf("%s://auth/verify?token=%s", s.appDeepScheme, escapedToken)
	}
	return data
}

func (s *SMTPSender) sendMail(ctx context.Context, to string, msg []byte) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		if err := conn.SetDeadline(deadline); err != nil {
			return err
		}
	}

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.Mail(s.from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		_ = w.Close()
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return client.Quit()
}
