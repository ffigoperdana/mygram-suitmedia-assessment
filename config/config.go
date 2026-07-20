package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port                     string
	GinMode                  string
	DBHost                   string
	DBUser                   string
	DBPassword               string
	DBName                   string
	DBPort                   string
	DBSSLMode                string
	RedisEnabled             bool
	RedisAddr                string
	RedisPassword            string
	RedisDB                  int
	RedisCacheTTLSeconds     int
	RateLimitRequests        int
	AuthRateLimitRequests    int
	RateLimitWindowSeconds   int
	JWTSecret                string
	JWTExpirationHours       int
	CORSAllowedOrigins       []string
	PublicOpenAPI            bool
	SwaggerUIMode            string
	S3Endpoint               string
	S3Region                 string
	S3Bucket                 string
	S3AccessKeyID            string
	S3SecretAccessKey        string
	S3ForcePathStyle         bool
	S3PublicBaseURL          string
	S3UploadMaxBytes         int64
	PushNotificationsEnabled bool
	VAPIDPublicKey           string
	VAPIDPrivateKey          string
	VAPIDSubject             string
	BootstrapAdminEmail      string
	BootstrapAdminUsername   string
	BootstrapAdminPassword   string
	BootstrapAdminAge        int
	BootstrapUserEmail       string
	BootstrapUserUsername    string
	BootstrapUserPassword    string
	BootstrapUserAge         int
}

func Load() Config {
	_ = LoadDotEnv(".env")
	ginMode := env("GIN_MODE", "debug")

	return Config{
		Port:                     env("PORT", "8080"),
		GinMode:                  ginMode,
		DBHost:                   env("DB_HOST", "localhost"),
		DBUser:                   env("DB_USER", "postgres"),
		DBPassword:               env("DB_PASSWORD", "admin"),
		DBName:                   env("DB_NAME", "finalproject"),
		DBPort:                   env("DB_PORT", "5432"),
		DBSSLMode:                env("DB_SSLMODE", "disable"),
		RedisEnabled:             envBool("REDIS_ENABLED", false),
		RedisAddr:                env("REDIS_ADDR", ""),
		RedisPassword:            env("REDIS_PASSWORD", ""),
		RedisDB:                  envIntAllowZero("REDIS_DB", 0),
		RedisCacheTTLSeconds:     envInt("REDIS_CACHE_TTL_SECONDS", 60),
		RateLimitRequests:        envInt("RATE_LIMIT_REQUESTS", 120),
		AuthRateLimitRequests:    envInt("AUTH_RATE_LIMIT_REQUESTS", 10),
		RateLimitWindowSeconds:   envInt("RATE_LIMIT_WINDOW_SECONDS", 60),
		JWTSecret:                env("JWT_SECRET", ""),
		JWTExpirationHours:       envInt("JWT_EXPIRATION_HOURS", 24),
		CORSAllowedOrigins:       envCSV("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173"),
		PublicOpenAPI:            envBool("PUBLIC_OPENAPI_ENABLED", true),
		SwaggerUIMode:            swaggerUIMode(ginMode),
		S3Endpoint:               strings.TrimRight(env("S3_ENDPOINT", ""), "/"),
		S3Region:                 env("S3_REGION", "garage"),
		S3Bucket:                 env("S3_BUCKET", ""),
		S3AccessKeyID:            env("S3_ACCESS_KEY_ID", ""),
		S3SecretAccessKey:        env("S3_SECRET_ACCESS_KEY", ""),
		S3ForcePathStyle:         envBool("S3_FORCE_PATH_STYLE", true),
		S3PublicBaseURL:          strings.TrimRight(env("S3_PUBLIC_BASE_URL", ""), "/"),
		S3UploadMaxBytes:         int64(envInt("S3_UPLOAD_MAX_MB", 4)) * 1024 * 1024,
		PushNotificationsEnabled: envBool("PUSH_NOTIFICATIONS_ENABLED", false),
		VAPIDPublicKey:           env("VAPID_PUBLIC_KEY", ""),
		VAPIDPrivateKey:          env("VAPID_PRIVATE_KEY", ""),
		VAPIDSubject:             env("VAPID_SUBJECT", "mailto:admin@example.com"),
		BootstrapAdminEmail:      env("BOOTSTRAP_ADMIN_EMAIL", ""),
		BootstrapAdminUsername:   env("BOOTSTRAP_ADMIN_USERNAME", ""),
		BootstrapAdminPassword:   env("BOOTSTRAP_ADMIN_PASSWORD", ""),
		BootstrapAdminAge:        envInt("BOOTSTRAP_ADMIN_AGE", 21),
		BootstrapUserEmail:       env("BOOTSTRAP_USER_EMAIL", ""),
		BootstrapUserUsername:    env("BOOTSTRAP_USER_USERNAME", ""),
		BootstrapUserPassword:    env("BOOTSTRAP_USER_PASSWORD", ""),
		BootstrapUserAge:         envInt("BOOTSTRAP_USER_AGE", 18),
	}
}

func (cfg Config) ObjectStorageConfigured() bool {
	return cfg.S3Endpoint != "" &&
		cfg.S3Region != "" &&
		cfg.S3Bucket != "" &&
		cfg.S3AccessKeyID != "" &&
		cfg.S3SecretAccessKey != ""
}

func (cfg Config) PushNotificationsConfigured() bool {
	return cfg.PushNotificationsEnabled &&
		cfg.VAPIDPublicKey != "" &&
		cfg.VAPIDPrivateKey != "" &&
		cfg.VAPIDSubject != ""
}

func LoadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)

		if key == "" {
			continue
		}

		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		_ = os.Setenv(key, value)
	}

	if err := scanner.Err(); err != nil {
		_ = file.Close()
		return err
	}

	return file.Close()
}

func env(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}

func envIntAllowZero(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return fallback
	}

	return parsed
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if value == "" {
		return fallback
	}

	switch value {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}

func envCSV(key string, fallback string) []string {
	raw := env(key, fallback)
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))

	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value != "" {
			values = append(values, value)
		}
	}

	return values
}

func swaggerUIMode(ginMode string) string {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("SWAGGER_UI_MODE")))
	if mode == "" {
		if ginMode == "release" {
			return "disabled"
		}
		return "internal"
	}

	switch mode {
	case "internal", "public", "disabled":
		return mode
	default:
		if ginMode == "release" {
			return "disabled"
		}
		return "internal"
	}
}
