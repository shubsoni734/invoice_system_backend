package email

import (
	"fmt"
	"net/smtp"
)

type Client struct {
	smtpHost string
	smtpPort string
	smtpUser string // Usually your Brevo login email
	apiKey   string // This is your xsmtpsib- key
}

func NewClient(apiKey, smtpUser string) *Client {
	return &Client{
		smtpHost: "smtp-relay.brevo.com",
		smtpPort: "587",
		smtpUser: smtpUser,
		apiKey:   apiKey,
	}
}


func (c *Client) SendWelcomeEmail(toEmail, orgName, password string) error {
	if c.apiKey == "" {
		return fmt.Errorf("bravo SMTP key is not configured")
	}

	subject := "Subject: Welcome to Invoice System - Your Login Credentials\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
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

	msg := []byte(subject + mime + body)
	auth := smtp.PlainAuth("", c.smtpUser, c.apiKey, c.smtpHost)

	addr := fmt.Sprintf("%s:%s", c.smtpHost, c.smtpPort)
	err := smtp.SendMail(addr, auth, c.smtpUser, []string{toEmail}, msg)
	if err != nil {
		fmt.Printf("[Email Service] SMTP Error: %v\n", err)
		return err
	}

	fmt.Printf("[Email Service] Successfully sent welcome email via SMTP to %s\n", toEmail)
	return nil
}



