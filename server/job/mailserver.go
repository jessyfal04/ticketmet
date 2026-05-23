package job

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/smtp"
	"strconv"
	"strings"
)

// Message contains the user-visible mail content.
type Message struct {
	Subject string
	Text    string
	HTML    string
}

// Envelope is the item sent through the mail channel.
type Envelope struct {
	Dst     string
	Message Message
}

// AlertMailItem represents a single alert in the alert email.
type AlertMailItem struct {
	Title   string
	Details string
	URL     string
}

type mailPage struct {
	Title   string
	Email   string
	Content template.HTML
}

// Config contains SMTP connection settings.
type Config struct {
	Host string
	Port int
	From string
}

// MailServer consumes mail envelopes from C and sends them through SMTP.
type MailServer struct {
	Config Config
	C      chan Envelope
}

func RunMailServer(ctx context.Context, config Config, c chan Envelope) {
	go (&MailServer{Config: config, C: c}).Run(ctx)
}

// Start the mail server
func (s *MailServer) Run(ctx context.Context) {
	if s.Config.Host == "" {
		log.Printf("mailserver disabled: SMTP_HOST is empty")
	}

	runChan(ctx, s.C, func(env Envelope) {
		if s.Config.Host == "" {
			log.Printf("mailserver disabled: dropped mail to %s", env.Dst)
			return
		}
		if err := s.send(env); err != nil {
			log.Printf("mailserver send to %s failed: %v", env.Dst, err)
			return
		}
		log.Printf("mailserver sent to %s subject=%q", env.Dst, env.Message.Subject)
	})
}

// Send an email
func (s *MailServer) send(env Envelope) error {
	to := strings.TrimSpace(env.Dst)
	if to == "" {
		return fmt.Errorf("empty destination")
	}
	if env.Message.Subject == "" {
		return fmt.Errorf("empty subject")
	}

	addr := net.JoinHostPort(s.Config.Host, strconv.Itoa(s.Config.Port))
	client, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.Hello("ticketmet"); err != nil {
		return err
	}

	if err := client.Mail(s.Config.From); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}
	defer writer.Close()

	return writeMessage(writer, s.Config.From, to, env.Message)
}

// Write the email message in MIME multipart format to the given writer
func writeMessage(w io.Writer, from string, to string, msg Message) error {
	mw := multipart.NewWriter(w)
	boundary := mw.Boundary()
	headers := []string{
		"From: " + from,
		"To: " + to,
		"Subject: " + encodeHeader(msg.Subject),
		"MIME-Version: 1.0",
		`Content-Type: multipart/alternative; boundary="` + boundary + `"`,
		"",
	}
	if _, err := io.WriteString(w, strings.Join(headers, "\r\n")+"\r\n"); err != nil {
		return err
	}

	text := msg.Text
	part, err := mw.CreatePart(map[string][]string{
		"Content-Type":              {`text/plain; charset="UTF-8"`},
		"Content-Transfer-Encoding": {"8bit"},
	})
	if err != nil {
		return err
	}
	if _, err := io.WriteString(part, text); err != nil {
		return err
	}

	part, err = mw.CreatePart(map[string][]string{
		"Content-Type":              {`text/html; charset="UTF-8"`},
		"Content-Transfer-Encoding": {"8bit"},
	})
	if err != nil {
		return err
	}
	if _, err := io.WriteString(part, msg.HTML); err != nil {
		return err
	}

	return mw.Close()
}

// Constructor of the welcome email message, takes the user email as parameter
func WelcomeMail(email string) Message {
	baseURL := appBaseURL()
	loginURL := baseURL + "/"
	content := template.HTML(fmt.Sprintf(`<p>Your Ticketmet account is ready.</p><p>You can now sign in with <strong>%s</strong>, follow your favorite concerts, and manage alerts.</p>`, template.HTMLEscapeString(email)))
	page := mailPage{
		Title:   "Welcome to Ticketmet",
		Email:   email,
		Content: content,
	}
	return Message{
		Subject: "Welcome to Ticketmet",
		Text:    fmt.Sprintf("Welcome to Ticketmet\n\nYour account %s is ready.\n\nSign in: %s\n", email, loginURL),
		HTML:    renderMail(page),
	}
}

// Constructor of the alert email message, takes the user email and a list of AlertMailItem
func AlertMail(email string, items []AlertMailItem) Message {
	subject := fmt.Sprintf("Ticketmet: %d new alerts", len(items))
	if len(items) == 1 {
		subject = items[0].Title
	}

	var text strings.Builder
	text.WriteString("Ticketmet alerts\n\n")
	var content strings.Builder
	content.WriteString("<p>Your Ticketmet radar found new matches.</p><div style=\"margin:18px 0 0;\">")
	for _, item := range items {
		text.WriteString("- ")
		text.WriteString(item.Title)
		if item.Details != "" {
			text.WriteString(" — ")
			text.WriteString(item.Details)
		}
		if item.URL != "" {
			text.WriteString("\n  ")
			text.WriteString(item.URL)
		}
		text.WriteString("\n")

		content.WriteString(`<div style="margin:0 0 14px;padding:14px 16px;border:1px solid #e5e7eb;border-radius:16px;background:#f9fafb;"><div style="font-weight:700;color:#111827;">`)
		content.WriteString(template.HTMLEscapeString(item.Title))
		content.WriteString(`</div>`)
		if item.Details != "" {
			content.WriteString(`<div style="margin-top:6px;color:#4b5563;">`)
			content.WriteString(template.HTMLEscapeString(item.Details))
			content.WriteString(`</div>`)
		}
		if item.URL != "" {
			content.WriteString(`<div style="margin-top:10px;"><a href="`)
			content.WriteString(template.HTMLEscapeString(item.URL))
			content.WriteString(`" style="color:#7c3aed;font-weight:700;text-decoration:none;">View concert</a></div>`)
		}
		content.WriteString(`</div>`)
	}
	content.WriteString("</div>")

	page := mailPage{
		Title:   subject,
		Email:   email,
		Content: template.HTML(content.String()),
	}
	return Message{
		Subject: subject,
		Text:    text.String(),
		HTML:    renderMail(page),
	}
}

func renderMail(page mailPage) string {
	var builder strings.Builder
	buttonURL := appBaseURL()
	tmpl := template.Must(template.New("mail").Parse(`<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>{{.Title}}</title></head><body style="margin:0;background:#f5f7fb;font-family:Arial,Helvetica,sans-serif;color:#111827;"><table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="background:#f5f7fb;padding:32px 12px;"><tr><td align="center"><table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="max-width:620px;background:white;border-radius:24px;overflow:hidden;box-shadow:0 20px 60px rgba(15,23,42,.12);"><tr><td style="padding:34px 32px;background:linear-gradient(135deg,#7c3aed,#ec4899);color:white;"><div style="font-size:13px;letter-spacing:.12em;text-transform:uppercase;opacity:.85;">Ticketmet</div><h1 style="margin:10px 0 0;font-size:30px;line-height:1.15;">{{.Title}}</h1></td></tr><tr><td style="padding:30px 32px 8px;"><div style="padding:16px 18px;background:#f9fafb;border:1px solid #e5e7eb;border-radius:16px;"><div style="font-size:12px;text-transform:uppercase;letter-spacing:.08em;color:#6b7280;">Account</div><div style="margin-top:5px;font-size:16px;font-weight:700;color:#111827;">{{.Email}}</div></div></td></tr><tr><td style="padding:18px 32px 8px;font-size:16px;line-height:1.65;color:#374151;">{{.Content}}</td></tr><tr><td style="padding:14px 32px 34px;"><a href="` + buttonURL + `" style="display:inline-block;padding:13px 18px;border-radius:999px;background:#111827;color:white;text-decoration:none;font-weight:700;">Open Ticketmet</a></td></tr><tr><td style="padding:18px 32px;background:#f9fafb;border-top:1px solid #e5e7eb;color:#6b7280;font-size:13px;line-height:1.5;">This email was sent automatically by Ticketmet.</td></tr></table></td></tr></table></body></html>`))
	if err := tmpl.Execute(&builder, page); err != nil {
		return "<p>" + template.HTMLEscapeString(page.Title) + "</p>"
	}
	return builder.String()
}

// Helpeurs

// Remove newlines in the headers
func encodeHeader(value string) string {
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "\n", "")
	return value
}

// url of the app
func appBaseURL() string {
	return strings.TrimRight(Getenv("APP_BASE_URL", "https://ticketmet.jessyfal04.dev"), "/")
}
