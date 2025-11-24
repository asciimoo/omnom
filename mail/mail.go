// Package mail provides email sending functionality for Omnom notifications.
//
// This package handles SMTP-based email delivery for various application events:
//   - User login tokens (passwordless authentication)
//   - Feed update notifications
//   - System alerts
//
// It supports both HTML and plain text email formats using template-based rendering.
// Templates are embedded in the application and loaded from the templates/mail directory.
//
// The package maintains a persistent SMTP connection that is reused across multiple
// sends for efficiency. If the connection is lost, it automatically reconnects.
// Email sending can be disabled entirely by leaving the SMTP host empty in configuration.
//
// Configuration supports:
//   - Standard SMTP authentication
//   - TLS/SSL encryption
//   - Configurable timeouts
//   - Custom sender addresses
//
// Example usage:
//
//	err := mail.Init(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	err = mail.Send(
//	    "user@example.com",
//	    "Welcome to Omnom",
//	    "welcome",
//	    map[string]any{"username": "john"},
//	)
package mail

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"syscall"
	"time"

	html "html/template"
	text "text/template"

	"github.com/asciimoo/omnom/config"
	templatesfs "github.com/asciimoo/omnom/templates"

	smtp "github.com/xhit/go-simple-mail/v2"
)

var (
	// ErrTemplateNotFound is returned when a mail template cannot be found.
	ErrTemplateNotFound = errors.New("mail template not found")
)

// Templates holds HTML and text email templates.
type Templates struct {
	Text *text.Template
	HTML *html.Template
}

var client *smtp.SMTPClient
var server *smtp.SMTPServer
var sender = "Omnom <omnom@127.0.0.1>"
var disabled = false
var templates = &Templates{}

// Init initializes the mail client and loads templates.
func Init(c *config.Config) error {
	var err error
	templates.HTML, err = html.New("mail").ParseFS(templatesfs.FS, "mail/*html.tpl")
	if err != nil {
		return errors.New("failed to parse mail html templates")
	}
	templates.Text, err = text.New("mail").ParseFS(templatesfs.FS, "mail/*txt.tpl")
	if err != nil {
		return errors.New("failed to parse mail text templates")
	}

	sc := c.SMTP

	if sc.Host == "" {
		Disable(true)
		return nil
	}

	SetSender(sc.Sender)
	server = smtp.NewSMTPClient()

	//SMTP server
	server.Host = sc.Host
	server.Port = sc.Port
	server.Username = sc.Username
	server.Password = sc.Password
	if sc.TLS {
		server.Encryption = smtp.EncryptionTLS
	} else {
		server.Encryption = smtp.EncryptionNone
	}
	server.ConnectTimeout = time.Duration(sc.ConnectionTimeout) * time.Second
	server.SendTimeout = time.Duration(sc.SendTimeout) * time.Second
	server.KeepAlive = true

	client, err = server.Connect()
	if err != nil {
		return err
	}
	return nil
}

// Send sends an email using the specified template.
func Send(to string, subject string, msgType string, args map[string]any) error {
	if disabled {
		return nil
	}
	email := smtp.NewMSG()
	email.SetFrom(sender).
		AddTo(to).
		SetSubject(subject)

	h, err := templates.RenderHTML(msgType, args)
	if err != nil {
		return err
	}

	t, err := templates.RenderText(msgType, args)
	if err != nil {
		return err
	}
	email.SetBody(smtp.TextPlain, t).AddAlternative(smtp.TextHTML, h)

	if email.GetError() != nil {
		return email.GetError()
	}
	err = email.Send(client)
	if errors.Is(err, io.EOF) || errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) {
		client, err = server.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect to mail server: %w", err)
		}
		return email.Send(client)
	}
	return err
}

// Disable enables or disables mail sending.
func Disable(t bool) {
	disabled = t
}

// SetSender sets the sender email address.
func SetSender(s string) {
	sender = s
}

// RenderHTML renders html template with given arguments.
func (t *Templates) RenderHTML(tname string, args map[string]any) (string, error) {
	m := templates.HTML.Lookup(tname + ".html.tpl")
	if m == nil {
		return "", ErrTemplateNotFound
	}
	b := bytes.NewBuffer(nil)
	err := m.Execute(b, args)
	if err != nil {
		return "", fmt.Errorf("failed to execute template '%s': %w", tname, err)
	}
	s, err := b.ReadString(0)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return s, nil
}

// RenderHTML renders text template with given arguments.
func (t *Templates) RenderText(tname string, args map[string]any) (string, error) {
	m := templates.Text.Lookup(tname + ".txt.tpl")
	if m == nil {
		return "", ErrTemplateNotFound
	}
	b := bytes.NewBuffer(nil)
	err := m.Execute(b, args)
	if err != nil {
		return "", fmt.Errorf("failed to execute template '%s': %w", tname, err)
	}
	s, err := b.ReadString(0)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return s, nil
}
