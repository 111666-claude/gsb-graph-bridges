package graphbridges

import (
	"errors"
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestParseEdges(t *testing.T) {
	edges, err := ParseEdges("a-b, b-c")
	if err != nil {
		t.Fatalf("ParseEdges 出错：%v", err)
	}
	if len(edges) != 2 || edges[0] != (Edge{A: "a", B: "b"}) {
		t.Fatalf("解析结果不对：%+v", edges)
	}
	if _, err := ParseEdges("ab"); err == nil {
		t.Fatal("缺 - 应该报错")
	}
}

func TestSingleEdgeIsABridge(t *testing.T) {
	edges, _ := ParseEdges("a-b")
	result := Solve(edges)
	if len(result.Bridges) != 1 || result.Components != 1 {
		t.Fatalf("单条边应是桥且只有一个分量：%+v", result)
	}
}

func TestPathBridges(t *testing.T) {
	edges, _ := ParseEdges("a-b,b-c")
	result := Solve(edges)
	if len(result.Bridges) != 2 {
		t.Fatalf("链上两条边都是桥：%+v", result.Bridges)
	}
}

func TestCycleHasNoBridges(t *testing.T) {
	edges, _ := ParseEdges("a-b,b-c,c-a")
	result := Solve(edges)
	if len(result.Bridges) != 0 {
		t.Fatalf("环上没有桥：%+v", result.Bridges)
	}
	if len(result.Articulation) != 0 {
		t.Fatalf("环上没有割点：%+v", result.Articulation)
	}
}

func TestDisconnectedComponents(t *testing.T) {
	edges, _ := ParseEdges("a-b,c-d,e-f")
	result := Solve(edges)
	if result.Components != 3 {
		t.Fatalf("三个分量：%+v", result)
	}
}

func TestEmptyGraph(t *testing.T) {
	result := Solve(nil)
	if result.Components != 0 || len(result.Bridges) != 0 {
		t.Fatalf("空图应该什么都没有：%+v", result)
	}
}

func TestParallelEdgesAreNotBridges(t *testing.T) {
	edges, _ := ParseEdges("a-b,a-b,b-c")
	result := Solve(edges)
	if len(result.Bridges) != 1 || result.Bridges[0] != (Edge{A: "b", B: "c"}) {
		t.Fatalf("平行边不是桥，应只有 b-c：%+v", result.Bridges)
	}
	if result.Components != 1 {
		t.Fatalf("分量数应为 1：%+v", result)
	}
}

func TestArticulationUsesGreaterOrEqual(t *testing.T) {
	edges, _ := ParseEdges("a-b,b-c,c-d,d-b")
	result := Solve(edges)
	if len(result.Articulation) != 1 || result.Articulation[0] != "b" {
		t.Fatalf("割点应只有 b：%+v", result.Articulation)
	}
	if len(result.Bridges) != 1 || result.Bridges[0] != (Edge{A: "a", B: "b"}) {
		t.Fatalf("桥应只有 a-b：%+v", result.Bridges)
	}
}

func TestRootArticulationNeedsTwoChildren(t *testing.T) {
	edges, _ := ParseEdges("a-b,a-c")
	result := Solve(edges)
	if len(result.Articulation) != 1 || result.Articulation[0] != "a" {
		t.Fatalf("根节点有两个子节点才是割点：%+v", result.Articulation)
	}
}

func TestSelfLoopIgnored(t *testing.T) {
	edges, _ := ParseEdges("a-a,a-b")
	result := Solve(edges)
	if len(result.Bridges) != 1 || result.Bridges[0] != (Edge{A: "a", B: "b"}) {
		t.Fatalf("自环不算桥：%+v", result.Bridges)
	}
	if len(result.Articulation) != 0 {
		t.Fatalf("自环不产生割点：%+v", result.Articulation)
	}
	if result.Components != 1 {
		t.Fatalf("分量数应为 1：%+v", result)
	}
	only, _ := ParseEdges("a-a")
	if got := Solve(only); got.Components != 1 || len(got.Bridges) != 0 {
		t.Fatalf("只有自环的顶点仍算一个分量：%+v", got)
	}
}

func TestParseErrorReportsEdgeIndex(t *testing.T) {
	if _, err := ParseEdges("a-b,cd"); err == nil || !strings.Contains(err.Error(), "第 2 条") {
		t.Fatalf("缺 - 要报第几条边：%v", err)
	}
	if _, err := ParseEdges("a-b,c-"); err == nil || !strings.Contains(err.Error(), "第 2 条") {
		t.Fatalf("空顶点名要报第几条边：%v", err)
	}
	if _, err := ParseEdges("-b"); err == nil || !strings.Contains(err.Error(), "第 1 条") {
		t.Fatalf("空顶点名要报第几条边：%v", err)
	}
	if _, err := ParseEdges("ab"); !errors.Is(err, ErrFormat) {
		t.Fatalf("错误应可 errors.Is 匹配 ErrFormat：%v", err)
	}
}

func TestDeepChain(t *testing.T) {
	const n = 100000
	edges := make([]Edge, 0, n-1)
	for i := 0; i < n-1; i++ {
		edges = append(edges, Edge{A: strconv.Itoa(i), B: strconv.Itoa(i + 1)})
	}
	result := Solve(edges)
	if len(result.Bridges) != n-1 {
		t.Fatalf("深链应有 %d 条桥，得到 %d", n-1, len(result.Bridges))
	}
	if result.Components != 1 {
		t.Fatalf("深链只有一个分量：%+v", result.Components)
	}
}

func TestDeterministic(t *testing.T) {
	edges, _ := ParseEdges("a-b,a-b,b-c,c-d,d-b,e-f,f-g,g-e,b-e")
	first := Solve(edges)
	second := Solve(edges)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("同一输入两次求解结果不同：%+v vs %+v", first, second)
	}
}

// bruteForce 用「逐条删边、逐点删点再数分量」的笨办法求口径答案。
func bruteForce(edges []Edge) Result {
	normal := []Edge{}
	vertexSet := map[string]bool{}
	for _, e := range edges {
		vertexSet[e.A] = true
		vertexSet[e.B] = true
		if e.A != e.B {
			normal = append(normal, e)
		}
	}
	vertices := []string{}
	for v := range vertexSet {
		vertices = append(vertices, v)
	}
	countComponents := func(dropEdge int, dropVertex string) int {
		adj := map[string][]string{}
		alive := map[string]bool{}
		for _, v := range vertices {
			if v != dropVertex {
				alive[v] = true
			}
		}
		for i, e := range normal {
			if i == dropEdge {
				continue
			}
			if e.A == dropVertex || e.B == dropVertex {
				continue
			}
			adj[e.A] = append(adj[e.A], e.B)
			adj[e.B] = append(adj[e.B], e.A)
		}
		seen := map[string]bool{}
		components := 0
		for v := range alive {
			if seen[v] {
				continue
			}
			components++
			queue := []string{v}
			seen[v] = true
			for len(queue) > 0 {
				cur := queue[0]
				queue = queue[1:]
				for _, next := range adj[cur] {
					if !seen[next] {
						seen[next] = true
						queue = append(queue, next)
					}
				}
			}
		}
		return components
	}
	base := countComponents(-1, "")
	bridges := []Edge{}
	for i, e := range normal {
		if countComponents(i, "") > base {
			a, b := e.A, e.B
			if b < a {
				a, b = b, a
			}
			bridges = append(bridges, Edge{A: a, B: b})
		}
	}
	articulation := []string{}
	for _, v := range vertices {
		if countComponents(-1, v) > base {
			articulation = append(articulation, v)
		}
	}
	sort.Slice(bridges, func(i, j int) bool {
		if bridges[i].A != bridges[j].A {
			return bridges[i].A < bridges[j].A
		}
		return bridges[i].B < bridges[j].B
	})
	sort.Strings(articulation)
	return Result{Bridges: bridges, Articulation: articulation, Components: base}
}

func TestAgainstBruteForce(t *testing.T) {
	rng := rand.New(rand.NewSource(20260928))
	for trial := 0; trial < 300; trial++ {
		n := 1 + rng.Intn(8)
		names := make([]string, n)
		for i := range names {
			names[i] = string(rune('a' + i))
		}
		m := rng.Intn(2 * n * (n + 1))
		edges := make([]Edge, 0, m)
		for i := 0; i < m; i++ {
			edges = append(edges, Edge{
				A: names[rng.Intn(n)],
				B: names[rng.Intn(n)],
			})
		}
		got := Solve(edges)
		want := bruteForce(edges)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("第 %d 个随机图不一致：\n边：%+v\n得到：%+v\n暴力：%+v", trial, edges, got, want)
		}
	}
}

func TestBridgeRemovalIncreasesComponents(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	for trial := 0; trial < 100; trial++ {
		n := 2 + rng.Intn(7)
		names := make([]string, n)
		for i := range names {
			names[i] = string(rune('a' + i))
		}
		m := rng.Intn(2 * n * n)
		edges := make([]Edge, 0, m)
		for i := 0; i < m; i++ {
			edges = append(edges, Edge{
				A: names[rng.Intn(n)],
				B: names[rng.Intn(n)],
			})
		}
		result := Solve(edges)
		for _, bridge := range result.Bridges {
			rest := make([]Edge, 0, len(edges))
			removed := false
			for _, e := range edges {
				if !removed && e.A != e.B &&
					((e.A == bridge.A && e.B == bridge.B) || (e.A == bridge.B && e.B == bridge.A)) {
					removed = true
					continue
				}
				rest = append(rest, e)
			}
			// 顶点集固定：给每个原顶点补一条自环，避免删边后顶点消失。
			seen := map[string]bool{}
			fixed := rest
			for _, e := range edges {
				for _, v := range []string{e.A, e.B} {
					if !seen[v] {
						seen[v] = true
						fixed = append(fixed, Edge{A: v, B: v})
					}
				}
			}
			if got := Solve(fixed).Components; got != result.Components+1 {
				t.Fatalf("删掉桥 %+v 后分量数应从 %d 变成 %d，得到 %d",
					bridge, result.Components, result.Components+1, got)
			}
		}
	}
}

func TestLargeGraphWithinTimeBudget(t *testing.T) {
	if testing.Short() {
		t.Skip("短模式跳过规模测试")
	}
	const v, e = 1000000, 5000000
	edges := make([]Edge, 0, e)
	// 先造一条保证连通的链，再补随机边。
	for i := 0; i < v-1; i++ {
		edges = append(edges, Edge{A: strconv.Itoa(i), B: strconv.Itoa(i + 1)})
	}
	rng := rand.New(rand.NewSource(1))
	for len(edges) < e {
		edges = append(edges, Edge{
			A: strconv.Itoa(rng.Intn(v)),
			B: strconv.Itoa(rng.Intn(v)),
		})
	}
	start := time.Now()
	result := Solve(edges)
	elapsed := time.Since(start)
	if result.Components != 1 {
		t.Fatalf("连通图分量数应为 1：%d", result.Components)
	}
	if elapsed > 5*time.Second {
		t.Fatalf("100 万点 500 万边应在 5 秒内跑完，实际 %v", elapsed)
	}
	t.Logf("100 万点 500 万边耗时 %v", elapsed)
}
