package mailer

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"github.com/wneessen/go-mail"
)

const (
	smtpMailtrapHost = "sandbox.smtp.mailtrap.io"
	smtpMailtrapPort = 587
)

var (
	//go:embed "templates"
	templateFS embed.FS
)

type EmailSender interface {
	SendEmail(
		subject string,
		data any,
		recipients []string,
		cc []string,
		bcc []string,
		attachFiles []string,
		htmlTemplateFile string,
	) error
}

type MailtrapSender struct {
	username         string
	password         string
	fromEmailName    string
	fromEmailAddress string
}

func NewMailtrapSender(username, password, fromEmailName, fromEmailAddress string) EmailSender {
	return &MailtrapSender{
		username,
		password,
		fromEmailName,
		fromEmailAddress,
	}
}

func (sender *MailtrapSender) SendEmail(
	subject string,
	data any,
	recipients []string,
	cc []string,
	bcc []string,
	attachFiles []string,
	htmlTemplateFile string,
) error {
	m := mail.NewMsg()

	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+htmlTemplateFile)
	if err != nil {
		return fmt.Errorf("failed to parse template files %w", err)
	}

	err = m.FromFormat(sender.fromEmailName, sender.fromEmailAddress)
	if err != nil {
		return fmt.Errorf("failed to set From address %w", err)
	}

	m.Subject(subject)

	if err = m.To(recipients...); err != nil {
		return fmt.Errorf("failed to set To address: %w", err)
	}

	if err = m.Cc(cc...); err != nil {
		return fmt.Errorf("failed to set Cc addresses: %w", err)
	}

	if err = m.Bcc(bcc...); err != nil {
		return fmt.Errorf("failed to set Bcc addresses: %w", err)
	}

	for _, v := range attachFiles {
		m.AttachFile(v)
	}

	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return fmt.Errorf("failed to read plainBody %w", err)
	}

	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return fmt.Errorf("failed to read htmlBody %w", err)
	}

	m.SetBodyString(mail.TypeTextHTML, htmlBody.String())
	m.AddAlternativeString(mail.TypeTextPlain, plainBody.String())

	c, err := mail.NewClient(smtpMailtrapHost, mail.WithPort(smtpMailtrapPort), mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(sender.username), mail.WithPassword(sender.password))
	if err != nil {
		return fmt.Errorf("failed to create mail client: %w", err)
	}

	if err := c.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send mail: %w", err)
	}

	return nil
}
