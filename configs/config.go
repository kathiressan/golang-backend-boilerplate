package config

import (
	"fmt"
	"os"
	"strconv"
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

	// Timeouts
	ReadTimeoutSeconds  int
	WriteTimeoutSeconds int
	IdleTimeoutSeconds  int
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

	return &Config{
		// Core settings
		GinMode: ginMode,
		Port: port,
		Environment: Environment(environment),
		TestMode: testMode,

		// Timeouts
		ReadTimeoutSeconds:  readTimeout,
		WriteTimeoutSeconds: writeTimeout,
		IdleTimeoutSeconds:  idleTimeout,
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