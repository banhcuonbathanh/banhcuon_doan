package logger

import (
	"log"
	"os"
)

// Logger levels
const (
	InfoLevel = iota
	WarningLevel
	ErrorLevel
	FatalLevel
)

// Logger structure
type Logger struct {
	infoLogger    *log.Logger
	warningLogger *log.Logger
	errorLogger   *log.Logger
	fatalLogger   *log.Logger
}

// NewLogger creates a new Logger instance
func NewLogger() *Logger {
	return &Logger{
		infoLogger:    log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		warningLogger: log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger:   log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		fatalLogger:   log.New(os.Stderr, "FATAL: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// Log logs a message with the specified severity level
func (l *Logger) Log(level int, message string) {
	switch level {
	case InfoLevel:
		l.infoLogger.Println(message)
	case WarningLevel:
		l.warningLogger.Println(message)
	case ErrorLevel:
		l.errorLogger.Println(message)
	case FatalLevel:
		l.fatalLogger.Fatalln(message)
	default:
		l.infoLogger.Println(message)
	}
}

// Info logs an info message
func (l *Logger) Info(message string) {
	l.Log(InfoLevel, message)
}

// Warning logs a warning message
func (l *Logger) Warning(message string) {
	l.Log(WarningLevel, message)
}

// Error logs an error message
func (l *Logger) Error(message string) {
	l.Log(ErrorLevel, message)
}

// Fatal logs a fatal message and exits the program
func (l *Logger) Fatal(message string) {
	l.Log(FatalLevel, message)
}
