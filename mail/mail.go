package mail

import (
	"bytes"
	"errors"
	html "html/template"
	"io"
	text "text/template"
	"time"

	"github.com/asciimoo/omnom/config"

	smtp "github.com/xhit/go-simple-mail/v2"
)

type Templates struct {
	Text *text.Template
	HTML *html.Template
}

var client *smtp.SMTPClient
var sender string = "Omnom <omnom@127.0.0.1>"
var disabled bool = false
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
	server := smtp.NewSMTPClient()

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
	server.KeepAlive = false

	var err error
	client, err = server.Connect()
	if err != nil {
		return err
	}
	return nil
}

func Send(to string, msgType string, args map[string]interface{}) error {
	if disabled {
		return nil
	}
	email := smtp.NewMSG()
	email.SetFrom(sender).
		AddTo(to).
		SetSubject("todo")

	h, err := templates.RenderHTML(msgType, args)
	if err != nil {
		return err
	}
	email.SetBody(smtp.TextHTML, h)

	t, err := templates.RenderText(msgType, args)
	if err != nil {
		return err
	}
	email.SetBody(smtp.TextPlain, t)

	if email.Error != nil {
		return email.Error
	}
	return email.Send(client)
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
		return "", errors.New("Mail template not found")
	}
	b := bytes.NewBuffer(nil)
	m.Execute(b, args)
	s, err := b.ReadString(0)
	if err != nil && err != io.EOF {
		return "", err
	}
	return s, nil
}

func (t *Templates) RenderText(tname string, args map[string]interface{}) (string, error) {
	m := templates.Text.Lookup(tname + ".txt.tpl")
	if m == nil {
		return "", errors.New("Mail template not found")
	}
	b := bytes.NewBuffer(nil)
	m.Execute(b, args)
	s, err := b.ReadString(0)
	if err != nil && err != io.EOF {
		return "", err
	}
	return s, nil
}
