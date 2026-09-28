// 命令行入口：graph-bridges --edges 'a-b,b-c'
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	graphbridges "example.com/graphbridges"
)

func main() {
	edgesFlag := flag.String("edges", "", "逗号分隔的边表，如 a-b,b-c")
	flag.Parse()

	edges, err := graphbridges.ParseEdges(*edgesFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	result := graphbridges.Solve(edges)

	bridgeNames := make([]string, len(result.Bridges))
	for i, bridge := range result.Bridges {
		bridgeNames[i] = bridge.A + "-" + bridge.B
	}
	fmt.Printf("bridges=%s\n", strings.Join(bridgeNames, ","))
	fmt.Printf("articulation=%s\n", strings.Join(result.Articulation, ","))
	fmt.Printf("components=%d\n", result.Components)
}
