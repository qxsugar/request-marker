package request_marker

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestNewLogger_DebugLevel(t *testing.T) {
	logger := NewLogger("DEBUG")
	if logger.level != LogLevelDebug {
		t.Errorf("expected LogLevelDebug, got %v", logger.level)
	}
}

func TestNewLogger_InfoLevel(t *testing.T) {
	logger := NewLogger("INFO")
	if logger.level != LogLevelInfo {
		t.Errorf("expected LogLevelInfo, got %v", logger.level)
	}
}

func TestNewLogger_ErrorLevel(t *testing.T) {
	logger := NewLogger("ERROR")
	if logger.level != LogLevelError {
		t.Errorf("expected LogLevelError, got %v", logger.level)
	}
}

func TestNewLogger_InvalidLevel(t *testing.T) {
	logger := NewLogger("INVALID")
	if logger.level != LogLevelError {
		t.Errorf("expected LogLevelError for invalid level, got %v", logger.level)
	}
}

func TestParseLogLevel_Debug(t *testing.T) {
	level := parseLogLevel("DEBUG")
	if level != LogLevelDebug {
		t.Errorf("expected LogLevelDebug, got %v", level)
	}
}

func TestParseLogLevel_Info(t *testing.T) {
	level := parseLogLevel("INFO")
	if level != LogLevelInfo {
		t.Errorf("expected LogLevelInfo, got %v", level)
	}
}

func TestParseLogLevel_Error(t *testing.T) {
	level := parseLogLevel("ERROR")
	if level != LogLevelError {
		t.Errorf("expected LogLevelError, got %v", level)
	}
}

func TestParseLogLevel_Unknown(t *testing.T) {
	level := parseLogLevel("UNKNOWN")
	if level != LogLevelError {
		t.Errorf("expected LogLevelError for unknown level, got %v", level)
	}
}

func TestLogger_DebugOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := &Logger{
		level:       LogLevelDebug,
		debugLogger: log.New(buf, "DEBUG: ", 0),
		infoLogger:  log.New(buf, "INFO: ", 0),
		errorLogger: log.New(buf, "ERROR: ", 0),
	}

	logger.Debug("test debug message")
	output := buf.String()

	if !strings.Contains(output, "test debug message") {
		t.Errorf("expected debug message in output, got: %s", output)
	}
}

func TestLogger_InfoOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := &Logger{
		level:       LogLevelInfo,
		debugLogger: log.New(buf, "DEBUG: ", 0),
		infoLogger:  log.New(buf, "INFO: ", 0),
		errorLogger: log.New(buf, "ERROR: ", 0),
	}

	logger.Info("test info message")
	output := buf.String()

	if !strings.Contains(output, "test info message") {
		t.Errorf("expected info message in output, got: %s", output)
	}
}

func TestLogger_ErrorOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := &Logger{
		level:       LogLevelError,
		debugLogger: log.New(buf, "DEBUG: ", 0),
		infoLogger:  log.New(buf, "INFO: ", 0),
		errorLogger: log.New(buf, "ERROR: ", 0),
	}

	logger.Error("test error message")
	output := buf.String()

	if !strings.Contains(output, "test error message") {
		t.Errorf("expected error message in output, got: %s", output)
	}
}

func TestLogger_DebugLevelFiltering(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := &Logger{
		level:       LogLevelInfo,
		debugLogger: log.New(buf, "DEBUG: ", 0),
		infoLogger:  log.New(buf, "INFO: ", 0),
		errorLogger: log.New(buf, "ERROR: ", 0),
	}

	logger.Debug("debug message")
	logger.Info("info message")

	output := buf.String()

	if strings.Contains(output, "debug message") {
		t.Errorf("debug message should not appear at INFO level")
	}
	if !strings.Contains(output, "info message") {
		t.Errorf("info message should appear at INFO level")
	}
}

func TestLogger_ErrorLevelFiltering(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := &Logger{
		level:       LogLevelError,
		debugLogger: log.New(buf, "DEBUG: ", 0),
		infoLogger:  log.New(buf, "INFO: ", 0),
		errorLogger: log.New(buf, "ERROR: ", 0),
	}

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Error("error message")

	output := buf.String()

	if strings.Contains(output, "debug message") {
		t.Errorf("debug message should not appear at ERROR level")
	}
	if strings.Contains(output, "info message") {
		t.Errorf("info message should not appear at ERROR level")
	}
	if !strings.Contains(output, "error message") {
		t.Errorf("error message should appear at ERROR level")
	}
}

func TestLogger_MultipleMessages(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := &Logger{
		level:       LogLevelDebug,
		debugLogger: log.New(buf, "DEBUG: ", 0),
		infoLogger:  log.New(buf, "INFO: ", 0),
		errorLogger: log.New(buf, "ERROR: ", 0),
	}

	logger.Debug("message1")
	logger.Info("message2")
	logger.Error("message3")

	output := buf.String()

	if !strings.Contains(output, "message1") {
		t.Errorf("expected message1 in output")
	}
	if !strings.Contains(output, "message2") {
		t.Errorf("expected message2 in output")
	}
	if !strings.Contains(output, "message3") {
		t.Errorf("expected message3 in output")
	}
}

func TestLogLevelConstants(t *testing.T) {
	if LogLevelDebug <= LogLevelInfo {
		t.Errorf("LogLevelDebug should be greater than LogLevelInfo")
	}
	if LogLevelInfo <= LogLevelError {
		t.Errorf("LogLevelInfo should be greater than LogLevelError")
	}
}
