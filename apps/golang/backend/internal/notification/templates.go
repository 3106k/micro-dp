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
