/*
Copyright © 2024 Montasser abed majid zehri <montasser.zehri@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package logger

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

type ILogger interface {
	Debug(message interface{}, args ...interface{})
	Info(message string, args ...interface{})
	Warn(message string, args ...interface{})
	Error(message interface{}, args ...interface{})
	Fatal(message interface{}, args ...interface{})
}

type Logger struct {
	logger *zerolog.Logger
}

var _ ILogger = (*Logger)(nil)

type LogLevel int

const (
	// it bugs if the intial value is 0
	DEBUG LogLevel = iota + 1
	INFO
	WARN
	ERROR
	FATAL
)

func (lvl LogLevel) String() string {
	return [...]string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}[lvl-1]
}

func (lvl *LogLevel) UnmarshalText(text []byte) error {
	str := strings.ToUpper(string(text))
	switch str {
	case "DEBUG":
		*lvl = DEBUG
	case "INFO":
		*lvl = INFO
	case "WARN":
		*lvl = WARN
	case "ERROR":
		*lvl = ERROR
	case "FATAL":
		*lvl = FATAL
	default:
		return errors.New("invalid log level")
	}
	return nil
}

func (f *LogLevel) SetValue(s string) error {
	if s == "" {
		return fmt.Errorf("missing log level. choose between [\"DEBUG\", \"INFO\", \"WARN\", \"ERROR\", \"FATAL\"] ")
	}
	switch strings.ToUpper(s) {
	case "DEBUG":
		*f = DEBUG
	case "INFO":
		*f = INFO
	case "WARN":
		*f = WARN
	case "ERROR":
		*f = ERROR
	case "FATAL":
		*f = FATAL
	default:
		fmt.Printf("%s unknown log level. falling down to INFO", s)
		*f = INFO
	}
	return nil
}

func New(lvl LogLevel) *Logger {
	var l zerolog.Level
	switch lvl {
	case DEBUG:
		l = zerolog.DebugLevel
	case INFO:
		l = zerolog.InfoLevel
	case WARN:
		l = zerolog.WarnLevel
	case ERROR:
		l = zerolog.ErrorLevel
	case FATAL:
		l = zerolog.FatalLevel
	}

	zerolog.SetGlobalLevel(l)

	logger := zerolog.New(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.Out = os.Stdout
		w.NoColor = false
		
	})).With().Timestamp().Logger()

	return &Logger{
		logger: &logger,
	}
}

func (l *Logger) Debug(message interface{}, args ...interface{}) {
	l.msg(DEBUG, message, args...)
}

func (l *Logger) Info(message string, args ...interface{}) {
	l.msg(INFO, message, args...)
}

func (l *Logger) Warn(message string, args ...interface{}) {
	l.msg(WARN, message, args...)
}

func (l *Logger) Error(message interface{}, args ...interface{}) {
	if l.logger.GetLevel() == zerolog.DebugLevel {
		l.Debug(message, args...)
	}
	l.msg(ERROR, message, args...)
}

func (l *Logger) Fatal(message interface{}, args ...interface{}) {
	l.msg(FATAL, message, args...)
	os.Exit(1)
}

func (l *Logger) msg(lvl LogLevel, message interface{}, args ...interface{}) {
	logEvent := l.logger.WithLevel(zerolog.Level(lvl - 1))
	switch msg := message.(type) {
	case error:
		if len(args) == 0 {
			logEvent.Err(msg).Msg(msg.Error())
		} else {
			logEvent.Err(msg).Msgf(msg.Error(), args...)
		}
	case string:
		if len(args) == 0 {
			logEvent.Msg(msg)
		} else {
			logEvent.Msgf(msg, args...)
		}
	default:
		logEvent.Msg(fmt.Sprintf("%s message %v has no known Type: %v", lvl.String(), message, msg))
	}
}
