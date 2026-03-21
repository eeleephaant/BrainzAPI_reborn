package config

import (
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		Postgres PostgresConfig `env-prefix:"DB_"`
		Auth     AuthConfig
		Redis    RedisConfig `env-prefix:"REDIS_"`
		App      AppConfig   `env-prefix:"APP_"`
		// CorsAllowedOrigins is a comma-separated list (e.g. "https://pentapulse.ru,http://localhost:5173").
		// Required for browser requests with credentials: the server must echo a concrete Origin, not "*".
		// If empty, any non-empty Origin is allowed (development only).
		CorsAllowedOrigins string `env:"CORS_ALLOWED_ORIGINS"`
		// Session cookie (set on login; clients may also send X-Session-Token).
		SessionCookieName     string `env:"SESSION_COOKIE_NAME"`      // default brainz_session
		SessionCookieDomain   string `env:"SESSION_COOKIE_DOMAIN"`    // e.g. .example.com
		SessionCookiePath     string `env:"SESSION_COOKIE_PATH"`      // default /
		SessionCookieMaxAge   int    `env:"SESSION_COOKIE_MAX_AGE"`    // seconds; 0 = 14 days
		SessionCookieSecure   bool   `env:"SESSION_COOKIE_SECURE"`    // set true behind HTTPS
		SessionCookieSameSite string `env:"SESSION_COOKIE_SAMESITE"` // lax, strict, none (default lax)
	}

	AppConfig struct {
		Address     string `env:"ADDRESS, required"`
		Port        int    `env:"PORT, required"`
		Environment string `env:"ENV"`
	}

	AuthConfig struct {
		Pepper        string `env:"PEPPER, required"`
		Argon2Memory  int    `env:"ARGON2_MEMORY, required"`
		Argon2Time    int    `env:"ARGON2_TIME, required"`
		Argon2Threads int    `env:"ARGON2_THREADS, required"`
	}

	PostgresConfig struct {
		Host     string `env:"HOST, required"`
		Port     string `env:"PORT, required"`
		User     string `env:"USER, required"`
		Name     string `env:"NAME, required"`
		Password string `env:"PASSWORD, required"`
		PoolMax  int32  `env:"POOL_MAX,required"`
	}

	RedisConfig struct {
		Host     string `env:"HOST,required"`
		Password string `env:"PASSWORD,required"`
	}
)

func MustLoad() *Config {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		panic("cannot read env: " + err.Error())
	}
	return &cfg
}

// CorsOriginSet returns trimmed, non-empty origins from CorsAllowedOrigins.
func (c *Config) CorsOriginSet() map[string]bool {
	if strings.TrimSpace(c.CorsAllowedOrigins) == "" {
		return nil
	}
	out := make(map[string]bool)
	for _, o := range strings.Split(c.CorsAllowedOrigins, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			out[o] = true
		}
	}
	return out
}
