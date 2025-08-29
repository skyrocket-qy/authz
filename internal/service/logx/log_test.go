package logx_test

import (
	"bytes"
	"os"
	"testing"

	"authz/internal/config"
	"authz/internal/service/logx"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestInitLogger(t *testing.T) {
	// Test with local environment
	config.Env = config.EnvLocal
	err := logx.InitLogger()
	assert.NoError(t, err)
	assert.Equal(t, zerolog.DebugLevel, zerolog.GlobalLevel())

	// Test with production environment
	config.Env = config.EnvProd
	err = logx.InitLogger()
	assert.NoError(t, err)
	assert.Equal(t, zerolog.InfoLevel, zerolog.GlobalLevel())

	// Test the output format
	var buf bytes.Buffer

	log.Logger = log.Output(&buf)
	log.Info().Msg("test message")

	output := buf.String()
	assert.Contains(t, output, "test message")
}

func TestSimplifyCaller(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"/go/src/project/internal/service/logx/log.go:10", "logx/log.go:10"},
		{"/go/src/project/main.go:20", "project/main.go:20"},
		{"log.go:5", "log.go:5"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.expected, logx.SimplifyCaller(tc.input))
		})
	}
}

func TestMain(m *testing.M) {
	// Save original logger
	originalLogger := log.Logger
	// Run tests
	exitCode := m.Run()
	// Restore original logger
	log.Logger = originalLogger

	os.Exit(exitCode)
}
