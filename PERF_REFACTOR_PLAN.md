# Cosy 请求解码与校验管线设计说明

> 状态：核心实现、兼容性修复和热路径优化均已完成。
> 最低 Go 版本：1.27.0。

面向使用者的完整基准、实现细节与迁移说明见 [`docs/performance/v1.35.0.md`](docs/performance/v1.35.0.md)。

## 目标与结论

Create、Update、Custom 和 BatchUpdate 的请求管线已经改为“编译一次、执行多次”的纯 Go 实现：

```text
JSON bytes
  -> encoding/json/v2 -> gin.H
  -> internal/rulecheck 校验
  -> BeforeDecode hooks
  -> map2struct.WeakDecode -> internal/structcodec
  -> GORM
```

- `github.com/mitchellh/mapstructure` 已移除；`cosy/map2struct` 继续作为兼容门面，内部调用 `structcodec.Decode`。
- `validator/v10` 不再承担常用规则的热路径，只处理未知规则、覆盖规则和复杂语法的 fallback。
- JSON 解析使用标准库 `encoding/json/v2`，不自研 JSON 解析器，也不引入 JIT 或 runtime 私有接口。
- `Payload gin.H`、`map2struct.WeakDecode`、`SetValidRules`、`GetValidator` 等主要调用方式保持不变。

典型写请求的框架侧开销由约 82.6 us 降至约 3.9 us，分配由 276 次降至 34 次。当前瓶颈主要是 JSON 解码到 `map[string]any`；由于 hooks 和公开 API 依赖 `Payload`，不做单遍融合解析。

## 实现设计

### map 到 struct

`internal/structcodec` 按目标 `reflect.Type` 编译并缓存解码计划：

- 扫描 JSON tag、字段偏移和 embedded squash，生成字段索引与特化解码闭包。
- 字符串、布尔、整数、浮点和嵌套 struct 使用编译期选定的快路径；其他输入回落到通用弱类型转换。
- `time.Time`、`decimal.Decimal`、`null.String`、`pgtype.Date` 等特殊类型由注册表绑定转换器。
- `map2struct.RegisterTypeDecoder` 可注册下游自定义类型，注册后会使已有计划失效并重新编译。
- slice 和 map 始终逐元素复制，不与请求 payload 共享底层数据；map 合并进已有值，空输入不清空目标。
- 解码先写入副本，全部成功后才赋给目标，失败不会留下半写对象。
- 计划缓存与转换器注册表的读取路径无锁；递归嵌入指针在编译和执行时均有终止保护。

### 规则校验

`internal/rulecheck` 以完整规则字符串为 key 编译并缓存检查函数：

- 内置实现 `required`、`omitempty`、`omitzero`、`omitnil`、`email`、`url`、`date`、`safety_text`、`hostname_port`、`min`、`max`、`oneof` 和 `dive`。
- 命中缓存后的常用规则校验为 0 alloc。
- 含 `|`、转义 token、未知 tag 或被覆盖 tag 的规则交给 `validator.Var`，避免重复实现 validator 的完整语法。
- 覆盖内置 tag 必须通过 `cosy.RegisterValidation`，以同时更新 validator 和 rulecheck 的路由。
- `db_unique` 仍在规则引擎之外通过现有 GORM 管线处理。

### JSON、请求体与上下文

`decodeJSON` 使用 `encoding/json/v2`，叠加 v1 兼容选项，并显式拒绝重复键与非法 UTF-8。字符串由解析器复制，不引用请求缓冲区；原始控制字符和超过标准库深度上限的输入会返回错误。

`settings.Server.PayloadMaxBytes` 控制请求体上限：

- `0` 使用默认 10 MiB。
- 正数使用指定字节数。
- 负数关闭限制。

路由中间件限制所有请求体，绑定函数再提供一层兜底。CRUD 超限沿用 406 错误契约，`BindAndValid` 超限返回 413。

模型注册完成后会预热 structcodec 计划。`model.RequestContext` 将池化的 `*gin.Context` 转成独立上下文再交给 GORM，同时保留 deadline 和 `c.Set` 值，避免请求结束后的上下文复用竞争。

## 兼容性变化

| 变化 | 性质 |
|---|---|
| 重复 JSON 键或非法 UTF-8 返回 406 | 有意的安全收紧 |
| 最低 Go 版本调整为 1.27.0 | Breaking |
| 默认请求体上限为 10 MiB | 安全修复，可配置 |
| decode 类型错误返回 406，不再 panic | 安全修复 |
| `BindAndValid` 超限返回 413，畸形 JSON 返回 406 | 错误分类修正 |
| BatchUpdate 的 `data` 非对象时返回 406 | panic 修复 |
| 内置规则覆盖改用 `cosy.RegisterValidation` | 新扩展入口 |
| 新增 `map2struct.RegisterTypeDecoder` | 新扩展入口 |
| 移除旧 decode hook 的 7 个导出符号 | Breaking；能力由类型注册表替代 |

## 安全不变量与测试

请求体、键名、值类型、数值范围和嵌套结构均视为不可信；模型、rules 和 column mapping 由开发者定义并视为可信。

必须持续满足以下不变量：

| 不变量 | 验收方式 |
|---|---|
| 只有同时存在于 rules 和模型中的字段可以写入，防止 mass assignment | 请求管线回归测试 |
| 任意输入下 unsafe 写入不越界、不错位 | 类型混淆矩阵、边界语料、race、fuzz |
| 解码结果不引用请求缓冲区或跨请求共享容器 | JSON 字符串复制与容器别名测试 |
| 转换或校验失败不产生半解码对象 | 事务式赋值回归测试 |
| 单请求不能触发 panic 或非线性资源消耗 | 对抗语料与请求体上限测试 |
| validator fallback 不弱于原有语义 | 与 `validator.ValidateMap` 的差分测试 |

测试覆盖包括弱类型转换矩阵、embedded squash、特殊类型、递归嵌入、typed Go 输入、JSON 严格语义、规则差分、批量更新错误输入以及 gin.Context 到 GORM 的竞争修复。

CI 在 push、pull request 和定时任务中运行普通测试与 `govulncheck`；每周定时对以下目标各执行 10 分钟 fuzz：

- `map2struct.FuzzWeakDecodeNeverPanics`
- `rulecheck.FuzzValidateMapNeverPanics`

## 性能基线

参考环境为 Apple M5 Pro、darwin/arm64、Go 1.27.0，`-cpu=1 -count=5` 取中位数。旧基线为引入本次重构前的 v1.34.x 请求管线。

| 阶段 | 旧实现 | 当前实现 | 提升 |
|---|---:|---:|---:|
| JSON bytes -> map | 2,477 ns / 54 allocs | 2,076 ns / 24 allocs | 1.19x |
| 规则校验 | 53,737 ns / 139 allocs | 466.2 ns / 0 allocs | 115.3x |
| 规则引擎成对基准 | 444.8 ns / 5 allocs | 124.6 ns / 0 allocs | 3.6x |
| map -> struct | 23,376 ns / 83 allocs | 886.7 ns / 10 allocs | 26.4x |
| 端到端 | 82,627 ns / 276 allocs | 3,888 ns / 34 allocs | 21.3x |

复现命令：

```bash
go test -run '^$' -bench 'Pipeline|StdJSONToMap' -benchmem -count=5 -cpu=1 .
go test -run '^$' -bench 'Rulecheck|Validator' -benchmem -count=5 -cpu=1 ./internal/rulecheck/
go test -run '^$' -bench 'WeakDecode' -benchmem -count=5 -cpu=1 ./map2struct/
```

进一步优化的收益已经有限：slice 和 map 的元素路径可改为仅在出错时拼接；`decimal.Decimal` 的 `big.Int` 分配属于依赖内部。除非公开 API 允许取消中间 `Payload`，否则不考虑融合 JSON 解析与 struct 写入。

## 代码位置

```text
payload.go                 JSON 解码、请求体读取与错误分类
validate.go                请求校验、BindAndValid、规则扩展入口
internal/structcodec/      编译式 map 到 struct 引擎
internal/rulecheck/        编译式规则引擎与 validator fallback
map2struct/                兼容门面与类型解码器注册入口
model/model.go             模型预热与独立请求上下文
router/middleware.go       全路由请求体限制
settings/server.go         请求体上限配置
```
