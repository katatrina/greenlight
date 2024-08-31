package mailer

import (
	"bytes"
	"fmt"
	"html/template"

	mail "github.com/wneessen/go-mail"
)

const (
	smtpMailtrapHost = "sandbox.smtp.mailtrap.io"
	smtpMailtrapPort = 587
)

type MailtrapSender struct {
	client *mail.Client
}

func NewMailtrapSender(username, password string) (EmailSender, error) {
	client, err := mail.NewClient(smtpMailtrapHost, mail.WithPort(smtpMailtrapPort), mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(username), mail.WithPassword(password))
	if err != nil {
		return nil, err
	}

	return &MailtrapSender{
		client: client,
	}, nil
}

func (sender *MailtrapSender) SendEmail(
	header EmailHeader,
	data any,
	htmlTemplateFile string,
) error {
	// Initialize a new email message
	m := mail.NewMsg()

	// Try to parse the email template
	tmpl, err := template.ParseFS(templateFS, "templates/"+htmlTemplateFile)
	if err != nil {
		return fmt.Errorf("failed to parse template files %w", err)
	}

	// Prepare email header fields

	// Set "From: Greenlight <noreply@mail.accounts.greenlight.com>"
	err = m.FromFormat(senderEmailName, senderEmailAddress)
	if err != nil {
		return fmt.Errorf("failed to set From address %w", err)
	}

	// Set the subject title
	m.Subject(header.Subject)

	// Set the recipient email addresses
	if err = m.To(header.To...); err != nil {
		return fmt.Errorf("failed to set To address: %w", err)
	}

	// Set the Cc email addresses
	if err = m.Cc(header.Cc...); err != nil {
		return fmt.Errorf("failed to set Cc addresses: %w", err)
	}

	// Set the Bcc email addresses
	if err = m.Bcc(header.Bcc...); err != nil {
		return fmt.Errorf("failed to set Bcc addresses: %w", err)
	}

	// Prepare email body

	// Attach files
	for _, v := range header.AttachedFiles {
		m.AttachFile(v)
	}

	// Try to execute the plain body template
	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return fmt.Errorf("failed to read plainBody %w", err)
	}

	// Try to execute the html body template
	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return fmt.Errorf("failed to read htmlBody %w", err)
	}

	// Set the primary message body (text/html)
	m.SetBodyString(mail.TypeTextHTML, htmlBody.String())
	// Add an alternative message body (text/plain) if the HTML body is not supported by the client
	m.AddAlternativeString(mail.TypeTextPlain, plainBody.String())

	// Send email
	if err := sender.client.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
