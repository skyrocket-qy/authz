package service

import (
	"authz/cmd/service/job"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "service",
	Short: "run service",
	// Long:  `The longer description`,
	// Run: RunServer,
}

func init() {
	Cmd.AddCommand(job.Cmd)
}
