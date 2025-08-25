package server

import (
	"authz/cmd/server/job"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "server",
	Short: "run service",
	// Long:  `The longer description`,
	// Run: RunServer,
}

func init() {
	Cmd.AddCommand(job.Cmd)
}
