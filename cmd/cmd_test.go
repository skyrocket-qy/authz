package cmd

import (
	"context"
	"net/http"
	"testing"
	"time"

	"authz/internal/util"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	originalRun := Cmd.Run
	defer func() { Cmd.Run = originalRun }()

	var executed bool
	Cmd.Run = func(cmd *cobra.Command, args []string) {
		executed = true
	}

	// Since Execute() calls cobra.Command.Execute() which can be blocking (e.g. running a server),
	// and we replaced the Run function with a non-blocking one, we can call it directly.
	Execute()
	assert.True(t, executed)
}

func TestNewHttpServer(t *testing.T) {
	lc := util.NewLifecycleParallel()
	handler := http.NewServeMux()
	server := NewHttpServer(lc, handler)

	assert.NotNil(t, server)
	assert.Equal(t, ":8080", server.Addr)
	assert.Equal(t, handler, server.Handler)

	// Test the lifecycle shutdown function
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err := lc.Shutdown(ctx)
	assert.NoError(t, err)
}
