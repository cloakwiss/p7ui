package src

import (
	"fmt"
	"time"
)

type LogLevel string

// const ConsoleId string = "console-output"

const (
	LevelInfo  LogLevel = "INFO"
	LevelError LogLevel = "ERROR"
	LevelFatal LogLevel = "FATAL"
	LevelDebug LogLevel = "DEBUG"
)

type Logger struct {
	logChan chan<- LogLine
	close   <-chan struct{}
}

func NewLogger(logChan chan<- LogLine, control <-chan struct{}) Logger {
	return Logger{logChan, control}
}

func (l *Logger) log(level LogLevel, msg string, args ...any) {
	select {
	case <-l.close:
		return
	default:
		timestamp := time.Now().Format("15:04:05")
		fullMsg := fmt.Sprintf(msg, args...)
		line := fmt.Sprintf("\n[%s] %-5s: %s\n", timestamp, level, fullMsg)

		fmt.Print(line)
		// hope this mitigates go routine leak
		l.logChan <- NewLogLine(timestamp, level, fullMsg)
	}
}

func (l *Logger) Info(msg string, args ...any) {
	l.log(LevelInfo, msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.log(LevelError, msg, args...)
}

func (l *Logger) Fatal(msg string, args ...any) {
	l.log(LevelFatal, msg, args...)
}

func (l *Logger) Debug(msg string, args ...any) {
	l.log(LevelDebug, msg, args...)
}
