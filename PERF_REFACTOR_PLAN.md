# Cosy 请求解码/校验管线"编译式"重构方案

> 状态：P0–P3 全部完成，2026-08-21 完成 code review 修复轮（15 项）与热路径优化轮
> 分支：`refactor/compiled-codec`
> 最低 Go 版本：1.27.0（`encoding/json/v2` 正式可用）
>
> **决策记录**
> - 2026-07-03：自研范围限定为 map→struct 解码（`internal/structcodec`）与规则校验（`internal/rulecheck`）；
>   **不自研 JSON 解析器**，bytes→map 交给成熟实现。
> - 2026-08-21：bytes→map 采用标准库 `encoding/json/v2`，保留 v1 的字段匹配语义，同时**拒绝重复键与非法 UTF-8**；
>   最低 Go 版本升至 1.27.0。
> - 2026-08-21：code review 15 项确认缺陷全部修复并补回归测试（见 §8）。
>
> **实施进度**
> - ✅ P0 + P1（commit `c6e2844`）：`internal/structcodec` 编译式 map→struct 引擎替换 `mapstructure`（已从 go.mod 移除）。
> - ✅ P2（commit `4f6cfe7`）：`internal/rulecheck` 编译式规则引擎替换 `validator.ValidateMap`，validator/v10 保留为 fallback。
> - ✅ P3（commits `a591042` → `c3cf525`）：bytes→map 切 `encoding/json/v2`、请求体大小上限、模型预热。
> - ✅ Review 修复轮（commit `135e38b` 清理，commit `c6c7a65` 修复 15 项）。
> - ✅ 热路径优化轮：按计划字段查输入 map、编译期特化标量 / 嵌套 struct 解码器、无锁转换器注册表与计划缓存、
>   嵌入指针每次解码只克隆一次、RFC 3339 快路径、按 Content-Length 预分配读缓冲。

## 0. 目标

移除 `mitchellh/mapstructure`（已归档停止维护）与 `go-playground/validator/v10` 热路径上的冗余反射，
用"**每类型 / 每规则编译一次，缓存，执行 N 次**"的架构实现 cosy 自己的解码/校验引擎：

- 全平台（所有 GOOS/GOARCH）、全 Go 版本可用，只用 `unsafe.Pointer` 稳定 API，不绑 runtime 内部实现；
- 公开 API 不变：`map2struct.WeakDecode` 签名、`c.Payload`、`SetValidRules`、错误格式（`gin.H{key: rule}`）、`GetValidator()`；
- 典型写请求的框架侧开销从 ~82 µs 降到 ~3.7 µs（约 22×），内存分配从 276 次降到 34 次（§3 实测）。

## 1. 现状诊断（origin/main）

Create/Update/Custom 的固定管线（`chain.go`）：

```
JSON bytes → gin.H (encoding/json v1) → v.ValidateMap (validator/v10)
           → BeforeDecode hooks（可修改 Payload）
           → map2struct.WeakDecode (mapstructure) → GORM
```

旧实现的三处结构性开销，都无法在不换引擎的前提下修复：

1. **mapstructure 每次 `WeakDecode` 都 `NewDecoder`**：它的 API 把 `Result` 绑死在 decoder 创建时，decoder 天然不可复用、不可缓存。
2. **6 个 DecodeHook 对每个字段值都经 `reflect.Call` 跑一遍**（decimal / null.String / time / *time / pgtype.Date / *pgtype.Date），
   即使 99% 的字段与这些类型无关。
3. **`safety_text` 每次调用 `regexp.MustCompile` 两个正则**（`valid/safety_text.go`，§7.1 F3），校验阶段被它放大到 50 µs 量级。

另有两个已证实可远程触发的缺陷：decode hook 对非预期类型做无保护断言导致 panic（F1）、请求体无大小上限（F2），详见 §7.1。

## 2. 架构：编译一次，执行 N 次

```
┌──────────────────────────────────────────────────────┐
│ 编译层（每类型 / 每规则串只跑一次，结果全局缓存）          │
│   structcodec: reflect.Type  → decodePlan（闭包数组）    │
│   rulecheck:   "required,email" → []checkFn            │
├──────────────────────────────────────────────────────┤
│ 执行层                                                  │
│   闭包 + unsafe 偏移量直接写字段（全平台）                  │
│   未知 / 覆盖的 tag、'|' 语法 → validator.Var 兜底         │
└──────────────────────────────────────────────────────┘
```

### 2.1 `internal/structcodec`：map→struct

- 首次遇到类型 `T` 时扫描字段（json tag、`unsafe.Offsetof`、embedded squash），生成 `decodePlan` 以及
  精确名 / 小写名两张索引，以 `reflect.Type` 为 key 缓存（`cache.go`，`sync.Map` + 代数计数器，读路径无锁）。
  递归嵌入指针（`type Node struct{ *Node }`）按普通字段处理，编译必然终止。
- 解码时直接遍历输入 `map[string]any`（`gin.H` 等命名类型经 `reflect.Convert` 零拷贝转换），按索引命中计划字段：
  精确名优先、小写名兜底，字段表放在栈上的 32 槽 scratch 里，不再为每次解码复制两份 map。
- 标量（string / bool / 整数 / 浮点）与嵌套 struct 字段在编译期生成特化闭包，JSON 产出的输入形态不经 reflect 直接写入；
  其余形态回落到通用弱类型转换，语义不变。转换器注册表是 copy-on-write 快照，查询只是一次原子读。
- 弱类型转换在编译时按目标类型选定特化函数；运行时对输入先解一层非 nil 指针，再按 kind 转换，
  与 mapstructure 的 `reflect.Indirect` 语义一致。
- 特殊类型（`time.Time` / `*time.Time` / `decimal.Decimal` / `null.String` / `pgtype.Date` / `*pgtype.Date`）
  由 `registry.go` 的类型注册表直接绑定转换器；下游可用 `map2struct.RegisterTypeDecoder` 注册自定义类型。
- 容器语义与 mapstructure 对齐：slice / map 永远逐元素拷贝，不引用请求 payload；map 合并进已有值，空输入 map 不清空目标。
- 解码先写入副本、成功后整体赋值，失败不会留下半写的 struct（不变量 I4）。

### 2.2 `internal/rulecheck`：规则校验

- 缓存 key 是**规则字符串本身**（`"required,email"` → `[]checkFn`），跨请求跨端点复用，命中后 0 alloc。
- 内置零反射实现：required、omitempty、omitzero、omitnil、email、url、date、safety_text、hostname_port、min、max、oneof、dive。
  email 用 validator 的正则，hostname_port 用 RFC1123 正则，`date` / `safety_text` 与 `valid` 包共用同一份实现。
- 与 validator 的语义对齐点：`dive` 遍历 `[]any` 时元素按 validator 的"可空"语义（required 只拒 nil、omitempty 只跳过 nil），
  `[]string` 元素按零值语义；`omitzero` 一律零值语义；含 `|` 或 `0x2C` / `0x7C` 转义的 token 整体交给 validator；
  min / max 参数按 base-0 解析（接受 `0x10`）。
- **覆盖内置 tag 必须走 `cosy.RegisterValidation`**：它在 validator 上注册并调用 `rulecheck.Override`，
  之后该 tag 的规则全部转给 validator；直接对 `GetValidator()` 注册只影响 fallback 路径。
- `db_unique` 仍在校验管线外单独处理。`ValidateMap` 无错误时返回 nil map，调用方写入前需判空（`validate.go` 已处理）。

### 2.3 bytes→map：`encoding/json/v2`

`payload.go` 的 `decodeJSON` 使用：

```go
jsonv2.JoinOptions(
	jsonv1.DefaultOptionsV1(),          // 字段匹配大小写不敏感、null 保留预填值、兼容 ,string 等旧 tag
	jsontext.AllowDuplicateNames(false), // 重复键 → 错误（封住解析差异攻击）
	jsontext.AllowInvalidUTF8(false),    // 非法 UTF-8 → 错误（不再静默替换成 U+FFFD）
)
```

- 字符串永远拷贝（I3）、原始控制字符拒收、嵌套深度封顶 10000。
- `BindAndValid` 与 CRUD 管线走同一个 `decodeJSON`，结构体绑定语义与原来 gin 的 `ShouldBindJSON` 一致。
- 为什么不是 sonic：sonic 的构建约束是 `!go1.27`，在 Go 1.27 上退化成 compat 路径，实测比标准库 v1 还慢；
  json/v2 是标准库、无 JIT / unsafe 依赖、不随 Go 版本漂移。

### 2.4 预热

`model.Init` 在 `ResolvedModels()` 之后对所有已注册模型调用 `structcodec.Pretouch`，
覆盖 `cosy.RegisterModels`、`model.RegisterModels` 与 sandbox 三条注册路径，并保证在 converter 注册之后执行。

### 2.5 请求体大小上限（F2）

- `settings.Server.PayloadMaxBytes`：0 用默认 10 MiB，负数关闭；`Server.PayloadLimit()` 统一解析。
- `router.Init()` 安装 `limitRequestBody` 中间件，对**所有路由**的 `c.Request.Body` 包 `http.MaxBytesReader`，
  `bindJSONPayload` 内再包一层作为直接构造 gin.Context 时的兜底。
- 超限响应：CRUD 管线 406 `{"body": ...}`（沿用既有契约）；`BindAndValid` 413。

### 2.6 gin.Context 与 GORM

`model.RequestContext(c)`（`cosy.UseDB` 内部已调用）把池化的 `*gin.Context` 换成由 `c.Request.Context()` 派生、
不随请求取消的上下文，避免 `database/sql` 的 `awaitDone` goroutine 在请求结束后读到被复用的 `*gin.Context`（data race）。
`c.Set` 的值通过快照继续对 GORM hook 可见；调用方在 `*gin.Context` 之上加的 deadline 会被保留。
不要把 `*gin.Context` 或 `c.Request.Context()` 直接交给 `db.WithContext`。

### 2.7 明确不做的事

- **不自研机器码 JIT**。map→struct 的输入是已装箱的 `any`，开销大头是 map 遍历与 interface 拆箱，机器码省不掉，
  却要背上逐 Go 版本适配的维护债。
- **不自研 JSON 解析器 / 不做单遍融合解析**。`Payload gin.H` 是公开 API，BeforeDecode hooks 依赖它在校验后、解码前真实存在。
- **不保留旧引擎开关**。mapstructure 已从依赖中移除；兼容性靠 §4 的语义矩阵、差分测试与对抗语料保证。

## 3. 实测指标

同一台机器（Apple M5 Pro, darwin/arm64）、同一工具链（Go 1.27.0）、`-cpu=1 -count=3` 取中位数；
旧 = origin/main `9c00f11`（模块声明 go 1.26.5，`encoding/json` 走 v1 实现），新 = 本分支修复轮之后。
模型为 10 字段（含 `time.Time`、`decimal.Decimal`、`null.String`、slice），规则含 `safety_text`。

| 阶段 | 旧（origin/main） | 新 | 提升 |
|---|---|---|---|
| bytes → map | 2,477 ns / 54 allocs（encoding/json v1） | 1,884 ns / 24 allocs（json/v2） | 1.3× |
| 校验 | 53,737 ns / 139 allocs（含 F3：safety_text 每次重编译正则） | 466 ns / **0 allocs** | 115× |
| 校验（同规则、正则预编译） | 460 ns / 5 allocs（validator，`BenchmarkValidatorValidateMap`） | 178 ns / 0 allocs（`BenchmarkRulecheckValidateMap`） | 2.6× |
| map → struct | 23,376 ns / 83 allocs（mapstructure） | 857 ns / 10 allocs | 27× |
| **端到端** | **82,627 ns / 276 allocs** | **3,650 ns / 34 allocs** | **22.6×** |
| 端到端（18 核并行） | 80,288 ns | 3,221 ns | 24.9× |

说明：

- 校验阶段旧实现的绝对值被 F3 放大；第三行用 `internal/rulecheck` 的成对基准（4 条规则、正则预编译）给出引擎本身的对比。
- map → struct 从 P1 落地时的 2,270 ns / 60 allocs 降到 857 ns / 10 allocs：去掉每次解码的双 map 拷贝与逐值注册表查询后，
  剩下的 10 次分配是根对象的副本（I4）、slice 背板、3 个元素路径字符串和 `decimal` 内部的 `big.Int`。
- 端到端现在由 bytes → map 主导（~2 µs，json/v2 解到 `map[string]any` 的固有成本）；`Payload gin.H` 是公开 API，这一段不做融合解析。

## 4. 兼容性与测试

- **语义矩阵**（`map2struct/map2struct_test.go`）：弱类型转换全组合、embedded squash、时间三种输入形态、decimal、null.String、pgtype.Date 及指针变体；
  `regression_test.go` 钉住 review 修复的每一项（递归嵌入、对象数组进 map、接口字段报错、不别名 / 合并、typed 输入）。
- **差分测试**（`internal/rulecheck/rulecheck_test.go`、`parity_test.go`）：与 `validator.ValidateMap` 逐条对拍，含 `|` 语法、转义、
  hex、dive 可空语义、omitzero / omitnil、email / hostname_port 边界输入、`Override`。
- **JSON 语义矩阵**（`payload_semantics_test.go`）：std v1 / json/v2 严格 / 管线解码器在 11 个 D / E 语料上的行为记录，断言永不 panic。
- **fuzz**：`map2struct/fuzz_test.go`、`FuzzValidateMapNeverPanics`。
- **对抗语料**（§7.2）：A–F 全部有表驱动测试；`-race` 下通过（含 gin.Context→GORM 的竞争修复）。
- 公开 API 不变；行为变化集中在 §6。

## 5. 包结构

```
cosy/
  payload.go               # bytes→map：json/v2 选项、请求体上限、bind 错误分类
  validate.go              # validate / validateBatchUpdate / BindAndValid / RegisterValidation
  internal/structcodec/
    compile.go             # reflect.Type → decodePlan（递归类型安全）
    convert.go             # 弱类型转换矩阵、容器、特殊类型解码器
    decode.go              # Decode 入口、mapView、错误聚合
    registry.go            # 类型转换器注册表（Register / Unregister）
    cache.go               # 以 reflect.Type 为 key 的计划缓存
    pretouch.go            # 预热入口
  internal/rulecheck/
    compile.go             # "required,email" → []checkFn，'|' / 转义 / 覆盖路由
    builtin.go             # 内置 tag 的零反射实现（含 nullable 变体）
    validate.go            # ValidateMap、diveCheck
    fallback.go            # validator.Var 兜底
    override.go            # Override 注册表
  map2struct/              # 门面：WeakDecode、RegisterTypeDecoder
  model/model.go           # RequestContext、UseDB、Init 预热
  router/middleware.go     # limitRequestBody
  settings/server.go       # PayloadMaxBytes / PayloadLimit
```

## 6. 行为变化（changelog 要点）

| 变化 | 性质 |
|---|---|
| 请求体含重复键或非法 UTF-8 → 406 `{"body": ...}`（此前静默接受） | **BREAKING**，有意的安全收紧 |
| 最低 Go 版本 1.27.0 | **BREAKING** |
| decode hook 对非预期类型不再 panic，改为解码错误 → 406 | 安全修复（F1） |
| 请求体默认上限 10 MiB（`server.PayloadMaxBytes` 可调，负数关闭），对所有路由生效 | 安全修复（F2） |
| `BindAndValid`：超限 413、畸形 JSON 406（此前 500） | 修正 |
| 批量更新 `data` 不是对象 → 406 `{"data":"required"}`（此前 panic → 500） | 修正 |
| 覆盖内置 tag 需通过 `cosy.RegisterValidation` | 新 API，文档已注明 |
| `map2struct.RegisterTypeDecoder` 可注册自定义类型解码器 | 新 API |

## 7. 安全对抗检查

自研引擎直接处理不可信输入（HTTP 请求体），且使用 `unsafe` 绕过 Go 的类型 / 边界检查，安全对抗是验收门槛而非可选项。

### 7.0 威胁模型

- 攻击者完全控制请求体字节（结构、类型、数值、键名、嵌套深度、体积），以及 `SetValidRules` 覆盖不到的额外键。
- 模型结构体、`rules`、`columnMapping` 由开发者定义，视为可信；请求体一律不可信。
- 要防护的资产：进程可用性、内存安全、数据隔离（不跨请求泄露）、写入范围（不写未授权字段）。

### 7.1 已修复的现网问题

| 编号 | 问题 | 处置 |
|---|---|---|
| **F1** | decode hook 类型混淆 panic（`{"bio":123}` / `{"balance":true}` 直接 `interface conversion`） | 所有转换器改为检查型转换 + 返回 error；接口类型字段同样返回 error |
| **F2** | 无请求体大小限制 | `limitRequestBody` 中间件 + `bindJSONPayload` 兜底，默认 10 MiB |
| **F3** | `safety_text` 每次调用 `regexp.MustCompile` | 包级预编译，`valid` 与 `rulecheck` 共用一份实现 |

> `safety_text` 的两个正则为线性匹配、无嵌套量词，无 ReDoS；`db_unique` 走 GORM 参数化、列名来自开发者白名单，无 SQL 注入。

### 7.2 对抗语料库

判定标准统一为：**要么正确解码、要么返回错误——永不 panic、永不越界、永不静默写错字段。**

- **A. 类型混淆矩阵**：每种目标字段类型 × 7 种 JSON 值（string / number / bool / null / array / object / 缺失），unsafe 写入的偏移与宽度绝不错配。
- **B. 数值边界**：`1e309`、`NaN`、`-0`、超出 int64 的整数、高精度小数、时间戳溢出。
- **C. 深度与规模**：深层嵌套在 jsontext 深度上限处报错不 panic；超大 body 被上限拒绝；单请求内存 / CPU 有界。
- **D. Unicode 与字符串**：非法 UTF-8 拒收、控制字符拒收、解码出的字符串不引用请求缓冲区。
- **E. 键空间**：重复键拒收；未在 rules / 模型中声明的键被丢弃（I1）；`json:"-"` 不会被 payload 键 `"-"` 写入。
- **F. 规则引擎**：超长 rule、畸形 rule（`min=`、`oneof=` 空枚举、未闭合参数）、未知 tag 的 fallback、`|` 与转义语法。

### 7.3 必须持续成立的安全不变量

| 编号 | 不变量 | 验收 |
|---|---|---|
| **I1** | 只有同时出现在 `rules` 且存在于模型的键才允许写入 struct（防 mass-assignment） | `TestValidateRulecheckPreservesMassAssignmentInvariant` |
| **I2** | 任意输入下 unsafe 写入不越界、不错位 | 对抗语料在 `-race` 下跑；fuzz 长跑 |
| **I3** | 解码结果不引用请求缓冲区或跨请求共享内存 | json/v2 永远拷贝字符串；slice / map 逐元素拷贝；`TestJSONDecoderCopiesStrings` |
| **I4** | 任何转换 / 校验失败返回 error 并中止请求，不产生半解码的脏 struct | 解码先写副本；`TestWeakDecodeDoesNotAliasPayloadAndMergesMaps` 等 |
| **I5** | 单个请求无法导致 panic 或非线性 CPU / 内存 | A–F 语料 + 顶层 `recover` 兜底 |
| **I6** | validator fallback 的安全属性不弱于快路径 | 差分测试覆盖 fallback 路径 |

### 7.4 依赖与供应链

- `encoding/json/v2` 为标准库；`mitchellh/mapstructure`（已归档）已移除；validator/v10 只作 fallback。
- CI 加 `govulncheck`，fuzz target 纳入定期长跑（非仅 PR 时短跑）。

## 8. Code review 修复记录（2026-08-21）

15 项全部 CONFIRMED 并修复，回归测试随修复提交：

1. `walkFields` 递归嵌入指针无限递归 → 祖先链去重。
2. 对象数组解进 map 字段失败 → 逐元素重新反射。
3. json/v2 默认选项改变 `BindAndValid` 语义 → `DefaultOptionsV1` + 两个严格开关。
4. 规则 `|` / 转义 / hex 编译失败且被永久缓存 → 转 validator，limit 按 base-0 解析。
5. `dive` 对 `[]any` 元素的 required / omitempty 语义 → nullable 变体。
6. 接口类型字段 `reflect.Set` panic → 返回错误。
7. 可赋值快路径别名 payload、map 被替换 → 容器逐元素拷贝并合并，空 map 不清空。
8. `RequestContext` 丢掉派生上下文的 deadline → 重新施加。
9. `RequestContext` 丢掉 `c.Set` 的值 → 快照暴露。
10. 内置 tag 遮蔽自定义校验、email / hostname_port 与 validator 漂移 → `Override` + 正则对齐。
11. `omitzero` / `omitnil` 不短路 → 显式实现。
12. 批量更新 `data` 非对象 panic → 406。
13. `BindAndValid` 对 nil Validator panic、绑定错误答 500 → 判空 + 413 / 406 分类。
14. typed Go 输入（time / 指针 / struct→map / 根特殊类型）回归 → 全部恢复。
15. 请求体上限只覆盖 bind 路径 → router 中间件。

## 9. 基准复现与下一步

```bash
# 端到端与分段（根包）
go test -run '^$' -bench 'Pipeline|StdJSONToMap' -benchmem -count=3 -cpu=1 .
# 校验引擎成对基准
go test -run '^$' -bench 'Rulecheck|Validator' -benchmem -cpu=1 ./internal/rulecheck/
# map→struct 与并行
go test -run '^$' -bench 'WeakDecode' -benchmem ./map2struct/
```

旧基线用 `git archive origin/main` 解出的快照加同一份 `pipeline_bench_test.go`（把 `decodeJSON` 换成 `json.Unmarshal`、
`rulecheck.ValidateMap` 换成 `v.ValidateMap`）复现。

热路径优化轮已完成 review 列出的项目：按计划字段查输入 map、编译期特化解码器与无锁注册表、嵌入指针单次克隆、
`time.Parse(RFC3339Nano)` 快路径、读缓冲按 `Content-Length` 预分配。仍可做但收益有限：slice / map 元素路径改为出错时再拼接
（每个元素省 1 次分配）；`decimal.Decimal` 的 `big.Int` 分配属库内部。

## 10. 结论

"每类型 / 每规则编译一次 + 执行 N 次"的纯 Go 实现就拿到了 22× 的端到端提升（瓶颈 mapstructure 为 27×，校验 0 alloc），
全平台无条件可用、不绑 Go 版本、无第三方 JIT 依赖；bytes→map 交给标准库 `encoding/json/v2`，并借机把重复键与非法 UTF-8 挡在入口。
