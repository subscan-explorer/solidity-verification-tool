package util

import (
	"io"
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

func SetLogger(output io.Writer) {
	defaultLogger = &MLogger{
		infoLogger:    log.New(output, "INFO: ", 0),
		warningLogger: log.New(output, "WARNING: ", 0),
		errorLogger:   log.New(output, "ERROR: ", 0),
		debugLogger:   log.New(output, "DEBUG: ", 0),
	}
}

type MLogger struct {
	infoLogger    *log.Logger
	warningLogger *log.Logger
	errorLogger   *log.Logger
	debugLogger   *log.Logger
}

func NewLogger() *MLogger {
	return &MLogger{
		infoLogger:    log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		warningLogger: log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Llongfile),
		errorLogger:   log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Llongfile),
		debugLogger:   log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Llongfile),
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

func (l *MLogger) Error(err error) {
	if err == nil {
		return
	}
	l.logWithCallerDepth(l.errorLogger, 3, err.Error())
}

func (l *MLogger) Debug(msg string) {
	l.logWithCallerDepth(l.debugLogger, 3, msg)
}
