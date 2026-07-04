package mail

import (
	"bytes"
	_ "embed"
	"html/template"
	texttemplate "text/template"
)

//go:embed templates/email_verification.txt.tmpl
var emailVerificationTextTemplate string

//go:embed templates/email_verification.html.tmpl
var emailVerificationHTMLTemplate string

type EmailVerification struct {
	Username        string
	Email           string
	VerificationURL string
}

func (m EmailVerification) Envelope() Envelope {
	return Envelope{
		To:      []Address{{Name: m.Username, Address: m.Email}},
		Subject: "Verify your email",
	}
}

func (m EmailVerification) Content() (Content, error) {
	data := struct {
		Username        string
		VerificationURL string
	}{Username: m.Username, VerificationURL: m.VerificationURL}
	text, err := renderVerificationTextTemplate("email-verification.txt", emailVerificationTextTemplate, data)
	if err != nil {
		return Content{}, err
	}
	html, err := renderVerificationHTMLTemplate("email-verification.html", emailVerificationHTMLTemplate, data)
	if err != nil {
		return Content{}, err
	}
	return Content{Text: text, HTML: html}, nil
}

func (EmailVerification) Attachments() []Attachment { return nil }

func renderVerificationTextTemplate(name, source string, data any) (string, error) {
	tmpl, err := texttemplate.New(name).Parse(source)
	if err != nil {
		return "", err
	}
	var output bytes.Buffer
	if err := tmpl.Execute(&output, data); err != nil {
		return "", err
	}
	return output.String(), nil
}

func renderVerificationHTMLTemplate(name, source string, data any) (string, error) {
	tmpl, err := template.New(name).Parse(source)
	if err != nil {
		return "", err
	}
	var output bytes.Buffer
	if err := tmpl.Execute(&output, data); err != nil {
		return "", err
	}
	return output.String(), nil
}
