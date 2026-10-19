package notify

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/invoice-backend/internal/config"
	"github.com/your-org/invoice-backend/internal/pkg/email"
)

type orgMail struct {
	APIKey             string
	SenderEmail        string
	SenderName         string
	SMTPHost           string
	SMTPPort           string
	SMTPLogin          string
	SMTPKey            string
	NotifyEmailEnabled bool
	PushEnabled        bool
}

var (
	pool     *pgxpool.Pool
	baseMail config.EmailConfig
	pushCfg  config.PushConfig
)

// Configure stores the database and server keys used by activity alerts.
func Configure(db *pgxpool.Pool, mail config.EmailConfig, push config.PushConfig) {
	pool = db
	baseMail = mail
	pushCfg = push
}

// PushConfigured reports whether the server has a VAPID key pair.
func PushConfigured() bool {
	return strings.TrimSpace(pushCfg.VAPIDPublicKey) != "" && strings.TrimSpace(pushCfg.VAPIDPrivateKey) != ""
}

// PublicKey is the VAPID public key the browser uses to subscribe.
func PublicKey() string {
	return strings.TrimSpace(pushCfg.VAPIDPublicKey)
}

// Publish stores an in-app alert and, when configured, emails and pushes it to admins.
func Publish(orgID uuid.UUID, title, message, kind string) {
	if pool == nil || orgID == uuid.Nil || strings.TrimSpace(title) == "" {
		return
	}
	if kind == "" {
		kind = "info"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if _, err := pool.Exec(ctx, `
		INSERT INTO notifications (organisation_id, title, message, kind)
		VALUES ($1, $2, $3, $4)
	`, orgID, title, message, kind); err != nil {
		log.Printf("notify: save alert: %v", err)
		return
	}
	go deliver(orgID, title, message)
}

func deliver(orgID uuid.UUID, title, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	settings, err := loadOrgMail(ctx, orgID)
	if err != nil {
		log.Printf("notify: load channel settings: %v", err)
		return
	}
	if settings.NotifyEmailEnabled {
		sendAdminEmail(ctx, orgID, settings, title, message)
	}
	if settings.PushEnabled && PushConfigured() {
		sendAdminPush(ctx, orgID, title, message)
	}
}

func loadOrgMail(ctx context.Context, orgID uuid.UUID) (orgMail, error) {
	var settings orgMail
	err := pool.QueryRow(ctx, `
		SELECT COALESCE(brevo_api_key, ''), COALESCE(brevo_sender_email, ''), COALESCE(brevo_sender_name, ''),
		       COALESCE(brevo_smtp_host, ''), COALESCE(brevo_smtp_port, ''), COALESCE(brevo_smtp_login, ''),
		       COALESCE(brevo_smtp_key, ''), COALESCE(notify_email_enabled, true), COALESCE(push_enabled, false)
		FROM settings WHERE organisation_id = $1
	`, orgID).Scan(
		&settings.APIKey, &settings.SenderEmail, &settings.SenderName,
		&settings.SMTPHost, &settings.SMTPPort, &settings.SMTPLogin, &settings.SMTPKey,
		&settings.NotifyEmailEnabled, &settings.PushEnabled,
	)
	if err != nil {
		return orgMail{NotifyEmailEnabled: true}, nil
	}
	return settings, nil
}

func resolveMail(settings orgMail) config.EmailConfig {
	orgSMTP := strings.TrimSpace(settings.SMTPLogin) != "" && strings.TrimSpace(settings.SMTPKey) != ""
	orgAPI := strings.TrimSpace(settings.APIKey) != "" && strings.TrimSpace(settings.SenderEmail) != ""
	if !orgSMTP && !orgAPI {
		return baseMail
	}
	cfg := config.EmailConfig{
		APIURL:      baseMail.APIURL,
		SMTPHost:    firstText(settings.SMTPHost, "smtp-relay.brevo.com"),
		SMTPPort:    firstText(settings.SMTPPort, "587"),
		SenderName:  firstText(settings.SenderName, "Invoice Flow"),
		SenderEmail: strings.TrimSpace(settings.SenderEmail),
	}
	if orgSMTP {
		cfg.SMTPLogin = strings.TrimSpace(settings.SMTPLogin)
		cfg.SMTPKey = strings.TrimSpace(settings.SMTPKey)
		if cfg.SenderEmail == "" {
			cfg.SenderEmail = cfg.SMTPLogin
		}
	}
	if orgAPI {
		cfg.APIKey = strings.TrimSpace(settings.APIKey)
	}
	return cfg
}

func sendAdminEmail(ctx context.Context, orgID uuid.UUID, settings orgMail, title, message string) {
	client := email.NewClient(resolveMail(settings))
	if !client.Configured() {
		return
	}
	rows, err := pool.Query(ctx, `
		SELECT DISTINCT u.email
		FROM users u
		LEFT JOIN roles r ON r.id = u.role_id
		WHERE u.organisation_id = $1 AND u.is_active = true
		  AND (LOWER(u.role) = 'admin' OR COALESCE(r.is_system, false) = true)
	`, orgID)
	if err != nil {
		log.Printf("notify: admin emails: %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var address string
		if err := rows.Scan(&address); err != nil {
			continue
		}
		if err := client.SendActivityEmail(address, title, message); err != nil {
			log.Printf("notify: email %s: %v", address, err)
		}
	}
}

func sendAdminPush(ctx context.Context, orgID uuid.UUID, title, message string) {
	rows, err := pool.Query(ctx, `
		SELECT ps.endpoint, ps.p256dh, ps.auth_secret
		FROM push_subscriptions ps
		JOIN users u ON u.id = ps.user_id
		LEFT JOIN roles r ON r.id = u.role_id
		WHERE ps.organisation_id = $1 AND u.is_active = true
		  AND (LOWER(u.role) = 'admin' OR COALESCE(r.is_system, false) = true)
	`, orgID)
	if err != nil {
		log.Printf("notify: push subscriptions: %v", err)
		return
	}
	defer rows.Close()
	payload, _ := json.Marshal(map[string]string{"title": title, "body": message})
	subject := strings.TrimSpace(pushCfg.VAPIDSubject)
	if subject == "" {
		subject = "mailto:admin@localhost"
	}
	for rows.Next() {
		var endpoint, p256dh, authSecret string
		if err := rows.Scan(&endpoint, &p256dh, &authSecret); err != nil {
			continue
		}
		resp, err := webpush.SendNotification(payload, &webpush.Subscription{
			Endpoint: endpoint,
			Keys:     webpush.Keys{P256dh: p256dh, Auth: authSecret},
		}, &webpush.Options{
			Subscriber:      subject,
			VAPIDPublicKey:  strings.TrimSpace(pushCfg.VAPIDPublicKey),
			VAPIDPrivateKey: strings.TrimSpace(pushCfg.VAPIDPrivateKey),
			TTL:             60,
		})
		if err != nil {
			log.Printf("notify: push: %v", err)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == 404 || resp.StatusCode == 410 {
			_, _ = pool.Exec(ctx, `DELETE FROM push_subscriptions WHERE endpoint = $1`, endpoint)
		}
	}
}

func firstText(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
