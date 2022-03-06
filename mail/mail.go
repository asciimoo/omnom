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

	smtp "github.com/xhit/go-simple-mail/v2"
)

var (
	ErrTemplateNotFound = errors.New("mail template not found")
)

type Templates struct {
	Text *text.Template
	HTML *html.Template
}

var client *smtp.SMTPClient
var server *smtp.SMTPServer
var sender = "Omnom <omnom@127.0.0.1>"
var disabled = false
var templates = &Templates{}

func init() {
	var err error
	templates.HTML, err = html.New("mail").ParseGlob("templates/mail/*html.tpl")
	if err != nil {
		panic(err)
	}
	templates.Text, err = text.New("mail").ParseGlob("templates/mail/*txt.tpl")
	if err != nil {
		panic(err)
	}
}

func Init(c *config.Config) error {
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

	var err error
	client, err = server.Connect()
	if err != nil {
		return err
	}
	return nil
}

func Send(to string, subject string, msgType string, args map[string]interface{}) error {
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

func Disable(t bool) {
	disabled = t
}

func SetSender(s string) {
	sender = s
}

func (t *Templates) RenderHTML(tname string, args map[string]interface{}) (string, error) {
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

func (t *Templates) RenderText(tname string, args map[string]interface{}) (string, error) {
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
