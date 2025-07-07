package discordant

import (
	"log"
	"os"
)

// Logger is implemented by any logging system that is used for standard logs.
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})

	Debug(args ...interface{})
	Info(args ...interface{})
	Warning(args ...interface{})
	Error(args ...interface{})

	Debugln(args ...interface{})
	Infoln(args ...interface{})
	Warningln(args ...interface{})
	Errorln(args ...interface{})
}

type defaultLog struct {
	*log.Logger
}

// NewDefaultLog creates and returns default logger to stderr.
func NewDefaultLog() Logger {
	return &defaultLog{Logger: log.New(os.Stderr, "discordant ", log.LstdFlags)}
}

func (l *defaultLog) Debugf(f string, v ...interface{}) {
	l.printf("DEBUG: "+f, v...)
}

func (l *defaultLog) Debug(args ...interface{}) {
	l.printf("DEBUG: %v", args...)
}

func (l *defaultLog) Debugln(args ...interface{}) {
	l.printf("DEBUG: %v\n", args...)
}

func (l *defaultLog) Infof(f string, v ...interface{}) {
	l.printf("INFO: "+f, v...)
}

func (l *defaultLog) Info(args ...interface{}) {
	l.printf("INFO: %v", args...)
}

func (l *defaultLog) Infoln(args ...interface{}) {
	l.printf("INFO: %v\n", args...)
}

func (l *defaultLog) Warningf(f string, v ...interface{}) {
	l.printf("WARNING: "+f, v...)
}

func (l *defaultLog) Warning(args ...interface{}) {
	l.printf("WARNING: %v", args...)
}

func (l *defaultLog) Warningln(args ...interface{}) {
	l.printf("WARNING: %v\n", args...)
}

func (l *defaultLog) Errorf(f string, v ...interface{}) {
	l.printf("ERROR: "+f, v...)
}

func (l *defaultLog) Error(args ...interface{}) {
	l.printf("ERROR: %v", args...)
}

func (l *defaultLog) Errorln(args ...interface{}) {
	l.printf("ERROR: %v\n", args...)
}

func (l *defaultLog) printf(f string, v ...interface{}) {
	l.Logger.Printf(f, v...)
}
