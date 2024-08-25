package logger

import (
	"fmt"
	"log"
	"os"
)

type Logger struct {
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		infoLogger:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		warnLogger:  log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}
func (l *Logger) Info(msg string, args ...interface{}) {
	l.infoLogger.Output(2, fmt.Sprintf(msg, args...))
}

func (l *Logger) Warn(msg string, args ...interface{}) {
	l.warnLogger.Output(2, fmt.Sprintf(msg, args...))
}

func (l *Logger) Error(msg string, args ...interface{}) {
	l.errorLogger.Output(2, fmt.Sprintf(msg, args...))
}
