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
	GinMode     string
	Port        int
	Environment Environment
	TestMode    bool
	AppName     string
	ProjectRoot string
	PlatformName string

	// Timeouts
	ReadTimeoutSeconds  int
	WriteTimeoutSeconds int
	IdleTimeoutSeconds  int

	// Security settings
	JWTSecret      string
	JWTExpiryHours int
	TrustedProxies []string
	AllowedOrigins []string
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
		return nil, fmt.Errorf("APP_NAME cannot be empty")
	}

	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		return nil, fmt.Errorf("PROJECT_ROOT cannot be empty")
	}

	platformName := os.Getenv("PLATFORM_NAME")
	if platformName == "" {
		return nil, fmt.Errorf("PLATFORM_NAME cannot be empty")
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
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" && environment == "production" {
		return nil, fmt.Errorf("JWT_SECRET must be set in production")
	}
	if jwtSecret == "" {
		jwtSecret = "default-jwt-secret-for-dev-only" // Default for non-production
	}

	jwtExpiryStr := os.Getenv("JWT_EXPIRY_HOURS")
	jwtExpiryHours := 24 // Default to 24 hours
	if jwtExpiryStr != "" {
		jwtExpiryHours, err = strconv.Atoi(jwtExpiryStr)
		if err != nil {
			return nil, fmt.Errorf("invalid JWT_EXPIRY_HOURS value: %s. Must be a number", jwtExpiryStr)
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

	return &Config{
		// Core settings
		GinMode: ginMode,
		Port: port,
		Environment: Environment(environment),
		TestMode: testMode,
		AppName: appName,
		ProjectRoot: projectRoot,
		PlatformName: platformName,

		// Timeouts
		ReadTimeoutSeconds:  readTimeout,
		WriteTimeoutSeconds: writeTimeout,
		IdleTimeoutSeconds:  idleTimeout,

		// Security settings
		JWTSecret:      jwtSecret,
		JWTExpiryHours: jwtExpiryHours,
		TrustedProxies: trustedProxies,
		AllowedOrigins: allowedOrigins,
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