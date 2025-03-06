package util

import (
	"log"
	"os"
)

var defaultLogger *MLogger

func Logger() *MLogger {
	return defaultLogger
}

func init() {
	defaultLogger = NewLogger()
}

type MLogger struct {
	infoLogger    *log.Logger
	warningLogger *log.Logger
	errorLogger   *log.Logger
}

func NewLogger() *MLogger {
	return &MLogger{
		infoLogger:    log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		warningLogger: log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Llongfile),
		errorLogger:   log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Llongfile),
	}
}

func (l *MLogger) logWithCallerDepth(logger *log.Logger, calldepth int, msg string) {
	_ = logger.Output(calldepth, msg)
}

func (l *MLogger) Info(msg string) {
	l.logWithCallerDepth(l.infoLogger, 3, msg)
}

func (l *MLogger) Warning(msg string) {
	l.logWithCallerDepth(l.warningLogger, 3, msg)
}

func (l *MLogger) Error(msg string) {
	l.logWithCallerDepth(l.errorLogger, 3, msg)
}
