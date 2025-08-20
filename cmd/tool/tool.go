package tool

import (
	"authz/cmd/tool/shardmem"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "tool",
	Short: "",
	// Long:  `The longer description`,
	// Run: start,
}

func init() {
	Cmd.AddCommand(shardmem.Cmd)
}
