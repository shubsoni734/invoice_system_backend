package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"mime"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/your-org/invoice-backend/internal/config"
)

const defaultAPIURL = "https://api.brevo.com/v3/smtp/email"

// Client sends transactional email through the Brevo API.
// The API key and sender address come from the environment so they can be
// added after the product is deployed.
type Client struct {
	cfg    config.EmailConfig
	apiURL string
	http   *http.Client
}

type InvoiceLine struct {
	Description string
	Quantity    string
	UnitPrice   string
	Amount      string
}

// PaymentEmailData holds the receipt fields sent after a payment is recorded.
type PaymentEmailData struct {
	BusinessName  string
	BusinessEmail string
	CustomerName  string
	InvoiceNumber string
	Currency      string
	Amount        string
	Method        string
	PaymentDate   string
	Reference     string
	BalanceDue    string
	FullyPaid     bool
}

// InvoiceEmailData holds the invoice fields rendered into the customer email.
type InvoiceEmailData struct {
	BusinessName  string
	BusinessEmail string
	InvoiceNumber string
	IssuedDate    string
	DueDate       string
	Currency      string
	CustomerName  string
	Subtotal      string
	TaxAmount     string
	GrandTotal    string
	Notes         string
	Lines         []InvoiceLine
}

type brevoAddress struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email"`
}

type brevoMessage struct {
	Sender      brevoAddress   `json:"sender"`
	To          []brevoAddress `json:"to"`
	ReplyTo     *brevoAddress  `json:"replyTo,omitempty"`
	Subject     string         `json:"subject"`
	HTMLContent string         `json:"htmlContent"`
}

func NewClient(cfg config.EmailConfig) *Client {
	apiURL := strings.TrimSpace(cfg.APIURL)
	if apiURL == "" {
		apiURL = defaultAPIURL
	}
	return &Client{
		cfg:    cfg,
		apiURL: apiURL,
		http:   &http.Client{Timeout: 20 * time.Second},
	}
}

func (c *Client) smtpConfigured() bool {
	return strings.TrimSpace(c.cfg.SMTPLogin) != "" && strings.TrimSpace(c.cfg.SMTPKey) != ""
}

// Configured reports whether a send can be attempted.
func (c *Client) Configured() bool {
	if c.smtpConfigured() {
		return true
	}
	return strings.TrimSpace(c.cfg.APIKey) != "" && strings.TrimSpace(c.cfg.SenderEmail) != ""
}

func (c *Client) fromEmail() string {
	if email := strings.TrimSpace(c.cfg.SenderEmail); email != "" {
		return email
	}
	return strings.TrimSpace(c.cfg.SMTPLogin)
}

func (c *Client) senderName() string {
	name := strings.TrimSpace(c.cfg.SenderName)
	if name == "" {
		return "Invoice Flow"
	}
	return name
}

func (c *Client) SendWelcomeEmail(toEmail, orgName, password string) error {
	body, err := render(welcomeTemplate, map[string]string{
		"OrgName":  orgName,
		"Email":    toEmail,
		"Password": password,
	})
	if err != nil {
		return err
	}
	return c.send(toEmail, "Welcome to Invoice Flow — your login details", body, "")
}

func (c *Client) SendForgotPasswordEmail(toEmail, token, frontendURL string) error {
	resetLink := strings.TrimRight(frontendURL, "/") + "/new-password?refreshtoken=" + token
	body, err := render(resetTemplate, map[string]string{
		"ResetLink": resetLink,
	})
	if err != nil {
		return err
	}
	return c.send(toEmail, "Reset your Invoice Flow password", body, "")
}

func (c *Client) SendInvoiceEmail(toEmail string, d InvoiceEmailData) error {
	subject := fmt.Sprintf("Invoice %s from %s", d.InvoiceNumber, d.BusinessName)
	body, err := render(invoiceTemplate, d)
	if err != nil {
		return err
	}
	replyTo := ""
	if strings.Contains(d.BusinessEmail, "@") {
		replyTo = d.BusinessEmail
	}
	return c.send(toEmail, subject, body, replyTo)
}

func (c *Client) SendPaymentEmail(toEmail string, d PaymentEmailData) error {
	subject := fmt.Sprintf("Payment received for %s", d.InvoiceNumber)
	body, err := render(paymentTemplate, d)
	if err != nil {
		return err
	}
	replyTo := ""
	if strings.Contains(d.BusinessEmail, "@") {
		replyTo = d.BusinessEmail
	}
	return c.send(toEmail, subject, body, replyTo)
}

// SendTestEmail confirms Brevo is configured without including a password.
func (c *Client) SendTestEmail(toEmail string) error {
	body, err := render(testTemplate, map[string]string{"Email": toEmail})
	if err != nil {
		return err
	}
	return c.send(toEmail, "Invoice Flow email test", body, "")
}

// SendActivityEmail tells an organisation admin that something changed.
func (c *Client) SendActivityEmail(toEmail, subject, message string) error {
	body, err := render(activityTemplate, map[string]string{
		"Subject": subject,
		"Message": message,
	})
	if err != nil {
		return err
	}
	return c.send(toEmail, subject, body, "")
}

func (c *Client) send(toEmail, subject, htmlBody, replyTo string) error {
	if !c.Configured() {
		return fmt.Errorf("email is not configured: set BREVO_SMTP_LOGIN and BREVO_SMTP_KEY, or BREVO_API_KEY and BREVO_SENDER_EMAIL")
	}
	if strings.TrimSpace(toEmail) == "" || !strings.Contains(toEmail, "@") {
		return fmt.Errorf("invalid recipient email")
	}
	if c.smtpConfigured() {
		return c.sendSMTP(toEmail, subject, htmlBody, replyTo)
	}

	msg := brevoMessage{
		Sender:      brevoAddress{Name: c.senderName(), Email: strings.TrimSpace(c.cfg.SenderEmail)},
		To:          []brevoAddress{{Email: toEmail}},
		Subject:     subject,
		HTMLContent: htmlBody,
	}
	if replyTo != "" {
		msg.ReplyTo = &brevoAddress{Email: replyTo, Name: c.senderName()}
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to encode email: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, c.apiURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create email request: %w", err)
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("api-key", c.cfg.APIKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("failed to reach Brevo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		limited := io.LimitReader(resp.Body, 512)
		raw, _ := io.ReadAll(limited)
		return fmt.Errorf("brevo api returned %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	return nil
}

func (c *Client) sendSMTP(toEmail, subject, htmlBody, replyTo string) error {
	host := strings.TrimSpace(c.cfg.SMTPHost)
	if host == "" {
		host = "smtp-relay.brevo.com"
	}
	port := strings.TrimSpace(c.cfg.SMTPPort)
	if port == "" {
		port = "587"
	}
	from := c.fromEmail()
	message := buildSMTPMessage(c.senderName(), from, toEmail, subject, htmlBody, replyTo)
	auth := smtp.PlainAuth("", c.cfg.SMTPLogin, c.cfg.SMTPKey, host)
	if err := smtp.SendMail(net.JoinHostPort(host, port), auth, from, []string{toEmail}, message); err != nil {
		return fmt.Errorf("brevo smtp send failed: %w", err)
	}
	return nil
}

func buildSMTPMessage(fromName, fromEmail, toEmail, subject, htmlBody, replyTo string) []byte {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "From: %s\r\n", formatAddress(fromName, fromEmail))
	fmt.Fprintf(&buf, "To: %s\r\n", toEmail)
	if replyTo != "" {
		fmt.Fprintf(&buf, "Reply-To: %s\r\n", replyTo)
	}
	fmt.Fprintf(&buf, "Subject: %s\r\n", mime.QEncoding.Encode("utf-8", subject))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(htmlBody)
	return buf.Bytes()
}

func formatAddress(name, email string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return email
	}
	return mime.QEncoding.Encode("utf-8", name) + " <" + email + ">"
}

func render(tmpl *template.Template, data any) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render email: %w", err)
	}
	return buf.String(), nil
}

var welcomeTemplate = template.Must(template.New("welcome").Parse(`<!DOCTYPE html>
<html><body style="font-family:Arial,sans-serif;color:#1f2937">
<h1>Welcome to Invoice Flow</h1>
<p>Your organisation <strong>{{.OrgName}}</strong> is ready.</p>
<p><strong>Email:</strong> {{.Email}}<br/><strong>Temporary password:</strong> {{.Password}}</p>
<p>Sign in and change this password before you share the account.</p>
</body></html>`))

var resetTemplate = template.Must(template.New("reset").Parse(`<!DOCTYPE html>
<html><body style="font-family:Arial,sans-serif;color:#1f2937">
<h1>Password reset</h1>
<p>This link is valid for 10 minutes. If you did not ask for a reset, you can ignore this email.</p>
<p><a href="{{.ResetLink}}">Reset password</a></p>
<p>{{.ResetLink}}</p>
</body></html>`))

var testTemplate = template.Must(template.New("test").Parse(`<!DOCTYPE html>
<html><body style="font-family:Arial,sans-serif;color:#1f2937">
<h1>Email is working</h1>
<p>Invoice Flow successfully sent this message to {{.Email}} through Brevo.</p>
</body></html>`))

var activityTemplate = template.Must(template.New("activity").Parse(`<!DOCTYPE html>
<html><body style="font-family:Arial,sans-serif;color:#1f2937">
<h1>{{.Subject}}</h1>
<p>{{.Message}}</p>
<p style="color:#6b7280;font-size:12px">This alert was sent because you are an administrator.</p>
</body></html>`))

var invoiceTemplate = template.Must(template.New("invoice").Parse(`<!DOCTYPE html>
<html><body style="margin:0;padding:0;background:#f3f4f6;font-family:Arial,sans-serif">
<table width="100%" cellpadding="0" cellspacing="0" style="background:#f3f4f6;padding:32px 16px"><tr><td align="center">
<table width="600" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:12px">
<tr><td style="background:#1e40af;padding:28px 32px">
<h1 style="color:#ffffff;font-size:22px;margin:0">INVOICE</h1>
<p style="color:#bfdbfe;margin:4px 0 0">{{.InvoiceNumber}}</p>
</td></tr>
<tr><td style="padding:28px 32px 0">
<p>Dear <strong>{{.CustomerName}}</strong>,</p>
<p>Please find your invoice from {{.BusinessName}}. Payment is due by <strong>{{.DueDate}}</strong>.</p>
</td></tr>
<tr><td style="padding:20px 32px">
<p>Issued {{.IssuedDate}} · Due {{.DueDate}} · Amount due {{.Currency}} {{.GrandTotal}}</p>
</td></tr>
<tr><td style="padding:0 32px">
<table width="100%" cellpadding="0" cellspacing="0">
<tr>
<th align="left">Description</th><th align="right">Qty</th><th align="right">Unit price</th><th align="right">Amount</th>
</tr>
{{range .Lines}}
<tr>
<td>{{.Description}}</td>
<td align="right">{{.Quantity}}</td>
<td align="right">{{.UnitPrice}}</td>
<td align="right">{{.Amount}}</td>
</tr>
{{end}}
</table>
</td></tr>
<tr><td style="padding:16px 32px">
<p>Subtotal {{.Currency}} {{.Subtotal}}<br/>Tax {{.Currency}} {{.TaxAmount}}<br/><strong>Grand total {{.Currency}} {{.GrandTotal}}</strong></p>
{{if .Notes}}<p>{{.Notes}}</p>{{end}}
</td></tr>
<tr><td style="padding:20px 32px;color:#6b7280">
This invoice was sent by {{.BusinessName}}{{if .BusinessEmail}} · {{.BusinessEmail}}{{end}}.
</td></tr>
</table>
</td></tr></table>
</body></html>`))

var paymentTemplate = template.Must(template.New("payment").Parse(`<!DOCTYPE html>
<html><body style="margin:0;padding:0;background:#f3f4f6;font-family:Arial,sans-serif">
<table width="100%" cellpadding="0" cellspacing="0" style="background:#f3f4f6;padding:32px 16px"><tr><td align="center">
<table width="600" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:12px">
<tr><td style="background:#047857;padding:28px 32px">
<h1 style="color:#ffffff;font-size:22px;margin:0">PAYMENT RECEIVED</h1>
<p style="color:#a7f3d0;margin:4px 0 0">{{.InvoiceNumber}}</p>
</td></tr>
<tr><td style="padding:28px 32px">
<p>Dear <strong>{{.CustomerName}}</strong>,</p>
<p>We received your payment of <strong>{{.Currency}} {{.Amount}}</strong> for invoice {{.InvoiceNumber}} on {{.PaymentDate}}.</p>
<p>Method: {{.Method}}{{if .Reference}}<br/>Reference: {{.Reference}}{{end}}</p>
{{if .FullyPaid}}
<p>This invoice is now fully paid. Thank you.</p>
{{else}}
<p>Balance still due: <strong>{{.Currency}} {{.BalanceDue}}</strong>.</p>
{{end}}
<p style="color:#6b7280">This receipt was sent by {{.BusinessName}}{{if .BusinessEmail}} · {{.BusinessEmail}}{{end}}.</p>
</td></tr>
</table>
</td></tr></table>
</body></html>`))
