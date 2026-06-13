package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"

	"github.com/your-org/invoice-backend/internal/config"
)

type Client struct {
	cfg config.EmailConfig
}

func NewClient(cfg config.EmailConfig) *Client {
	return &Client{
		cfg: cfg,
	}
}

func (c *Client) dialSMTP() (*smtp.Client, error) {
	addr := fmt.Sprintf("%s:%s", c.cfg.Host, c.cfg.Port)

	client, err := smtp.Dial(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SMTP server: %w", err)
	}

	tlsConfig := &tls.Config{
		ServerName: c.cfg.Host,
	}
	if err = client.StartTLS(tlsConfig); err != nil {
		return nil, fmt.Errorf("failed to start TLS: %w", err)
	}

	auth := smtp.PlainAuth("", c.cfg.Username, c.cfg.Password, c.cfg.Host)
	if err = client.Auth(auth); err != nil {
		return nil, fmt.Errorf("SMTP authentication failed: %w", err)
	}

	fmt.Println("[Email Service] SMTP connection and authentication successful")
	return client, nil
}

func (c *Client) sendEmail(toEmail, subject, htmlBody string) error {
	client, err := c.dialSMTP()
	if err != nil {
		return err
	}
	defer client.Quit()

	if err = client.Mail(c.cfg.Username); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	if err = client.Rcpt(toEmail); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open data writer: %w", err)
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n%s",
		c.cfg.Username, toEmail, subject, htmlBody)

	if _, err = fmt.Fprint(w, msg); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err = w.Close(); err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return nil
}

func (c *Client) SendWelcomeEmail(toEmail, orgName, password string) error {
	subject := "Welcome to Invoice System - Your Login Credentials"
	body := fmt.Sprintf(`
		<h1>Welcome to Invoice System, %s!</h1>
		<p>Your organisation has been created successfully.</p>
		<p><strong>Your login credentials:</strong></p>
		<ul>
			<li><strong>Email:</strong> %s</li>
			<li><strong>Initial Password:</strong> %s</li>
		</ul>
		<p>Please log in and change your password as soon as possible.</p>
	`, orgName, toEmail, password)

	if err := c.sendEmail(toEmail, subject, body); err != nil {
		fmt.Printf("[Email Service] SMTP Error sending welcome email to %s: %v\n", toEmail, err)
		return err
	}

	fmt.Printf("[Email Service] Successfully sent welcome email to %s\n", toEmail)
	return nil
}

func (c *Client) SendForgotPasswordEmail(toEmail, token, frontendURL string) error {
	subject := "Password Reset Request - Invoice System"
	resetLink := fmt.Sprintf("%s/new-password?refreshtoken=%v", frontendURL, token)
	body := fmt.Sprintf(`
		<h1>Password Reset Request</h1>
		<p>You requested a password reset for your Invoice System account.</p>
		<p>Click the link below to reset your password. This link is valid for 10 minutes.</p>
		<p><a href="%s" style="background-color: #4CAF50; color: white; padding: 14px 20px; text-decoration: none; border-radius: 4px; display: inline-block;">Reset Password</a></p>
		<p>If you did not request this, you can safely ignore this email.</p>
		<p>Link: %s</p>
	`, resetLink, resetLink)

	if err := c.sendEmail(toEmail, subject, body); err != nil {
		fmt.Printf("[Email Service] SMTP Error sending forgot password email to %s: %v\n", toEmail, err)
		return err
	}

	fmt.Printf("[Email Service] Successfully sent forgot password email to %s\n", toEmail)
	return nil
}

// InvoiceEmailData holds all data needed to render the invoice email.
type InvoiceEmailData struct {
	BusinessName    string
	BusinessEmail   string
	InvoiceNumber   string
	IssuedDate      string
	DueDate         string
	Currency        string
	CustomerName    string
	Subtotal        string
	TaxAmount       string
	GrandTotal      string
	Notes           string
	ItemsHTML       string
}

func (c *Client) SendInvoiceEmail(toEmail string, d InvoiceEmailData) error {
	subject := fmt.Sprintf("Invoice %s from %s", d.InvoiceNumber, d.BusinessName)

	notesSection := ""
	if d.Notes != "" {
		notesSection = fmt.Sprintf(`
		<tr><td colspan="2" style="padding:20px 32px 0">
			<p style="font-size:13px;color:#6b7280;font-weight:600;margin-bottom:4px;text-transform:uppercase;letter-spacing:.5px">Notes</p>
			<p style="font-size:13px;color:#374151;line-height:1.6">%s</p>
		</td></tr>`, d.Notes)
	}

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"/><meta name="viewport" content="width=device-width,initial-scale=1"/></head>
<body style="margin:0;padding:0;background:#f3f4f6;font-family:'Helvetica Neue',Arial,sans-serif">
<table width="100%%" cellpadding="0" cellspacing="0" style="background:#f3f4f6;padding:32px 16px">
<tr><td align="center">
<table width="600" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:12px;overflow:hidden;box-shadow:0 1px 3px rgba(0,0,0,.1)">

  <!-- Header -->
  <tr><td style="background:#1e40af;padding:28px 32px">
    <h1 style="color:#ffffff;font-size:22px;font-weight:800;margin:0">INVOICE</h1>
    <p style="color:#bfdbfe;font-size:13px;margin:4px 0 0"># %s</p>
  </td></tr>

  <!-- Greeting -->
  <tr><td style="padding:28px 32px 0">
    <p style="font-size:15px;color:#1f2937;margin:0">Dear <strong>%s</strong>,</p>
    <p style="font-size:14px;color:#6b7280;margin:8px 0 0;line-height:1.6">
      Please find your invoice details below. Payment is due by <strong>%s</strong>.
    </p>
  </td></tr>

  <!-- Invoice Meta -->
  <tr><td style="padding:20px 32px">
    <table width="100%%" cellpadding="0" cellspacing="0" style="background:#f9fafb;border-radius:8px;border:1px solid #e5e7eb">
      <tr>
        <td style="padding:14px 20px;border-right:1px solid #e5e7eb">
          <p style="font-size:11px;color:#6b7280;text-transform:uppercase;letter-spacing:.5px;margin:0 0 4px">Issue Date</p>
          <p style="font-size:14px;color:#1f2937;font-weight:600;margin:0">%s</p>
        </td>
        <td style="padding:14px 20px;border-right:1px solid #e5e7eb">
          <p style="font-size:11px;color:#6b7280;text-transform:uppercase;letter-spacing:.5px;margin:0 0 4px">Due Date</p>
          <p style="font-size:14px;color:#dc2626;font-weight:600;margin:0">%s</p>
        </td>
        <td style="padding:14px 20px">
          <p style="font-size:11px;color:#6b7280;text-transform:uppercase;letter-spacing:.5px;margin:0 0 4px">Amount Due</p>
          <p style="font-size:14px;color:#1e40af;font-weight:700;margin:0">%s %s</p>
        </td>
      </tr>
    </table>
  </td></tr>

  <!-- Items -->
  <tr><td style="padding:0 32px">
    <table width="100%%" cellpadding="0" cellspacing="0" style="border-collapse:collapse">
      <thead>
        <tr style="background:#eef2ff">
          <th style="padding:10px 12px;text-align:left;font-size:11px;color:#1e40af;text-transform:uppercase;letter-spacing:.3px">Description</th>
          <th style="padding:10px 12px;text-align:right;font-size:11px;color:#1e40af;text-transform:uppercase;letter-spacing:.3px">Qty</th>
          <th style="padding:10px 12px;text-align:right;font-size:11px;color:#1e40af;text-transform:uppercase;letter-spacing:.3px">Unit Price</th>
          <th style="padding:10px 12px;text-align:right;font-size:11px;color:#1e40af;text-transform:uppercase;letter-spacing:.3px">Amount</th>
        </tr>
      </thead>
      %s
    </table>
  </td></tr>

  <!-- Totals -->
  <tr><td style="padding:16px 32px">
    <table width="100%%" cellpadding="0" cellspacing="0">
      <tr>
        <td width="60%%"></td>
        <td width="40%%">
          <table width="100%%" cellpadding="0" cellspacing="0">
            <tr>
              <td style="padding:5px 0;font-size:13px;color:#6b7280">Subtotal</td>
              <td style="padding:5px 0;font-size:13px;color:#6b7280;text-align:right">%s %s</td>
            </tr>
            <tr>
              <td style="padding:5px 0;font-size:13px;color:#6b7280">Tax</td>
              <td style="padding:5px 0;font-size:13px;color:#6b7280;text-align:right">%s %s</td>
            </tr>
            <tr>
              <td style="padding:10px 0 5px;font-size:15px;font-weight:700;color:#1e40af;border-top:2px solid #1e40af">Grand Total</td>
              <td style="padding:10px 0 5px;font-size:15px;font-weight:700;color:#1e40af;text-align:right;border-top:2px solid #1e40af">%s %s</td>
            </tr>
          </table>
        </td>
      </tr>
    </table>
  </td></tr>

  %s

  <!-- Footer -->
  <tr><td style="background:#f9fafb;padding:20px 32px;border-top:1px solid #e5e7eb">
    <p style="font-size:12px;color:#9ca3af;margin:0;text-align:center">
      This invoice was sent by <strong>%s</strong>
      %s
      <br/>Thank you for your business.
    </p>
  </td></tr>

</table>
</td></tr>
</table>
</body>
</html>`,
		d.InvoiceNumber,
		d.CustomerName,
		d.DueDate,
		d.IssuedDate,
		d.DueDate,
		d.Currency, d.GrandTotal,
		d.ItemsHTML,
		d.Currency, d.Subtotal,
		d.Currency, d.TaxAmount,
		d.Currency, d.GrandTotal,
		notesSection,
		d.BusinessName,
		func() string {
			if d.BusinessEmail != "" {
				return fmt.Sprintf("&bull; %s", d.BusinessEmail)
			}
			return ""
		}(),
	)

	if err := c.sendEmail(toEmail, subject, body); err != nil {
		fmt.Printf("[Email Service] SMTP Error sending invoice email to %s: %v\n", toEmail, err)
		return err
	}

	fmt.Printf("[Email Service] Successfully sent invoice email %s to %s\n", d.InvoiceNumber, toEmail)
	return nil
}
