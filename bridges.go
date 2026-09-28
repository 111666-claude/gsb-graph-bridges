// Package graphbridges 求无向图的桥、割点与连通分量，只用标准库。
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

// ParseEdges 解析 "a-b,b-c" 形式的边表。
// 空段（如连续逗号）忽略；顶点名为空或缺 "-" 时报错，
// 并指出是边表中的第几条边（从 1 开始）。
func ParseEdges(text string) ([]Edge, error) {
	edges := []Edge{}
	for index, raw := range strings.Split(text, ",") {
		item := strings.TrimSpace(raw)
		if item == "" {
			continue
		}
		left, right, found := strings.Cut(item, "-")
		if !found {
			return nil, fmt.Errorf("%w：第 %d 条边 %q 缺 \"-\"", ErrFormat, index+1, item)
		}
		left = strings.TrimSpace(left)
		right = strings.TrimSpace(right)
		if left == "" || right == "" {
			return nil, fmt.Errorf("%w：第 %d 条边 %q 存在空顶点名", ErrFormat, index+1, item)
		}
		edges = append(edges, Edge{A: left, B: right})
	}
	return edges, nil
}

// graph 是以紧凑整数编号表示的无向图（平行边各占一条有向弧，自环忽略）。
type graph struct {
	names []string
	start []int32
	to    []int32
	edge  []int32
}

// buildGraph 两遍扫描构建邻接表，map 只用于顶点名到编号的映射，保证 O(V+E) 内存。
func buildGraph(edges []Edge) graph {
	id := make(map[string]int32, len(edges)*2)
	names := make([]string, 0)
	degree := make([]int32, 0)

	getID := func(name string) int32 {
		if v, ok := id[name]; ok {
			return v
		}
		v := int32(len(names))
		id[name] = v
		names = append(names, name)
		degree = append(degree, 0)
		return v
	}

	type rawEdge struct {
		a, b int32
		orig int
	}
	parsed := make([]rawEdge, 0, len(edges))
	for i, e := range edges {
		if e.A == e.B {
			getID(e.A) // 自环节点仍然计入顶点集与连通分量
			continue
		}
		a := getID(e.A)
		b := getID(e.B)
		parsed = append(parsed, rawEdge{a: a, b: b, orig: i})
		degree[a]++
		degree[b]++
	}
	n := int32(len(names))
	start := make([]int32, n+1)
	for i := int32(0); i < n; i++ {
		start[i+1] = start[i] + degree[i]
	}
	to := make([]int32, 2*len(parsed))
	edgeID := make([]int32, 2*len(parsed))
	cursor := make([]int32, n)
	copy(cursor, start[:n])
	for _, e := range parsed {
		p := cursor[e.a]
		to[p] = e.b
		edgeID[p] = int32(e.orig)
		cursor[e.a]++
		p = cursor[e.b]
		to[p] = e.a
		edgeID[p] = int32(e.orig)
		cursor[e.b]++
	}
	return graph{names: names, start: start, to: to, edge: edgeID}
}

// Solve 求桥、割点与连通分量数。结果已排序：桥按端点字典序，
// 每条桥的两个端点也按字典序；割点按字典序。对同一输入结果稳定。
func Solve(edges []Edge) Result {
	g := buildGraph(edges)
	n := int32(len(g.names))

	// 0 = 未访问；disc/low 从 1 开始编号。
	disc := make([]int32, n)
	low := make([]int32, n)
	parentV := make([]int32, n)
	parentE := make([]int32, n)
	nextArc := make([]int32, n)
	isArt := make([]bool, n)
	bridges := make([]int32, 0)

	stackV := make([]int32, 0, n)
	var clock, components int32

	for root := int32(0); root < n; root++ {
		if disc[root] != 0 {
			continue
		}
		components++
		clock++
		disc[root] = clock
		low[root] = clock
		parentV[root] = -1
		parentE[root] = -1
		nextArc[root] = g.start[root]
		stackV = stackV[:0]
		stackV = append(stackV, root)
		children := int32(0)

		for len(stackV) > 0 {
			u := stackV[len(stackV)-1]
			arc := nextArc[u]
			if arc < g.start[u+1] {
				nextArc[u] = arc + 1
				eid := g.edge[arc]
				if eid == parentE[u] {
					continue // 只跳过 DFS 树边自身；平行边会作为回边处理
				}
				v := g.to[arc]
				if v == u {
					continue // 自环：忽略
				}
				if disc[v] != 0 {
					if disc[v] < low[u] {
						low[u] = disc[v]
					}
					continue
				}
				if u == root {
					children++
				}
				clock++
				disc[v] = clock
				low[v] = clock
				parentV[v] = u
				parentE[v] = eid
				nextArc[v] = g.start[v]
				stackV = append(stackV, v)
				continue
			}

			// u 的邻接弧已处理完，回溯到父节点。
			stackV = stackV[:len(stackV)-1]
			p := parentV[u]
			if p < 0 {
				if children > 1 {
					isArt[root] = true
				}
				continue
			}
			if low[u] < low[p] {
				low[p] = low[u]
			}
			if low[u] > disc[p] {
				bridges = append(bridges, parentE[u])
			}
			if low[u] >= disc[p] && p != root {
				isArt[p] = true
			}
		}
	}

	canon := func(e Edge) (string, string) {
		if e.A < e.B {
			return e.A, e.B
		}
		return e.B, e.A
	}
	sort.Slice(bridges, func(i, j int) bool {
		ai, aj := canon(edges[bridges[i]])
		bi, bj := canon(edges[bridges[j]])
		if ai != bi {
			return ai < bi
		}
		return aj < bj
	})
	bridgeList := make([]Edge, len(bridges))
	for i, eid := range bridges {
		a, b := edges[eid].A, edges[eid].B
		if b < a {
			a, b = b, a
		}
		bridgeList[i] = Edge{A: a, B: b}
	}

	art := make([]string, 0)
	for i, marked := range isArt {
		if marked {
			art = append(art, g.names[i])
		}
	}
	sort.Strings(art)

	return Result{Bridges: bridgeList, Articulation: art, Components: int(components)}
}
