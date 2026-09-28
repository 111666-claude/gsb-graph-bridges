package graphbridges

import "testing"

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
