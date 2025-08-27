package tool

import (
	"authz/cmd/tool/shardmem"
	"authz/cmd/tool/tuplememory"
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
	Cmd.AddCommand(tuplememory.Cmd)
}
