package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	OrgJWT     JWTConfig
	SuperJWT   JWTConfig
	RateLimit  RateLimitConfig
	Upload     UploadConfig
	WhatsApp   WhatsAppConfig
	Logging    LoggingConfig
	SuperAdmin SuperAdminConfig
	Email      EmailConfig
	Push       PushConfig
}

type ServerConfig struct {
	Port            string
	Environment     string
	AllowedOrigins  []string
	FrontendURL     string
	EnableTestEmail bool
	EmailTestToken  string
}

type DatabaseConfig struct {
	URL             string
	MinConns        int
	MaxConns        int
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

type JWTConfig struct {
	Secret             string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

type RateLimitConfig struct {
	AuthRPM int
	APIRPM  int
}

type UploadConfig struct {
	Dir       string
	MaxSizeMB int
}

type WhatsAppConfig struct {
	APIURL string
	APIKey string
}

type LoggingConfig struct {
	Level  string
	Format string
}

type SuperAdminConfig struct {
	IPAllowlist     []string
	BootstrapSecret string
}

type EmailConfig struct {
	APIKey      string
	APIURL      string
	SenderEmail string
	SenderName  string
	SMTPHost    string
	SMTPPort    string
	SMTPLogin   string
	SMTPKey     string
}

// PushConfig holds the Web Push (VAPID) keys. Leave them empty until the
// keys are ready; in-app alerts still work without them.
type PushConfig struct {
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	VAPIDSubject    string
}

func Load() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	cfg := &Config{
		Server: ServerConfig{
			Port:            getEnv("SERVER_PORT", "8080"),
			Environment:     getEnv("ENVIRONMENT", "development"),
			AllowedOrigins:  splitList(getEnv("ALLOWED_ORIGINS", "http://localhost:5173")),
			FrontendURL:     getEnv("FRONTEND_URL", "http://localhost:5173"),
			EnableTestEmail: getBool("ENABLE_TEST_EMAIL", false),
			EmailTestToken:  getEnv("EMAIL_TEST_TOKEN", ""),
		},
		Database: DatabaseConfig{
			URL:             getEnv("DATABASE_URL", ""),
			MinConns:        getInt("DB_MIN_CONNS", 5),
			MaxConns:        getInt("DB_MAX_CONNS", 25),
			MaxConnLifetime: parseDuration(getEnv("DB_MAX_CONN_LIFETIME", "1h")),
			MaxConnIdleTime: parseDuration(getEnv("DB_MAX_CONN_IDLE_TIME", "30m")),
		},
		OrgJWT: JWTConfig{
			Secret:             getEnv("ORG_JWT_SECRET", ""),
			AccessTokenExpiry:  parseDuration(getEnv("ORG_ACCESS_TOKEN_EXPIRY", "15m")),
			RefreshTokenExpiry: parseDuration(getEnv("ORG_REFRESH_TOKEN_EXPIRY", "168h")),
		},
		SuperJWT: JWTConfig{
			Secret:             getEnv("SA_JWT_SECRET", ""),
			AccessTokenExpiry:  parseDuration(getEnv("SA_ACCESS_TOKEN_EXPIRY", "15m")),
			RefreshTokenExpiry: parseDuration(getEnv("SA_REFRESH_TOKEN_EXPIRY", "24h")),
		},
		RateLimit: RateLimitConfig{
			AuthRPM: getInt("RATE_LIMIT_AUTH_RPM", 10),
			APIRPM:  getInt("RATE_LIMIT_API_RPM", 300),
		},
		Upload: UploadConfig{
			Dir:       getEnv("UPLOAD_DIR", "./uploads"),
			MaxSizeMB: getInt("MAX_UPLOAD_SIZE_MB", 2),
		},
		WhatsApp: WhatsAppConfig{
			APIURL: getEnv("WHATSAPP_API_URL", ""),
			APIKey: getEnv("WHATSAPP_API_KEY", ""),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		SuperAdmin: SuperAdminConfig{
			IPAllowlist:     splitList(getEnv("SA_IP_ALLOWLIST", "")),
			BootstrapSecret: getEnv("SA_BOOTSTRAP_SECRET", ""),
		},
		Email: EmailConfig{
			APIKey:      firstNonEmpty(getEnv("BREVO_API_KEY", ""), getEnv("BRAVO_KEY_SECRET", "")),
			APIURL:      getEnv("BREVO_API_URL", "https://api.brevo.com/v3/smtp/email"),
			SenderEmail: firstNonEmpty(getEnv("BREVO_SENDER_EMAIL", ""), getEnv("BRAVO_SMTP_USER", "")),
			SenderName:  getEnv("BREVO_SENDER_NAME", "Invoice Flow"),
			SMTPHost:    getEnv("BREVO_SMTP_HOST", "smtp-relay.brevo.com"),
			SMTPPort:    getEnv("BREVO_SMTP_PORT", "587"),
			SMTPLogin:   getEnv("BREVO_SMTP_LOGIN", ""),
			SMTPKey:     getEnv("BREVO_SMTP_KEY", ""),
		},
		Push: PushConfig{
			VAPIDPublicKey:  getEnv("PUSH_VAPID_PUBLIC_KEY", ""),
			VAPIDPrivateKey: getEnv("PUSH_VAPID_PRIVATE_KEY", ""),
			VAPIDSubject:    getEnv("PUSH_VAPID_SUBJECT", "mailto:admin@localhost"),
		},
	}

	if missing := missingSettings(cfg.Database.URL, cfg.OrgJWT.Secret, cfg.SuperJWT.Secret, cfg.Server.Environment); len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	viper.SetDefault(key, defaultValue)
	return viper.GetString(key)
}

func getInt(key string, defaultValue int) int {
	viper.SetDefault(key, defaultValue)
	return viper.GetInt(key)
}

func getBool(key string, defaultValue bool) bool {
	viper.SetDefault(key, defaultValue)
	return viper.GetBool(key)
}

func splitList(val string) []string {
	if strings.TrimSpace(val) == "" {
		return []string{}
	}
	parts := strings.Split(val, ",")
	clean := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			clean = append(clean, p)
		}
	}
	return clean
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func missingSettings(databaseURL, orgSecret, superSecret, environment string) []string {
	var missing []string
	if strings.TrimSpace(databaseURL) == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if strings.TrimSpace(orgSecret) == "" {
		missing = append(missing, "ORG_JWT_SECRET")
	} else if environment == "production" && len(orgSecret) < 32 {
		missing = append(missing, "ORG_JWT_SECRET (must be at least 32 characters in production)")
	}
	if strings.TrimSpace(superSecret) == "" {
		missing = append(missing, "SA_JWT_SECRET")
	} else if environment == "production" && len(superSecret) < 32 {
		missing = append(missing, "SA_JWT_SECRET (must be at least 32 characters in production)")
	}
	return missing
}

func parseDuration(s string) time.Duration {
	d, _ := time.ParseDuration(s)
	return d
}
