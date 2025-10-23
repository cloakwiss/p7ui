package src

import (
	"fmt"
	"time"
)

type LogLevel string

const ConsoleId string = "console-output"

const (
	LevelInfo  LogLevel = "INFO"
	LevelError LogLevel = "ERROR"
	LevelFatal LogLevel = "FATAL"
	LevelDebug LogLevel = "DEBUG"
)

type Logger struct {
	logChan chan<- string
	close   <-chan struct{}
}

func NewLogger(logChan chan<- string, control <-chan struct{}) Logger {
	return Logger{logChan, control}
}

func (l *Logger) log(level LogLevel, msg string, args ...any) {
	timestamp := time.Now().Format("15:04:05")
	line := fmt.Sprintf("\n[%s] %-5s: %s\n", timestamp, level, fmt.Sprintf(msg, args...))

	fmt.Print(line)
	// hope this mitigates go routine leak
	if _, ok := <-l.close; ok {
		l.logChan <- line
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
