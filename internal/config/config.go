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
	Bravo      BravoConfig
}

type ServerConfig struct {
	Port           string
	Environment    string
	AllowedOrigins []string
	FrontendURL    string
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
	IPAllowlist []string
}

type BravoConfig struct {
	APIKey   string
	Username string
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
			Port:           getEnv("SERVER_PORT", "8080"),
			Environment:    getEnv("ENVIRONMENT", "development"),
			AllowedOrigins: strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:8081"), ","),
			FrontendURL:    getEnv("FRONTEND_URL", "http://localhost:5173"),
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
			IPAllowlist: func() []string {
				val := getEnv("SA_IP_ALLOWLIST", "")
				if val == "" {
					return []string{}
				}
				parts := strings.Split(val, ",")
				var clean []string
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if p != "" {
						clean = append(clean, p)
					}
				}
				return clean
			}(),
		},
		Bravo: BravoConfig{
			APIKey:   getEnv("BRAVO_KEY_SECRET", ""),
			Username: getEnv("BRAVO_SMTP_USER", ""),
		},
	}

	// Validate required fields
	var missing []string
	if cfg.Database.URL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if cfg.OrgJWT.Secret == "" {
		missing = append(missing, "ORG_JWT_SECRET")
	}
	if cfg.SuperJWT.Secret == "" {
		missing = append(missing, "SA_JWT_SECRET")
	}
	if len(missing) > 0 {
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

func parseDuration(s string) time.Duration {
	d, _ := time.ParseDuration(s)
	return d
}
