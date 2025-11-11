package log

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestNewLoggerSetsLevel(t *testing.T) {
	t.Parallel()

	provider := New("debug", true)
	require.NotNil(t, provider)

	require.Equal(t, zerolog.DebugLevel, zerolog.GlobalLevel())

	logger := provider.Logger()
	logger.Debug().Msg("test")
}

func TestNewLoggerDefaultLevel(t *testing.T) {
	t.Parallel()

	provider := New("", false)
	require.NotNil(t, provider)

	require.Equal(t, zerolog.InfoLevel, zerolog.GlobalLevel())

	logger := provider.Logger()
	logger.Info().Msg("info")
}

func TestParseLevelUnknown(t *testing.T) {
	t.Parallel()

	level := parseLevel("unknown")
	require.Equal(t, zerolog.InfoLevel, level)
}

func TestParseLevelVariants(t *testing.T) {
	t.Parallel()

	cases := map[string]zerolog.Level{
		"trace":   zerolog.TraceLevel,
		"debug":   zerolog.DebugLevel,
		"info":    zerolog.InfoLevel,
		"warn":    zerolog.WarnLevel,
		"warning": zerolog.WarnLevel,
		"error":   zerolog.ErrorLevel,
		"fatal":   zerolog.FatalLevel,
		"panic":   zerolog.PanicLevel,
	}

	for input, expected := range cases {
		lev := parseLevel(input)
		require.Equal(t, expected, lev, input)
	}
}
