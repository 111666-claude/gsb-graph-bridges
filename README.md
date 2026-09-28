# graph-bridges

无向图的桥、割点与连通分量，只用 Go 标准库。

```
go test ./...
go run ./cmd/graph-bridges --edges 'a-b,b-c,c-d,d-b'
go run ./cmd/graph-bridges --edges 'a-b,a-b,b-c'
```

## 口径（README 为准）

- **平行边不是桥**：`--edges 'a-b,a-b,b-c'` 按口径是 `bridges=b-c`，
  现在 DFS 把「回到父节点的边」一律跳过，两条平行边被当成一条，给 `bridges=a-b,b-c`。
- **割点判定要用 `low[child] >= disc[u]`**（根节点用「子节点数 > 1」）：`--edges 'a-b,b-c,c-d,d-b'`
  按口径是 `articulation=b`，现在少了等号，给空。
- 输出：`bridges` 与 `articulation` 都按字典序、逗号分隔；桥的两个端点按字典序写。
- 自环忽略（不算桥、也不产生割点）；空顶点名与缺 `-` 要报错并指出是第几条边。
- 不变量：删掉一条桥之后连通分量数必定增加；`components` 与实际分量数一致。
- 规模：100 万点、500 万边要在 5 秒内出结果，内存 O(V+E)；深链（10 万个点一条链）也要能跑完。

## 实现要点

- DFS 按边编号记父边：只跳过来时的那一条边，同一对顶点之间的第二条平行边算回边。
- 割点判定：非根节点用 `low[child] >= disc[u]`，根节点用「子节点数 > 1」。
- `ParseEdges` 校验顶点名，缺 `-` 或空名时报错并指出是第几条边（从 1 数起）。
- 迭代式 DFS，深链不爆栈；顶点映射成整数编号后用切片邻接表，时间 O(V+E)、内存 O(V+E)。

## 计划新增的能力

```
--edges 边表    逗号分隔的 顶点-顶点
```

## 输出

```
bridges=b-c
articulation=b
components=1
```

## 目录

```
bridges.go              解析与 DFS 求桥/割点
cmd/graph-bridges       命令行入口
bridges_test.go         go test 用例
```
