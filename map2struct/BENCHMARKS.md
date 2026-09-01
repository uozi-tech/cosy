# map2struct 性能基线

`benchmark_test.go` 使用包含标量、嵌套结构、slice、时间、decimal、null.String 和 pgtype.Date 的模型，作为 `map2struct.WeakDecode` 的稳定性能基线。

它与根包的 `BenchmarkPipelineDecode` 使用不同模型：这里用于观察公开 API 的完整解码能力；根包 benchmark 用于拆分典型 HTTP 写请求管线，二者的 ns/op 和 allocs/op 不应直接混用。

## 当前结果

复现命令：

```sh
go test ./map2struct -run '^$' -bench '^BenchmarkWeakDecode' -benchmem -count=5 -cpu=1
```

测试环境：2026-09-01，Go 1.27.0，darwin/arm64，Apple M5 Pro。

### 串行解码

| run | ns/op | B/op | allocs/op |
|---:|---:|---:|---:|
| 1 | 1,022 | 536 | 13 |
| 2 | 1,010 | 536 | 13 |
| 3 | 969.8 | 536 | 13 |
| 4 | 959.6 | 536 | 13 |
| 5 | 955.6 | 536 | 13 |

中位数：**969.8 ns/op、536 B/op、13 allocs/op**。

### 并行安全路径

| run | ns/op | B/op | allocs/op |
|---:|---:|---:|---:|
| 1 | 1,044 | 536 | 13 |
| 2 | 1,030 | 536 | 13 |
| 3 | 952.1 | 536 | 13 |
| 4 | 962.6 | 536 | 13 |
| 5 | 1,002 | 536 | 13 |

中位数：**1,002 ns/op、536 B/op、13 allocs/op**。

这里固定 `-cpu=1`，用于减少跨轮次调度差异并验证 `RunParallel` 路径；评估整机并行吞吐时应移除该参数，并单独记录 `GOMAXPROCS`。

## 旧引擎对照

重构前的 mapstructure v1.5.0 基线记录于 2026-07-18，环境为 Go 1.26.5、darwin/arm64、Apple M5 Pro：

- 中位数：**29,582 ns/op、8,160 B/op、102 allocs/op**。
- 当前串行中位数约快 **30.5x**。
- 每次解码减少 **7,624 B** 和 **89 次分配**。

旧引擎数据只保留为历史对照；后续更新只需覆盖“当前结果”的环境、五轮原始数据与中位数。
