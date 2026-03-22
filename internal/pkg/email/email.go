package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
)

const (
	smtpHost     = "smtp-relay.brevo.com"
	smtpPort     = "587"
	smtpLogin    = "7cfd96001@smtp-brevo.com"
	smtpPassword = "xsmtpsib-9a99dbe7504b118a2a2d961a7a97b3d299be2dd24298de8d92d877b3ecccace1-oGJL4w4RYcMsb1KJ"
)

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) dialSMTP() (*smtp.Client, error) {
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)

	client, err := smtp.Dial(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SMTP server: %w", err)
	}

	tlsConfig := &tls.Config{
		ServerName: smtpHost,
	}
	if err = client.StartTLS(tlsConfig); err != nil {
		return nil, fmt.Errorf("failed to start TLS: %w", err)
	}

	auth := smtp.PlainAuth("", smtpLogin, smtpPassword, smtpHost)
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

	if err = client.Mail(smtpLogin); err != nil {
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
		smtpLogin, toEmail, subject, htmlBody)

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
