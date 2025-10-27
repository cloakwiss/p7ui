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

func (l *Logger) log(level LogLevel, formatString string, args ...any) {
	timestamp := time.Now().Format("15:04:05")

	fullString := fmt.Sprintf(formatString, args...)
	line := fmt.Sprintf("\n[%s] %-5s: %s\n", timestamp, level, fullString)
	fmt.Print(line)
	l.logChan <- NewLogLine(timestamp, level, fullString)
}

func (l *Logger) Info(formatString string, args ...any) {
	l.log(LevelInfo, formatString, args...)
}

func (l *Logger) Error(formatString string, args ...any) {
	l.log(LevelError, formatString, args...)
}

func (l *Logger) Fatal(formatString string, args ...any) {
	l.log(LevelFatal, formatString, args...)
}

func (l *Logger) Debug(formatString string, args ...any) {
	l.log(LevelDebug, formatString, args...)
}
