package request_marker

import (
	"log"
	"os"
)

// LogLevel represents the logging level
type LogLevel int

const (
	LogLevelDebug = LogLevel(3)
	LogLevelInfo  = LogLevel(2)
	LogLevelError = LogLevel(1)
)

// Logger provides structured logging for the request marker plugin.
// Note: Traefik disables unsafe and syscall, which makes many logger libraries unusable.
// This simple logger implementation avoids those restrictions.
// Reference: https://github.com/tomMoulard/fail2ban/blob/main/fail2ban.go#L35-L38
type Logger struct {
	level       LogLevel
	infoLogger  *log.Logger
	debugLogger *log.Logger
	errorLogger *log.Logger
}

// Debug logs a debug-level message
func (l *Logger) Debug(args ...interface{}) {
	if l.level >= LogLevelDebug {
		l.debugLogger.Println(args...)
	}
}

// Info logs an info-level message
func (l *Logger) Info(args ...interface{}) {
	if l.level >= LogLevelInfo {
		l.infoLogger.Println(args...)
	}
}

// Error logs an error-level message
func (l *Logger) Error(args ...interface{}) {
	if l.level >= LogLevelError {
		l.errorLogger.Println(args...)
	}
}

func parseLogLevel(level string) LogLevel {
	switch level {
	case "DEBUG":
		return LogLevelDebug
	case "INFO":
		return LogLevelInfo
	case "ERROR":
		return LogLevelError
	default:
		return LogLevelError
	}
}

// NewLogger creates a new logger with the specified level
func NewLogger(level string) *Logger {
	return &Logger{
		level:       parseLogLevel(level),
		infoLogger:  log.New(os.Stdout, "INFO: [REQUEST_MARK] ", log.Ldate|log.Ltime|log.Lshortfile),
		debugLogger: log.New(os.Stdout, "DEBUG: [REQUEST_MARK] ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger: log.New(os.Stdout, "ERROR: [REQUEST_MARK] ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}
