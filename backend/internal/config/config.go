package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	// API Keys
	GrokAPIKey string

	// External APIs
	FederalRegisterAPIURL string
	GrokAPIURL            string
	GrokModel             string

	// Database
	DatabaseURLEnv string // Direct URL from DB_URL env var
	DatabaseHost   string
	DatabasePort   string
	DatabaseUser   string
	DatabasePass   string
	DatabaseName   string
	DatabaseSSL    string

	// Scraper settings
	ScraperIntervalMinutes int
	ScraperDaysLookback    int

	// CORS
	CORSEnabled    bool
	AllowedOrigins []string

	// Timeouts (seconds)
	FederalRegisterTimeout int
	GrokTimeout            int

	// Limits
	MaxRequestSizeBytes     int
	FederalRegisterPerPage  int
	FederalRegisterMaxPages int

	// Environment
	Debug       bool
	Environment string
	BehindProxy bool
	UseMockGrok bool
	Port        string

	// Authentication Security
	CookieSecure bool

	// Google OAuth
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURI  string

	// JWT
	JWTSecretKey            string
	JWTAlgorithm            string
	JWTAccessTokenExpireMin int

	// Frontend URL
	FrontendURL string
}

func parseBool(v string) bool {
	l := strings.ToLower(strings.TrimSpace(v))
	return l == "true" || l == "1" || l == "t" || l == "yes"
}

func Load() (*Config, error) {
	c := &Config{
		// Defaults
		FederalRegisterAPIURL:   "https://www.federalregister.gov/api/v1",
		GrokAPIURL:              "https://api.x.ai/v1",
		ScraperIntervalMinutes:  15,
		ScraperDaysLookback:     1,
		CORSEnabled:             true,
		AllowedOrigins:          []string{"http://localhost:5173", "http://localhost:3000"},
		FederalRegisterTimeout:  30,
		GrokTimeout:             60,
		MaxRequestSizeBytes:     10 * 1024 * 1024, // 10 MB
		FederalRegisterPerPage:  100,
		FederalRegisterMaxPages: 2,
		Debug:                   false,
		Environment:             "development",
		BehindProxy:             false,
		UseMockGrok:             false,
		CookieSecure:            false,
		JWTAlgorithm:            "HS256",
		JWTAccessTokenExpireMin: 60,
		FrontendURL:             "http://localhost:5173",
		GrokModel:               "grok-4-1-fast-non-reasoning",
		Port:                    "8000",
	}

	// Override with environment variables
	if v := os.Getenv("GROK_API_KEY"); v != "" {
		c.GrokAPIKey = v
	}

	if v := os.Getenv("GROK_API_URL"); v != "" {
		c.GrokAPIURL = v
	}

	if v := os.Getenv("FEDERAL_REGISTER_API_URL"); v != "" {
		c.FederalRegisterAPIURL = v
	}

	// Database URL (takes precedence if set)
	if v := os.Getenv("DB_URL"); v != "" {
		c.DatabaseURLEnv = v
	}

	// Database individual variables (used as fallback)
	if v := os.Getenv("DB_HOST"); v != "" {
		c.DatabaseHost = v
	} else {
		c.DatabaseHost = "localhost"
	}

	if v := os.Getenv("DB_PORT"); v != "" {
		c.DatabasePort = v
	} else {
		c.DatabasePort = "5432"
	}

	if v := os.Getenv("DB_USER"); v != "" {
		c.DatabaseUser = v
	} else {
		c.DatabaseUser = "postgres"
	}

	if v := os.Getenv("DB_PASSWORD"); v != "" {
		c.DatabasePass = v
	}

	if v := os.Getenv("DB_NAME"); v != "" {
		c.DatabaseName = v
	} else {
		c.DatabaseName = "opengov"
	}

	if v := os.Getenv("DB_SSLMODE"); v != "" {
		c.DatabaseSSL = v
	} else {
		c.DatabaseSSL = "disable"
	}

	if v := os.Getenv("SCRAPER_INTERVAL_MINUTES"); v != "" {
		if iv, err := strconv.Atoi(v); err == nil {
			c.ScraperIntervalMinutes = iv
		}
	}

	if v := os.Getenv("SCRAPER_DAYS_LOOKBACK"); v != "" {
		if iv, err := strconv.Atoi(v); err == nil {
			c.ScraperDaysLookback = iv
		}
	}

	if v := os.Getenv("CORS_ENABLED"); v != "" {
		c.CORSEnabled = parseBool(v)
	}

	if v := os.Getenv("ALLOWED_ORIGINS"); v != "" {
		c.AllowedOrigins = strings.Split(v, ",")
	}

	if v := os.Getenv("FEDERAL_REGISTER_TIMEOUT"); v != "" {
		if iv, err := strconv.Atoi(v); err == nil {
			c.FederalRegisterTimeout = iv
		}
	}

	if v := os.Getenv("GROK_TIMEOUT"); v != "" {
		if iv, err := strconv.Atoi(v); err == nil {
			c.GrokTimeout = iv
		}
	}

	if v := os.Getenv("MAX_REQUEST_SIZE_BYTES"); v != "" {
		if iv, err := strconv.Atoi(v); err == nil {
			c.MaxRequestSizeBytes = iv
		}
	}

	if v := os.Getenv("FEDERAL_REGISTER_PER_PAGE"); v != "" {
		if iv, err := strconv.Atoi(v); err == nil {
			c.FederalRegisterPerPage = iv
		}
	}

	if v := os.Getenv("FEDERAL_REGISTER_MAX_PAGES"); v != "" {
		if iv, err := strconv.Atoi(v); err == nil {
			c.FederalRegisterMaxPages = iv
		}
	}

	if v := os.Getenv("DEBUG"); v != "" {
		c.Debug = parseBool(v)
	}

	if v := os.Getenv("ENVIRONMENT"); v != "" {
		c.Environment = v
	}

	if v := os.Getenv("BEHIND_PROXY"); v != "" {
		c.BehindProxy = parseBool(v)
	}

	if v := os.Getenv("USE_MOCK_GROK"); v != "" {
		c.UseMockGrok = parseBool(v)
	}

	if v := os.Getenv("COOKIE_SECURE"); v != "" {
		c.CookieSecure = parseBool(v)
	}

	if v := os.Getenv("GOOGLE_CLIENT_ID"); v != "" {
		c.GoogleClientID = v
	}

	if v := os.Getenv("GOOGLE_CLIENT_SECRET"); v != "" {
		c.GoogleClientSecret = v
	}

	if v := os.Getenv("GOOGLE_REDIRECT_URI"); v != "" {
		c.GoogleRedirectURI = v
	}

	if v := os.Getenv("JWT_SECRET_KEY"); v != "" {
		c.JWTSecretKey = v
	} else if c.Environment == "development" {
		c.JWTSecretKey = "development-secret-key-change-in-production-32chars"
	} else {
		return nil, fmt.Errorf("JWT_SECRET_KEY is required")
	}

	if v := os.Getenv("JWT_ALGORITHM"); v != "" {
		c.JWTAlgorithm = v
	}

	if v := os.Getenv("JWT_ACCESS_TOKEN_EXPIRE_MINUTES"); v != "" {
		if iv, err := strconv.Atoi(v); err == nil {
			c.JWTAccessTokenExpireMin = iv
		}
	}

	if v := os.Getenv("FRONTEND_URL"); v != "" {
		c.FrontendURL = v
	}

	if v := os.Getenv("GROK_MODEL"); v != "" {
		c.GrokModel = v
	}

	if v := os.Getenv("PORT"); v != "" {
		c.Port = v
	}

	return c, nil
}

func (c *Config) DatabaseURL() string {
	// Use direct URL if provided
	if c.DatabaseURLEnv != "" {
		return c.DatabaseURLEnv
	}

	// Otherwise build from components
	host := c.DatabaseHost
	if c.DatabasePort != "" {
		host = net.JoinHostPort(c.DatabaseHost, c.DatabasePort)
	}

	u := &url.URL{
		Scheme: "postgres",
		Host:   host,
		Path:   "/" + c.DatabaseName,
	}

	if c.DatabasePass != "" {
		u.User = url.UserPassword(c.DatabaseUser, c.DatabasePass)
	} else {
		u.User = url.User(c.DatabaseUser)
	}

	q := u.Query()
	if c.DatabaseSSL != "" {
		q.Set("sslmode", c.DatabaseSSL)
	}
	u.RawQuery = q.Encode()

	return u.String()
}

func (c *Config) ScraperInterval() time.Duration {
	return time.Duration(c.ScraperIntervalMinutes) * time.Minute
}

func (c *Config) ValidateOAuth() bool {
	hasClientID := c.GoogleClientID != ""
	hasClientSecret := c.GoogleClientSecret != ""
	return hasClientID == hasClientSecret && hasClientID
}
