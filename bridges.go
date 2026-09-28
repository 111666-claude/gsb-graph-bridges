// Package graphbridges 求无向图的桥与割点，只用标准库。
package graphbridges

import (
	"errors"
	"fmt"
	"sort"
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

// ParseEdges 解析 "a-b,b-c"。缺 "-" 或顶点名为空时报错并指出是第几条边（从 1 数起）。
func ParseEdges(text string) ([]Edge, error) {
	edges := []Edge{}
	for i, item := range strings.Split(text, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		left, right, found := strings.Cut(item, "-")
		if !found {
			return nil, fmt.Errorf("%w：第 %d 条边 %q 缺少 \"-\"", ErrFormat, i+1, item)
		}
		left = strings.TrimSpace(left)
		right = strings.TrimSpace(right)
		if left == "" || right == "" {
			return nil, fmt.Errorf("%w：第 %d 条边 %q 顶点名不能为空", ErrFormat, i+1, item)
		}
		edges = append(edges, Edge{A: left, B: right})
	}
	return edges, nil
}

// half 是邻接表里的一条半边，edge 是输入边的编号。
type half struct {
	to   int
	edge int
}

// Solve 求桥、割点与连通分量数。
//
// 口径：
//   - 平行边不是桥：DFS 按边编号记父边，同一对顶点之间的第二条边算回边。
//   - 割点：非根节点用 low[child] >= disc[u]，根节点用「子节点数 > 1」。
//   - 自环忽略：不算桥、不产生割点，顶点本身保留。
//   - 桥与割点按字典序输出，桥的两个端点按字典序写。
//
// 用迭代式 DFS，深链不会爆栈；顶点先映射成整数编号，邻接表用切片，
// 时间 O(V+E)，内存 O(V+E)。
func Solve(edges []Edge) Result {
	ids := map[string]int{}
	names := []string{}
	adj := [][]half{}
	edgeCount := 0
	getID := func(name string) int {
		if id, ok := ids[name]; ok {
			return id
		}
		id := len(names)
		ids[name] = id
		names = append(names, name)
		adj = append(adj, nil)
		return id
	}
	for _, edge := range edges {
		a := getID(edge.A)
		b := getID(edge.B)
		if a == b {
			continue // 自环忽略
		}
		adj[a] = append(adj[a], half{to: b, edge: edgeCount})
		adj[b] = append(adj[b], half{to: a, edge: edgeCount})
		edgeCount++
	}

	n := len(names)
	disc := make([]int, n)
	low := make([]int, n)
	parent := make([]int, n)
	parentEdge := make([]int, n)
	iter := make([]int, n)
	articulation := make([]bool, n)
	bridges := []Edge{}
	clock := 0
	components := 0
	stack := []int{}

	for root := 0; root < n; root++ {
		if disc[root] != 0 {
			continue
		}
		components++
		clock++
		disc[root] = clock
		low[root] = clock
		parent[root] = -1
		parentEdge[root] = -1
		rootChildren := 0
		stack = append(stack[:0], root)
		for len(stack) > 0 {
			v := stack[len(stack)-1]
			if iter[v] < len(adj[v]) {
				h := adj[v][iter[v]]
				iter[v]++
				if h.edge == parentEdge[v] {
					continue // 只跳过来时的那条边，平行边算回边
				}
				w := h.to
				if disc[w] != 0 {
					if disc[w] < low[v] {
						low[v] = disc[w]
					}
					continue
				}
				clock++
				disc[w] = clock
				low[w] = clock
				parent[w] = v
				parentEdge[w] = h.edge
				if v == root {
					rootChildren++
				}
				stack = append(stack, w)
				continue
			}
			stack = stack[:len(stack)-1]
			p := parent[v]
			if p == -1 {
				continue
			}
			if low[v] < low[p] {
				low[p] = low[v]
			}
			if low[v] > disc[p] {
				a, b := names[p], names[v]
				if b < a {
					a, b = b, a
				}
				bridges = append(bridges, Edge{A: a, B: b})
			}
			if parent[p] != -1 && low[v] >= disc[p] {
				articulation[p] = true
			}
		}
		if rootChildren > 1 {
			articulation[root] = true
		}
	}

	sort.Slice(bridges, func(i, j int) bool {
		if bridges[i].A != bridges[j].A {
			return bridges[i].A < bridges[j].A
		}
		return bridges[i].B < bridges[j].B
	})
	cuts := []string{}
	for i, marked := range articulation {
		if marked {
			cuts = append(cuts, names[i])
		}
	}
	sort.Strings(cuts)
	return Result{Bridges: bridges, Articulation: cuts, Components: components}
}
