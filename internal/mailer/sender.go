package mailer

import (
	"embed"
)

const (
	senderEmailName    = "Greenlight"
	senderEmailAddress = "noreply@mail.accounts.greenlight.com"
)

var (
	//go:embed "templates"
	templateFS embed.FS
)

type EmailSender interface {
	SendEmail(
		header EmailHeader,
		data any,
		htmlTemplateFile string,
	) error
}

type EmailHeader struct {
	Subject       string
	To            []string
	Cc            []string
	Bcc           []string
	AttachedFiles []string
}
