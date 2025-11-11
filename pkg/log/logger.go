package log

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// LoggerProvider encapsula a criação do logger padrão da aplicação.
type LoggerProvider struct {
	logger zerolog.Logger
}

// New cria uma instância de LoggerProvider com nível configurável.
func New(levelStr string, pretty bool) *LoggerProvider {
	level := parseLevel(levelStr)
	writer := io.Writer(os.Stdout)

	if pretty {
		writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	zerolog.SetGlobalLevel(level)
	logger := zerolog.New(writer).With().Timestamp().Logger()

	return &LoggerProvider{logger: logger}
}

// Logger retorna o logger configurado.
func (p *LoggerProvider) Logger() zerolog.Logger {
	return p.logger
}

func parseLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info", "":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}

