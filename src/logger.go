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

func (l *Logger) log(level LogLevel, msg string, payload error) {
	timestamp := time.Now().Format("15:04:05")

	// hope this mitigates go routine leak
	if payload == nil {
		line := fmt.Sprintf("\n[%s] %-5s: %s\n", timestamp, level, msg)
		fmt.Print(line)
		l.logChan <- NewLogLine(timestamp, level, msg)
	} else {
		fullMsg := fmt.Sprintf(msg, payload)
		line := fmt.Sprintf("\n[%s] %-5s: %s\n", timestamp, level, fullMsg)
		fmt.Print(line)
		l.logChan <- NewLogLineWithPayload(timestamp, level, msg, payload.Error())
	}
}

func (l *Logger) Info(msg string) {
	l.log(LevelInfo, msg, nil)
}

func (l *Logger) Error(msg string) {
	l.log(LevelError, msg, nil)
}

func (l *Logger) Fatal(msg string) {
	l.log(LevelFatal, msg, nil)
}

func (l *Logger) Debug(msg string) {
	l.log(LevelDebug, msg, nil)
}

func (l *Logger) InfoWithPayload(msg string, payload error) {
	l.log(LevelInfo, msg, payload)
}

func (l *Logger) ErrorWithPayload(msg string, payload error) {
	l.log(LevelError, msg, payload)
}

func (l *Logger) FatalWithPayload(msg string, payload error) {
	l.log(LevelFatal, msg, payload)
}

func (l *Logger) DebugWithPayload(msg string, payload error) {
	l.log(LevelDebug, msg, payload)
}
