package shardmem

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "shardmem",
	Short: "",
	// Long:  `The longer description`,
	Run: start,
}

func start(cmd *cobra.Command, args []string) {
	shardConfigs := []int{1, 16, 32, 128, 256, 1024, 4096, 8192, 10e4}

	for _, n := range shardConfigs {
		shards := make([]map[int]struct{}, n)
		for i := range n {
			shards[i] = make(map[int]struct{})
		}

		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		fmt.Printf("NumShards: %4d, Memory Alloc: %.2f MB\n", n, float64(m.Alloc)/1024/1024)

		// Prevent compiler optimizing away
		if len(shards[0]) == 0 && n == 0 {
			fmt.Println()
		}
	}
}
