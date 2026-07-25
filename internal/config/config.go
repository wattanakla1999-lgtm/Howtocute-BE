package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port                            string
	DSN                             string
	AllowOrigin                     string
	JWTSecret                       string
	JWTTTL                          time.Duration
	LineLoginChannelID              string
	LineMessagingChannelAccessToken string
	LineShopOwnerUserID             string
	LineBookingDetailsURL           string
	CustomerJWTSecret               string
	CustomerJWTTTL                  time.Duration
	AdminUsername                   string
	AdminPassword                   string
	AdminName                       string
}

func Load() Config {
	loadEnvFile(".env")

	return Config{
		Port:                            getEnv("PORT", "8080"),
		DSN:                             getEnv("DATABASE_DSN", "host=localhost user=nailly password=nailly1234 dbname=nailly_db port=5432 sslmode=disable"),
		AllowOrigin:                     normalizeOrigin(getEnv("ALLOW_ORIGIN", "*")),
		JWTSecret:                       getEnv("JWT_SECRET", "dev-only-change-me-before-production"),
		JWTTTL:                          time.Duration(getEnvInt("JWT_TTL_HOURS", 24)) * time.Hour,
		LineLoginChannelID:              getEnv("LINE_LOGIN_CHANNEL_ID", getEnv("LINE_CHANNEL_ID", "")),
		LineMessagingChannelAccessToken: getEnv("LINE_MESSAGING_CHANNEL_ACCESS_TOKEN", ""),
		LineShopOwnerUserID:             getEnv("LINE_SHOP_OWNER_USER_ID", getEnv("UDI", "")),
		LineBookingDetailsURL:           normalizeOrigin(getEnv("LINE_BOOKING_DETAILS_URL", "")),
		CustomerJWTSecret:               getEnv("CUSTOMER_JWT_SECRET", ""),
		CustomerJWTTTL:                  time.Duration(getEnvInt("CUSTOMER_JWT_TTL_HOURS", 720)) * time.Hour,
		AdminUsername:                   getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword:                   getEnv("ADMIN_PASSWORD", "nailly2025"),
		AdminName:                       getEnv("ADMIN_NAME", "ผู้ดูแลระบบ"),
	}
}

func getEnvInt(key string, fallback int) int {
	value, err := strconv.Atoi(getEnv(key, ""))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func normalizeOrigin(origin string) string {
	if origin == "*" {
		return origin
	}

	return strings.TrimRight(origin, "/")
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func loadEnvFile(path string) {
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}
