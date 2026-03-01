package notification

import (
	"bytes"
	"embed"
	"html/template"
	"regexp"
	"strings"
)

//go:embed templates/*.html
var templateFS embed.FS

var tmpl = template.Must(template.ParseFS(templateFS, "templates/*.html"))

var htmlTagRe = regexp.MustCompile(`<[^>]*>`)

type welcomeData struct {
	DisplayName string
}

// RenderWelcome renders the welcome email template.
func RenderWelcome(displayName string) (subject, html, text string, err error) {
	var buf bytes.Buffer
	if err = tmpl.ExecuteTemplate(&buf, "welcome.html", welcomeData{DisplayName: displayName}); err != nil {
		return "", "", "", err
	}
	html = buf.String()
	text = htmlToPlainText(html)
	subject = "Welcome to micro-dp!"
	return subject, html, text, nil
}

type invitationData struct {
	InviterName string
	TenantName  string
	Role        string
	Token       string
	AcceptURL   string
}

// RenderInvitation renders the invitation email template.
func RenderInvitation(inviterName, tenantName, role, token, acceptURL string) (subject, html, text string, err error) {
	var buf bytes.Buffer
	data := invitationData{
		InviterName: inviterName,
		TenantName:  tenantName,
		Role:        role,
		Token:       token,
		AcceptURL:   acceptURL,
	}
	if err = tmpl.ExecuteTemplate(&buf, "invitation.html", data); err != nil {
		return "", "", "", err
	}
	html = buf.String()
	text = htmlToPlainText(html)
	subject = "You're invited to join " + tenantName + " on micro-dp"
	return subject, html, text, nil
}

func htmlToPlainText(s string) string {
	s = strings.ReplaceAll(s, "<br>", "\n")
	s = strings.ReplaceAll(s, "<br/>", "\n")
	s = strings.ReplaceAll(s, "<br />", "\n")
	s = strings.ReplaceAll(s, "</p>", "\n\n")
	s = strings.ReplaceAll(s, "</h1>", "\n")
	s = strings.ReplaceAll(s, "</h2>", "\n")
	s = htmlTagRe.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	return s
}
