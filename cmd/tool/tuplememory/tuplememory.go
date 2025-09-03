package tuplememory

import (
	"runtime"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type Instance struct {
	Ns   string
	Name string
}

var Cmd = &cobra.Command{
	Use:   "tuple-memory",
	Short: "Calculates memory usage of a large graph",
	Long:  `Creates a large in-memory graph and logs the memory usage.`,
	Run: func(cmd *cobra.Command, args []string) {
		calculateMemoryUsage(1000000, 50)
	},
}

func calculateMemoryUsage(numNodes int, edgesPerNode int) {
	// Force GC and record baseline memory usage
	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	baseline := m.Alloc
	log.Printf("Baseline memory: %.2f MB\n", float64(baseline)/1024/1024)

	// Create a big mock graph
	graph := make(map[Instance]map[string]map[Instance]struct{})

	rels := []string{"relA", "relB", "relC"} // relation keys
	total := 0

	st := time.Now()

	for i := range numNodes {
		from := Instance{Ns: strconv.Itoa(i), Name: "Node" + strconv.Itoa(i)}
		graph[from] = make(map[string]map[Instance]struct{})

		for _, rel := range rels {
			graph[from][rel] = make(map[Instance]struct{})
			for j := range edgesPerNode {
				to := Instance{
					Ns:   strconv.Itoa((i*edgesPerNode + j) % numNodes),
					Name: "Node" + strconv.Itoa((i*edgesPerNode+j)%numNodes),
				}
				graph[from][rel][to] = struct{}{}
				total++
			}
		}
	}

	log.Printf("Time to create graph: %s\n", time.Since(st))

	// Force GC and check memory usage again
	runtime.GC()
	runtime.ReadMemStats(&m)
	used := m.Alloc - baseline
	log.Printf("After graph creation: %.2f MB\n", float64(used)/1024/1024)

	// Keep the graph alive so GC doesn't free it before measurement
	if len(graph) == 0 {
		log.Print("graph is empty")
	}

	log.Printf("Total tuples: %d w\n ", total/10000)
}
