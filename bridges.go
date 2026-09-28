// Package graphbridges 求无向图的桥与割点，只用标准库。
package graphbridges

import (
	"errors"
	"strings"
)

// ErrFormat 表示边表格式不对。
var ErrFormat = errors.New("graphbridges: 边表格式不对")

// Edge 是输入的一条边，A 与 B 是顶点名。
type Edge struct {
	A string
	B string
}

// Result 是一次求解的结果。
type Result struct {
	Bridges      []Edge
	Articulation []string
	Components   int
}

// ParseEdges 解析 "a-b,b-c"。缺陷：不校验顶点名，报错也不给边的下标。
func ParseEdges(text string) ([]Edge, error) {
	edges := []Edge{}
	for _, item := range strings.Split(text, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		left, right, found := strings.Cut(item, "-")
		if !found {
			return nil, ErrFormat
		}
		edges = append(edges, Edge{A: left, B: right})
	}
	return edges, nil
}

type solver struct {
	adj          map[string][]string
	nodes        []string
	disc         map[string]int
	low          map[string]int
	bridges      []Edge
	articulation map[string]bool
	clock        int
}

// dfs 缺陷：把「回到父节点的边」一律跳过，平行边因此被当成一条；
// 割点判定又用了 low[child] > disc[u]，少了等号。
func (s *solver) dfs(node, parent string) {
	s.clock++
	s.disc[node] = s.clock
	s.low[node] = s.clock
	children := 0
	for _, next := range s.adj[node] {
		if next == parent {
			continue
		}
		if reached, ok := s.disc[next]; ok {
			if reached < s.low[node] {
				s.low[node] = reached
			}
			continue
		}
		children++
		s.dfs(next, node)
		if s.low[next] < s.low[node] {
			s.low[node] = s.low[next]
		}
		if s.low[next] > s.disc[node] {
			s.bridges = append(s.bridges, Edge{A: node, B: next})
		}
		if parent != "" && s.low[next] > s.disc[node] {
			s.articulation[node] = true
		}
	}
	if parent == "" && children > 1 {
		s.articulation[node] = true
	}
}

// Solve 求桥、割点与连通分量数。
func Solve(edges []Edge) Result {
	s := &solver{
		adj:          map[string][]string{},
		disc:         map[string]int{},
		low:          map[string]int{},
		articulation: map[string]bool{},
	}
	for _, edge := range edges {
		if edge.A == edge.B {
			if _, ok := s.disc[edge.A]; !ok {
				s.disc[edge.A] = 0
			}
			continue
		}
		if _, ok := s.adj[edge.A]; !ok {
			s.nodes = append(s.nodes, edge.A)
		}
		if _, ok := s.adj[edge.B]; !ok {
			s.nodes = append(s.nodes, edge.B)
		}
		s.adj[edge.A] = append(s.adj[edge.A], edge.B)
		s.adj[edge.B] = append(s.adj[edge.B], edge.A)
	}
	components := 0
	for _, node := range s.nodes {
		if _, ok := s.disc[node]; ok {
			continue
		}
		components++
		s.dfs(node, "")
	}
	names := []string{}
	for name, marked := range s.articulation {
		if marked {
			names = append(names, name)
		}
	}
	return Result{Bridges: s.bridges, Articulation: names, Components: components}
}
