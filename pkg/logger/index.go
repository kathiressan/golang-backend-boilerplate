// Package logger provides a structured logging system for the application
// This package wraps the zap logging library to provide consistent,
// high-performance logging across the application
package logger

import (
	"fmt"
	"net/http"
	"os"
	config "ovmsa-be/configs"
	"runtime"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// The main logger instance - used internally
	logger *zap.Logger

	// Sugar adds convenience methods to the logger
	// It's a bit slower but much more convenient to use
	sugar *zap.SugaredLogger

	once sync.Once
)

// ANSI color codes
const (
	colorReset     = "\033[0m"
	colorRed       = "\033[31m"
	colorGreen     = "\033[32m"
	colorYellow    = "\033[33m"
	colorBlue      = "\033[34m"
	colorPurple    = "\033[35m"
	colorCyan      = "\033[36m"
	colorGray      = "\033[37m"
	colorWhite     = "\033[97m"
	colorBold      = "\033[1m"
	colorDim       = "\033[2m"
	colorItalic    = "\033[3m"
	colorUnderline = "\033[4m"
)

// customCallerEncoder colors the caller in bold cyan
func customCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(fmt.Sprintf("%s%s%s%s", colorCyan, colorBold, caller.TrimmedPath(), colorReset))
}

// customTimeEncoder colors the timestamp in dim gray
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(fmt.Sprintf("%s%s%s", colorDim, t.Format("2006-01-02 15:04:05.000"), colorReset))
}

// addSeparator adds a visual separator for ERROR and FATAL logs
func addSeparator(level zapcore.Level) {
	if level == zapcore.ErrorLevel || level == zapcore.FatalLevel {
		fmt.Println("\n" + colorRed + colorBold + "==================== 🚨 ERROR 🚨 ====================" + colorReset)
	}
}

// isErrorStatus checks if the HTTP status code indicates an error
func isErrorStatus(status int) bool {
	return status >= 400
}

// formatStatus uses http.StatusText for the definition
func formatStatus(status int) string {
	var prefix string
	if status >= 500 {
		prefix = "Server Error"
	} else if status >= 400 {
		prefix = "Client Error"
	} else if status >= 300 {
		prefix = "Redirect"
	} else if status >= 200 {
		prefix = "Success"
	} else {
		prefix = "Informational"
	}
	def := http.StatusText(status)
	if def == "" {
		def = "Unknown Status"
	}
	return fmt.Sprintf("%d %s (%s)", status, prefix, def)
}

// formatField returns the value as plain text (no color)
func formatField(value any) string {
	return fmt.Sprintf("%v", value)
}

// getCodeLine returns the source code line at the caller, or "" if not available
func getCodeLine(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	if line-1 < 0 || line-1 >= len(lines) {
		return ""
	}
	return strings.TrimSpace(lines[line-1])
}

// customLevelEncoder makes the log level very prominent
func customLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var color string
	switch l {
	case zapcore.DebugLevel:
		color = colorBlue
	case zapcore.InfoLevel:
		color = colorGreen
	case zapcore.WarnLevel:
		color = colorYellow
	case zapcore.ErrorLevel:
		color = colorRed
	case zapcore.FatalLevel:
		color = colorPurple
	default:
		color = colorWhite
	}
	enc.AppendString(fmt.Sprintf("%s%s%s%s", color, colorBold, l.CapitalString(), colorReset))
}

// Initialize sets up the logger with proper configuration based on environment
// This function configures the logger differently for production vs development:
// - Production: JSON format, info level and above
// - Development: Console format with colors, debug level and above
func Initialize(environment config.Environment) {
	once.Do(func() {
		env := string(environment)

		// Setup encoder config to define log format
		// This defines how log entries are structured and formatted
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey = "timestamp"          // Field name for timestamps
		encoderConfig.EncodeTime = customTimeEncoder // Custom time encoder
		encoderConfig.StacktraceKey = "stacktrace"   // Field name for stack traces
		encoderConfig.MessageKey = "message"         // Field name for log messages
		encoderConfig.LevelKey = "level"             // Field name for log levels
		encoderConfig.CallerKey = "caller"           // Field name for caller information
		encoderConfig.NameKey = "logger"             // Field name for logger name
		encoderConfig.EncodeLevel = customLevelEncoder
		encoderConfig.EncodeCaller = customCallerEncoder
		encoderConfig.EncodeDuration = zapcore.StringDurationEncoder
		encoderConfig.EncodeName = zapcore.FullNameEncoder

		var encoder zapcore.Encoder
		var logLevel zapcore.Level

		// Configure logger differently based on environment
		// Production uses JSON format for better machine parsing
		// Development uses console format with colors for readability
		if env == string(config.EnvProduction) {
			// JSON format is better for production as it's easily parsed by log analyzers
			encoder = zapcore.NewJSONEncoder(encoderConfig)
			// Info level is less verbose than debug, better for production
			logLevel = zap.InfoLevel
		} else {
			// Custom console encoder for development with colors
			encoder = zapcore.NewConsoleEncoder(encoderConfig)
			// Debug level shows more information, good for development
			logLevel = zap.DebugLevel
		}

		// Create a multi-writer to output to both console and file
		output := zapcore.AddSync(os.Stdout)

		// Create the core logger with our configuration
		// This sets up the encoder, output, and minimum log level
		core := zapcore.NewCore(
			encoder,  // How to format logs
			output,   // Where to send logs (standard output)
			logLevel, // Minimum log level to record
		)

		// Create the final logger with additional options
		logger = zap.New(core,
			zap.AddCaller(),                   // Include the calling function in logs
			zap.AddCallerSkip(1),              // Skip one level of caller for better readability
			zap.AddStacktrace(zap.ErrorLevel), // Add stack traces for errors and above
			zap.Development(),                 // Enable development mode for more detailed logs
		)

		// Create the sugared logger for easier usage
		// The sugared logger is slower but has a more convenient API
		sugar = logger.Sugar()
	})
}

// Debug logs at debug level
// Debug logs are very detailed and typically only used during development
// Example: Debug("User login attempt", "username", "john.doe", "attempt", 3)
func Debug(msg string, fields ...any) {
	sugar.Debugw(msg, fields...)
}

// Info logs at info level with special handling for HTTP status codes
func Info(msg string, fields ...any) {
	// Check if this is a request completion log
	if strings.Contains(msg, "Request completed") {
		// Format all fields with appropriate colors
		for i := 0; i < len(fields); i += 2 {
			if i+1 < len(fields) {
				key := fields[i].(string)
				if key == "status" {
					if status, ok := fields[i+1].(int); ok {
						if isErrorStatus(status) {
							fields[i+1] = formatStatus(status)
							sugar.Errorw(msg, fields...)
							return
						} else {
							fields[i+1] = formatStatus(status)
						}
					}
				} else {
					fields[i+1] = formatField(fields[i+1])
				}
			}
		}
	}

	// If !config.IsProduction(), wrap the msg with colorWhite+colorBold+msg+colorReset
	if !config.IsProduction() {
		msg = fmt.Sprintf("%s%s%s%s", colorWhite, colorBold, msg, colorReset)
	}
	sugar.Infow(msg, fields...)
}

// Infof logs a templated message
func Infof(template string, args ...any) {
	msg := fmt.Sprintf(template, args...)

	// Maintain your existing color logic
	if !config.IsProduction() {
		msg = fmt.Sprintf("%s%s%s%s", colorWhite, colorBold, msg, colorReset)
	}

	sugar.Info(msg)
}

// Warn logs at warn level
// Warn logs indicate potential issues that aren't errors
// Example: Warn("High API usage detected", "requestsPerMinute", 1200, "threshold", 1000)
func Warn(msg string, fields ...any) {
	sugar.Warnw(msg, fields...)
}

// Error logs at error level with enhanced stack trace and code line
func Error(msg string, fields ...any) {
	// Add error context if available
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) && fields[i] == "error" {
			if err, ok := fields[i+1].(error); ok {
				// Add error message to the main message
				msg = fmt.Sprintf("%s: %s", msg, err.Error())
			}
		}
	}

	// Try to get the code line (skip=2: Error -> caller)
	if codeLine := getCodeLine(2); codeLine != "" {
		msg = fmt.Sprintf("%s | code: %s", msg, codeLine)
	}

	addSeparator(zapcore.ErrorLevel)

	// If !config.IsProduction(), wrap the msg with colorWhite+colorBold+msg+colorReset
	if !config.IsProduction() {
		msg = fmt.Sprintf("%s%s%s%s", colorWhite, colorBold, msg, colorReset)
	}
	sugar.Errorw(msg, fields...)
}

// Fatal logs at fatal level then calls os.Exit(1)
// Fatal logs indicate a serious error that requires shutting down the application
// Example: Fatal("Failed to initialize database", "error", err)
func Fatal(msg string, fields ...any) {
	addSeparator(zapcore.FatalLevel)

	// If !config.IsProduction(), wrap the msg with colorWhite+colorBold+msg+colorReset
	if !config.IsProduction() {
		msg = fmt.Sprintf("%s%s%s%s", colorWhite, colorBold, msg, colorReset)
	}
	sugar.Fatalw(msg, fields...)
}

// With adds context fields to the logger
// This is useful for creating loggers with preset fields
// Example: userLogger := With("userId", 12345)
//
//	userLogger.Info("Profile updated")
func With(fields ...any) *zap.SugaredLogger {
	return sugar.With(fields...)
}

// Sync flushes any buffered log entries - should be called before program exit
// This ensures all logs are written before the application terminates
// It's typically called in defer statements or shutdown hooks
func Sync() {
	if logger != nil {
		_ = logger.Sync()
	}
}

// How to use this logger:
//
// 1. Structured logging:
//    logger.Info("User registered",
//        "userId", 12345,
//        "email", "user@example.com",
//        "plan", "premium")
//
// 2. Error logging:
//    if err != nil {
//        logger.Error("Failed to update database",
//            "error", err,
//            "userId", 12345)
//    }
//
// 3. Creating a contextual logger:
//    requestLogger := logger.With(
//        "requestId", requestId,
//        "userId", userId)
//
//    requestLogger.Info("Request started")
//    // ... process request ...
//    requestLogger.Info("Request completed")

// Folow the following pattern in code logic to print the error message and stack trace:
// if err != nil {
//     logger.Error("Failed to do something", "error", err)
// }
