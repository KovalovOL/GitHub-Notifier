package notifier

import (
	"bytes"
	"context"
	"embed"
	"html/template"
	"log"

	"gopkg.in/gomail.v2"
)

type Notifier interface {
	SendVerificationCode(ctx context.Context, email string, token string) error
	SendRepoUpdate(ctx context.Context, email string, repoName string, newTag string) error
	SendSubscriptionSuccess(ctx context.Context, email string, repoName string) error
}

//go:embed templates/*.html
var templateFS embed.FS

type client struct {
	host      string
	port      int
	user      string
	pass      string
	fromEmail string
	templates *template.Template
}

func NewEmailNotifier(host string, port int, user, pass, fromEmail string) Notifier {
	tmpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	return &client{
		host:      host,
		port:      port,
		user:      user,
		pass:      pass,
		fromEmail: fromEmail,
		templates: tmpl,
	}
}

func (c *client) SendVerificationCode(ctx context.Context, email string, token string) error {
	data := struct {
		Token string
	}{
		Token: token,
	}

	var body bytes.Buffer
	if err := c.templates.ExecuteTemplate(&body, "confirmation.html", data); err != nil {
		return err
	}

	m := gomail.NewMessage()
	m.SetHeader("From", c.fromEmail)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Verify your email address")
	m.SetBody("text/html", body.String())

	d := gomail.NewDialer(c.host, c.port, c.user, c.pass)
	return d.DialAndSend(m)
}

func (c *client) SendRepoUpdate(ctx context.Context, email string, repoName string, newTag string) error {
	data := struct {
		RepoName string
		NewTag   string
	}{
		RepoName: repoName,
		NewTag:   newTag,
	}

	var body bytes.Buffer
	if err := c.templates.ExecuteTemplate(&body, "updateNotification.html", data); err != nil {
		return err
	}

	m := gomail.NewMessage()
	m.SetHeader("From", c.fromEmail)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Repository Update")
	m.SetBody("text/html", body.String())

	d := gomail.NewDialer(c.host, c.port, c.user, c.pass)
	return d.DialAndSend(m)
}

func (c *client) SendSubscriptionSuccess(ctx context.Context, email string, repoName string) error {
	data := struct {
		RepoName string
	}{
		RepoName: repoName,
	}

	var body bytes.Buffer
	if err := c.templates.ExecuteTemplate(&body, "confirmSubscription.html", data); err != nil {
		return err
	}

	m := gomail.NewMessage()
	m.SetHeader("From", c.fromEmail)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "New Subscription")
	m.SetBody("text/html", body.String())

	d := gomail.NewDialer(c.host, c.port, c.user, c.pass)
	return d.DialAndSend(m)
}
