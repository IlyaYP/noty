package logging

import (
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog"
)

// LoggerOption defines logger customization option.
type LoggerOption func(logger zerolog.Logger) zerolog.Logger

// WithLogLevel sets log level.
func WithLogLevel(level zerolog.Level) LoggerOption {
	return func(logger zerolog.Logger) zerolog.Logger {
		return logger.Level(level)
	}
}

// NewLogger creates a new customizable logger.
func NewLogger(opts ...LoggerOption) zerolog.Logger {

	// https://github.com/rs/zerolog#add-file-and-line-number-to-log
	zerolog.CallerMarshalFunc = func(file string, line int) string {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		return file + ":" + strconv.Itoa(line)
	}

	logger := zerolog.New(os.Stdout).
		Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Stamp}).
		Level(zerolog.TraceLevel).
		With().
		Timestamp().
		Caller().
		Logger()

	for _, opt := range opts {
		logger = opt(logger)
	}

	return logger
}
