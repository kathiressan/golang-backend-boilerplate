package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

type Environment string

const (
	EnvDevelopment Environment = "development"
	EnvStaging     Environment = "staging"
	EnvProduction  Environment = "production"
)

type Config struct {
	// Core settings
	GinMode      string
	Port         int
	Environment  Environment
	TestMode     bool
	AppName      string
	ProjectRoot  string
	PlatformName string

	// Timeouts
	ReadTimeoutSeconds  int
	WriteTimeoutSeconds int
	IdleTimeoutSeconds  int

	// Security settings
	JWTSigningMethod         string
	JWTSecret                string
	JWTPrivateKey            string
	JWTPublicKey             string
	AccessTokenExpiryMinutes int
	RefreshTokenExpiryDays   int
	EnforceAudienceValidation bool
	RateLimitEnabled          bool
	RateLimitRequestsPerMinute int
	TrustedProxies           []string
	AllowedOrigins           []string

	// Database settings
	DBHost      string
	DBPort      int
	DBUser      string
	DBPassword  string
	DBName      string
	DBSSLMode   string
	DatabaseURL string
}

var (
	config *Config
	once   sync.Once
)

func loadConfig() (*Config, error) {
	// Try to load .env file
	_ = godotenv.Load()

	// GIN_MODE
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = "debug" // Default to debug mode
	}
	if ginMode != "debug" && ginMode != "release" && ginMode != "test" {
		return nil, fmt.Errorf("invalid GIN_MODE value: %s. Must be one of: debug, release, test", ginMode)
	}

	// PORT
	portStr := os.Getenv("PORT")
	if portStr == "" {
		portStr = "8000" // Default port
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid PORT value: %s. Must be a number", portStr)
	}
	if port < 1 || port > 65535 {
		return nil, fmt.Errorf("invalid PORT value %d. Must be between 1 and 65535", port)
	}

	// ENVIRONMENT
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = string(EnvDevelopment)
	}
	if environment != string(EnvDevelopment) && environment != string(EnvStaging) && environment != string(EnvProduction) {
		return nil, fmt.Errorf("invalid ENVIRONMENT value: %s. Must be one of: development, staging, production", environment)
	}

	// TEST_MODE
	testMode := os.Getenv("TEST_MODE") == "true"

	// APP_NAME
	appName := os.Getenv("APP_NAME")
	if appName == "" {
		if ginMode == "test" {
			appName = "TestApp"
		} else {
			return nil, fmt.Errorf("APP_NAME cannot be empty")
		}
	}

	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		if ginMode == "test" {
			projectRoot = "."
		} else {
			return nil, fmt.Errorf("PROJECT_ROOT cannot be empty")
		}
	}

	platformName := os.Getenv("PLATFORM_NAME")
	if platformName == "" {
		if ginMode == "test" {
			platformName = "test-platform"
		} else {
			return nil, fmt.Errorf("PLATFORM_NAME cannot be empty")
		}
	}

	// Load timeout settings
	readTimeoutStr := os.Getenv("READ_TIMEOUT_SECONDS")
	readTimeout := 10 // Default to 10 seconds
	if readTimeoutStr != "" {
		readTimeout, err = strconv.Atoi(readTimeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid READ_TIMEOUT_SECONDS value: %s. Must be a number", readTimeoutStr)
		}
	}

	writeTimeoutStr := os.Getenv("WRITE_TIMEOUT_SECONDS")
	writeTimeout := 10 // Default to 10 seconds
	if writeTimeoutStr != "" {
		writeTimeout, err = strconv.Atoi(writeTimeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid WRITE_TIMEOUT_SECONDS value: %s. Must be a number", writeTimeoutStr)
		}
	}

	idleTimeoutStr := os.Getenv("IDLE_TIMEOUT_SECONDS")
	idleTimeout := 120 // Default to 120 seconds
	if idleTimeoutStr != "" {
		idleTimeout, err = strconv.Atoi(idleTimeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid IDLE_TIMEOUT_SECONDS value: %s. Must be a number", idleTimeoutStr)
		}
	}

	// Load JWT settings
	jwtSigningMethod := os.Getenv("JWT_SIGNING_METHOD")
	if jwtSigningMethod == "" {
		jwtSigningMethod = "HS256" // Default to HS256
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSigningMethod == "HS256" {
		if jwtSecret == "" && environment == "production" {
			return nil, fmt.Errorf("JWT_SECRET must be set in production for HS256")
		}
		if jwtSecret == "" {
			jwtSecret = "default-jwt-secret-for-dev-only"
		}
	}

	jwtPrivateKey := os.Getenv("JWT_PRIVATE_KEY")
	jwtPublicKey := os.Getenv("JWT_PUBLIC_KEY")
	if jwtSigningMethod == "RS256" {
		if (jwtPrivateKey == "" || jwtPublicKey == "") && environment == "production" {
			return nil, fmt.Errorf("JWT_PRIVATE_KEY and JWT_PUBLIC_KEY must be set in production for RS256")
		}
	}

	accessTokenExpiryStr := os.Getenv("ACCESS_TOKEN_EXPIRY_MINUTES")
	accessTokenExpiryMinutes := 60 // Default to 60 minutes
	if accessTokenExpiryStr != "" {
		accessTokenExpiryMinutes, err = strconv.Atoi(accessTokenExpiryStr)
		if err != nil {
			return nil, fmt.Errorf("invalid ACCESS_TOKEN_EXPIRY_MINUTES value: %s", accessTokenExpiryStr)
		}
	}

	refreshTokenExpiryStr := os.Getenv("REFRESH_TOKEN_EXPIRY_DAYS")
	refreshTokenExpiryDays := 30 // Default to 30 days
	if refreshTokenExpiryStr != "" {
		refreshTokenExpiryDays, err = strconv.Atoi(refreshTokenExpiryStr)
		if err != nil {
			return nil, fmt.Errorf("invalid REFRESH_TOKEN_EXPIRY_DAYS value: %s", refreshTokenExpiryStr)
		}
	}

	// Audience validation (default: false for backward compatibility)
	enforceAudienceValidation := os.Getenv("ENFORCE_AUDIENCE_VALIDATION") == "true"

	// Rate limiting settings
	rateLimitEnabled := os.Getenv("RATE_LIMIT_ENABLED") == "true"
	rateLimitRequestsPerMinute := 60 // Default to 60 requests per minute
	if rateLimitStr := os.Getenv("RATE_LIMIT_REQUESTS_PER_MINUTE"); rateLimitStr != "" {
		rateLimitRequestsPerMinute, err = strconv.Atoi(rateLimitStr)
		if err != nil {
			return nil, fmt.Errorf("invalid RATE_LIMIT_REQUESTS_PER_MINUTE value: %s", rateLimitStr)
		}
	}

	// Load CORS settings
	allowedOriginsStr := os.Getenv("ALLOWED_ORIGINS")
	var allowedOrigins []string
	if allowedOriginsStr == "" {
		if environment == "production" {
			return nil, fmt.Errorf("ALLOWED_ORIGINS must be set in production")
		}
		// Default for non-production
		allowedOrigins = []string{
			"http://localhost:3000",
			"http://localhost:8000",
			"http://localhost:8080",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:8000",
			"http://127.0.0.1:8080",
		}
	} else {
		allowedOrigins = strings.Split(allowedOriginsStr, ",")
	}

	// Load trusted proxies
	trustedProxiesStr := os.Getenv("TRUSTED_PROXIES")
	var trustedProxies []string
	if trustedProxiesStr != "" {
		trustedProxies = strings.Split(trustedProxiesStr, ",")
	}

	// Database settings
	dbURL := os.Getenv("DATABASE_URL")
	dbHost := os.Getenv("DB_HOST")
	dbPortStr := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbSSLMode := os.Getenv("DB_SSLMODE")

	// If no DATABASE_URL, we need individual components
	if dbURL == "" {
		if dbHost == "" {
			dbHost = "localhost"
		}
		if dbSSLMode == "" {
			dbSSLMode = "disable"
		}
		// In production, force-require the essentials if no URL is provided
		if environment == string(EnvProduction) {
			if dbUser == "" || dbName == "" {
				return nil, fmt.Errorf("either DATABASE_URL or (DB_USER and DB_NAME) must be set in production")
			}
		}
	}

	dbPort := 5432
	if dbPortStr != "" {
		dbPort, _ = strconv.Atoi(dbPortStr)
	}

	return &Config{
		// Core settings
		GinMode:      ginMode,
		Port:         port,
		Environment:  Environment(environment),
		TestMode:     testMode,
		AppName:      appName,
		ProjectRoot:  projectRoot,
		PlatformName: platformName,

		// Timeouts
		ReadTimeoutSeconds:  readTimeout,
		WriteTimeoutSeconds: writeTimeout,
		IdleTimeoutSeconds:  idleTimeout,

		// Security settings
		JWTSigningMethod:           jwtSigningMethod,
		JWTSecret:                  jwtSecret,
		JWTPrivateKey:              jwtPrivateKey,
		JWTPublicKey:               jwtPublicKey,
		AccessTokenExpiryMinutes:   accessTokenExpiryMinutes,
		RefreshTokenExpiryDays:     refreshTokenExpiryDays,
		EnforceAudienceValidation:  enforceAudienceValidation,
		RateLimitEnabled:           rateLimitEnabled,
		RateLimitRequestsPerMinute: rateLimitRequestsPerMinute,
		TrustedProxies:             trustedProxies,
		AllowedOrigins:             allowedOrigins,

		// Database
		DBHost:      dbHost,
		DBPort:      dbPort,
		DBUser:      dbUser,
		DBPassword:  dbPassword,
		DBName:      dbName,
		DBSSLMode:   dbSSLMode,
		DatabaseURL: dbURL,
	}, nil
}

func GetConfig() *Config {
	once.Do(func() {
		cfg, err := loadConfig()
		if err != nil {
			panic(fmt.Sprintf("Failed to load config: %v", err))
		}
		config = cfg
	})
	return config
}

func IsProduction() bool {
	return GetConfig().Environment == "production"
}

// ResetConfigForTest clears the singleton config, allowing it to be reloaded.
// This should ONLY be used in unit tests.
func ResetConfigForTest() {
	config = nil
	once = sync.Once{}
}
