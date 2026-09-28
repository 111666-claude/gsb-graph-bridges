package graphbridges

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

func parse(t *testing.T, text string) []Edge {
	t.Helper()
	edges, err := ParseEdges(text)
	if err != nil {
		t.Fatalf("ParseEdges(%q) 出错：%v", text, err)
	}
	return edges
}

func TestParseEdges(t *testing.T) {
	edges := parse(t, "a-b, b-c")
	if len(edges) != 2 || edges[0] != (Edge{A: "a", B: "b"}) {
		t.Fatalf("解析结果不对：%+v", edges)
	}
	if len(parse(t, "a-b,,b-c,")) != 2 {
		t.Fatal("空段应当忽略")
	}
}

func TestParseEdgesErrors(t *testing.T) {
	cases := []struct {
		text string
		line int
	}{
		{"ab", 1},
		{"a-b,cd", 2},
		{"a-b, ,c-d,ef", 4},
		{"a-", 1},
		{"-b", 1},
		{"a-b,-c", 2},
		{"a-b,c-", 2},
		{"a -", 1},
	}
	for _, tc := range cases {
		_, err := ParseEdges(tc.text)
		if !errors.Is(err, ErrFormat) {
			t.Fatalf("ParseEdges(%q) 应返回 ErrFormat，实际：%v", tc.text, err)
		}
		want := fmt.Sprintf("第 %d 条边", tc.line)
		if got := err.Error(); !strings.Contains(got, want) {
			t.Fatalf("ParseEdges(%q) 的错误 %q 应指出 %s", tc.text, got, want)
		}
	}
}

func TestSingleEdgeIsABridge(t *testing.T) {
	result := Solve(parse(t, "a-b"))
	if len(result.Bridges) != 1 || result.Components != 1 {
		t.Fatalf("单条边应是桥且只有一个分量：%+v", result)
	}
}

func TestPathBridges(t *testing.T) {
	result := Solve(parse(t, "a-b,b-c"))
	if len(result.Bridges) != 2 {
		t.Fatalf("链上两条边都是桥：%+v", result.Bridges)
	}
	if got := result.Articulation; len(got) != 1 || got[0] != "b" {
		t.Fatalf("链中间点 b 是割点：%+v", got)
	}
}

func TestCycleHasNoBridges(t *testing.T) {
	result := Solve(parse(t, "a-b,b-c,c-a"))
	if len(result.Bridges) != 0 {
		t.Fatalf("环上没有桥：%+v", result.Bridges)
	}
	if len(result.Articulation) != 0 {
		t.Fatalf("环上没有割点：%+v", result.Articulation)
	}
}

func TestParallelEdges(t *testing.T) {
	result := Solve(parse(t, "a-b,a-b,b-c"))
	if len(result.Bridges) != 1 || result.Bridges[0] != (Edge{A: "b", B: "c"}) {
		t.Fatalf("平行边不是桥，只有 b-c 是桥：%+v", result.Bridges)
	}
	if got := result.Articulation; len(got) != 1 || got[0] != "b" {
		t.Fatalf("只有 b 是割点：%+v", got)
	}

	result = Solve(parse(t, "a-b,a-b"))
	if len(result.Bridges) != 0 || len(result.Articulation) != 0 || result.Components != 1 {
		t.Fatalf("一对平行边无桥无割点：%+v", result)
	}
}

func TestArticulationExample(t *testing.T) {
	result := Solve(parse(t, "a-b,b-c,c-d,d-b"))
	if len(result.Bridges) != 1 || result.Bridges[0] != (Edge{A: "a", B: "b"}) {
		t.Fatalf("只有 a-b 是桥：%+v", result.Bridges)
	}
	if got := result.Articulation; len(got) != 1 || got[0] != "b" {
		t.Fatalf("按口径 articulation=b：%+v", got)
	}
}

func TestArticulationRoot(t *testing.T) {
	// 根节点有两棵 DFS 子树时才是割点。
	result := Solve(parse(t, "r-a,r-b,a-x,b-y"))
	art := map[string]bool{}
	for _, name := range result.Articulation {
		art[name] = true
	}
	for _, name := range []string{"r", "a", "b"} {
		if !art[name] {
			t.Fatalf("%s 应是割点：%+v", name, result.Articulation)
		}
	}
	if art["x"] || art["y"] {
		t.Fatalf("叶子不是割点：%+v", result.Articulation)
	}
}

func TestSelfLoops(t *testing.T) {
	result := Solve(parse(t, "a-a"))
	if result.Components != 1 || len(result.Bridges) != 0 || len(result.Articulation) != 0 {
		t.Fatalf("单个自环：一个分量、无桥无割点：%+v", result)
	}
	result = Solve(parse(t, "a-a,a-b,b-b,b-c,c-c"))
	if len(result.Bridges) != 2 || result.Components != 1 {
		t.Fatalf("自环不影响桥的判定：%+v", result)
	}
	art := map[string]bool{}
	for _, name := range result.Articulation {
		art[name] = true
	}
	if !art["b"] || art["a"] || art["c"] {
		t.Fatalf("链上自环后仍只有 b 是割点：%+v", result.Articulation)
	}
}

func TestOutputSorted(t *testing.T) {
	result := Solve(parse(t, "z-a,a-b,b-c,c-d"))
	wantBridges := []Edge{
		{A: "a", B: "b"},
		{A: "a", B: "z"},
		{A: "b", B: "c"},
		{A: "c", B: "d"},
	}
	if len(result.Bridges) != len(wantBridges) {
		t.Fatalf("桥数量不对：%+v", result.Bridges)
	}
	for i := range wantBridges {
		if result.Bridges[i] != wantBridges[i] {
			t.Fatalf("第 %d 条桥顺序不对：got %+v want %+v", i, result.Bridges, wantBridges)
		}
	}
	wantArt := []string{"a", "b", "c"}
	for i := range wantArt {
		if result.Articulation[i] != wantArt[i] {
			t.Fatalf("割点应按字典序：got %+v want %+v", result.Articulation, wantArt)
		}
	}
}

func TestDisconnectedComponents(t *testing.T) {
	result := Solve(parse(t, "a-b,c-d,e-f"))
	if result.Components != 3 {
		t.Fatalf("三个分量：%+v", result)
	}
	// 孤立自环节点也是一个分量。
	result = Solve(parse(t, "a-b,x-x"))
	if result.Components != 2 {
		t.Fatalf("自环节点单独一个分量：%+v", result)
	}
}

func TestEmptyGraph(t *testing.T) {
	result := Solve(nil)
	if result.Components != 0 || len(result.Bridges) != 0 || len(result.Articulation) != 0 {
		t.Fatalf("空图应该什么都没有：%+v", result)
	}
}

func TestDeterministic(t *testing.T) {
	edges := parse(t, "a-b,b-c,c-d,d-b,a-e,e-f,f-a")
	first := Solve(edges)
	for i := 0; i < 10; i++ {
		got := Solve(edges)
		if fmt.Sprint(got) != fmt.Sprint(first) {
			t.Fatalf("重复求解结果不一致：%+v != %+v", got, first)
		}
	}
}

func TestDeepChain(t *testing.T) {
	edges := make([]Edge, 99999)
	for i := range edges {
		edges[i] = Edge{A: fmt.Sprintf("n%d", i), B: fmt.Sprintf("n%d", i+1)}
	}
	result := Solve(edges)
	if len(result.Bridges) != 99999 {
		t.Fatalf("10 万点深链应有 99999 条桥，实际 %d", len(result.Bridges))
	}
	if result.Components != 1 {
		t.Fatalf("深链只有一个分量：%d", result.Components)
	}
	if len(result.Articulation) != 99998 {
		t.Fatalf("深链应有 99998 个割点，实际 %d", len(result.Articulation))
	}
	// 桥的端点按字典序、整体按字典序。
	for _, bridge := range result.Bridges {
		if bridge.A > bridge.B {
			t.Fatalf("桥的端点未按字典序：%+v", bridge)
		}
	}
}

// bruteComponents 用并查集计算连通分量数（自环与平行边自然支持）。
// skipEdge 为被删边下标（-1 表示不删），skipVertex 为被删顶点（空表示不删）。
// 删边时所有顶点仍计入；删点时被删顶点不计入。
func bruteComponents(edges []Edge, skipEdge int, skipVertex string) int {
	parent := map[string]string{}
	find := func(x string) string {
		for parent[x] != x {
			parent[x] = parent[parent[x]]
			x = parent[x]
		}
		return x
	}
	add := func(name string) {
		if _, ok := parent[name]; !ok {
			parent[name] = name
		}
	}
	for _, e := range edges {
		if e.A != skipVertex {
			add(e.A)
		}
		if e.B != skipVertex {
			add(e.B)
		}
	}
	for i, e := range edges {
		if i == skipEdge {
			continue
		}
		if skipVertex != "" && (e.A == skipVertex || e.B == skipVertex) {
			continue
		}
		ra, rb := find(e.A), find(e.B)
		if ra != rb {
			parent[ra] = rb
		}
	}
	seen := map[string]bool{}
	count := 0
	for name := range parent {
		root := find(name)
		if !seen[root] {
			seen[root] = true
			count++
		}
	}
	return count
}

func edgeKey(e Edge) (string, string) {
	if e.A < e.B {
		return e.A, e.B
	}
	return e.B, e.A
}

func TestMatchesBruteForce(t *testing.T) {
	rng := rand.New(rand.NewSource(20260928))
	for iter := 0; iter < 2000; iter++ {
		n := 1 + rng.Intn(8)
		names := make([]string, n)
		for i := range names {
			names[i] = fmt.Sprintf("v%d", i)
		}
		m := rng.Intn(12)
		edges := make([]Edge, m)
		for i := range edges {
			a, b := names[rng.Intn(n)], names[rng.Intn(n)]
			edges[i] = Edge{A: a, B: b} // 含自环与平行边
		}
		result := Solve(edges)
		base := bruteComponents(edges, -1, "")

		if result.Components != base {
			t.Fatalf("iter %d components: got %d want %d, edges=%v", iter, result.Components, base, edges)
		}

		gotBridges := map[[2]string]bool{}
		for _, e := range result.Bridges {
			gotBridges[[2]string{e.A, e.B}] = true
		}
		wantBridges := map[[2]string]bool{}
		for i, e := range edges {
			if e.A == e.B {
				continue
			}
			ka, kb := edgeKey(e)
			key := [2]string{ka, kb}
			if bruteComponents(edges, i, "") > base {
				wantBridges[key] = true
			}
		}
		if len(gotBridges) != len(wantBridges) {
			t.Fatalf("iter %d 桥数量不一致：got %v want %v，edges=%v", iter, gotBridges, wantBridges, edges)
		}
		for key := range wantBridges {
			if !gotBridges[key] {
				t.Fatalf("iter %d 缺少桥 %v，edges=%v", iter, key, edges)
			}
		}

		gotArt := map[string]bool{}
		for _, name := range result.Articulation {
			gotArt[name] = true
		}
		vertices := map[string]bool{}
		for _, e := range edges {
			vertices[e.A] = true
			vertices[e.B] = true
		}
		for v := range vertices {
			want := bruteComponents(edges, -1, v) > base
			if gotArt[v] != want {
				t.Fatalf("iter %d 顶点 %s 割点判定不一致：got %v want %v，edges=%v", iter, v, gotArt[v], want, edges)
			}
		}
	}
}

func buildSizedEdges(vertices, edgeCount int, extra func(rng *rand.Rand, i int) Edge) []Edge {
	edges := make([]Edge, edgeCount)
	rng := rand.New(rand.NewSource(1))
	for i := 0; i < vertices-1; i++ {
		edges[i] = Edge{A: fmt.Sprintf("v%d", i), B: fmt.Sprintf("v%d", i+1)}
	}
	for i := vertices - 1; i < edgeCount; i++ {
		edges[i] = extra(rng, i)
	}
	return edges
}

// BenchmarkMillionVertices 构造一条 100 万点的链（999999 条桥），
// 其余边只作平行边，便于校验桥数量同时达到 500 万边规模。
func BenchmarkMillionVertices(b *testing.B) {
	const vertices = 1_000_000
	const edgeCount = 5_000_000
	const doubled = 200_000
	edges := buildSizedEdges(vertices, edgeCount, func(rng *rand.Rand, i int) Edge {
		j := rng.Intn(doubled)
		return Edge{A: fmt.Sprintf("v%d", j), B: fmt.Sprintf("v%d", j+1)}
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := Solve(edges)
		if result.Components != 1 || len(result.Bridges) != vertices-1-doubled {
			b.Fatalf("基准结果异常：components=%d bridges=%d", result.Components, len(result.Bridges))
		}
	}
}

// BenchmarkDenseRandom 测稠密随机图上的纯吞吐。
func BenchmarkDenseRandom(b *testing.B) {
	const vertices = 1_000_000
	const edgeCount = 5_000_000
	edges := buildSizedEdges(vertices, edgeCount, func(rng *rand.Rand, i int) Edge {
		a := rng.Intn(vertices)
		c := rng.Intn(vertices - 1)
		if c >= a {
			c++
		}
		return Edge{A: fmt.Sprintf("v%d", a), B: fmt.Sprintf("v%d", c)}
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Solve(edges)
	}
}
