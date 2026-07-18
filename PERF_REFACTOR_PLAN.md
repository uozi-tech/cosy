# Cosy 请求解码/校验管线"编译式"重构方案

> 状态：方案探索阶段，未实施
> 分支：`refactor/compiled-codec`
> 日期：2026-07-03
>
> **决策记录（2026-07-03）**：JSON 解析（bytes → map）固定采用 sonic 做可靠交付，
> **不自研解析器**；自研范围限定为 map→struct 解码（structcodec）与校验（rulecheck）。

## 0. 目标

移除 `mitchellh/mapstructure`（已归档停止维护）与 `go-playground/validator/v10` 热路径上的冗余反射，
参考 sonic 的 **"每类型编译一次 + 分层降级"** 架构，实现 cosy 自己的高性能解码/校验引擎：

- 全平台（所有 GOOS/GOARCH）、全 Go 版本可用，不绑 runtime 内部实现；
- 公开 API 与语义完全不变，**原有各包单元测试原样通过**；
- 典型写请求的框架侧开销从 ~28µs 降至 ~1.1µs（约 25×），内存分配从 151 次降至 25 次。

## 1. 现状诊断

Create/Update/Custom 的固定管线（`chain.go`）：

```
JSON bytes → gin.H (encoding/json) → ValidateMap (validator/v10)
           → BeforeDecode hooks（可修改 Payload）
           → map2struct.WeakDecode (mapstructure) → GORM
```

针对典型模型（10 字段，含 `time.Time`、`decimal.Decimal`、`null.String`、slice）的实测
（Apple M5 Pro, arm64, Go 1.26）：

| 阶段 | 现实现 | 耗时 | 内存分配 |
|---|---|---|---|
| bytes → map | encoding/json | 1,970 ns | 53 allocs |
| 校验 | `v.ValidateMap` | 1,148 ns | 15 allocs |
| map → struct | `map2struct.WeakDecode` | **25,090 ns** | 83 allocs |
| **合计** | | **~28 µs** | **151 allocs** |

**mapstructure 占整条管线 89% 的开销**，根因有二，且都无法在不换引擎的前提下修复：

1. `map2struct/map2struct.go` 每次调用 `WeakDecode` 都要 `NewDecoder`。这不是使用不当——
   mapstructure 的 API 把 `Result` 绑死在 decoder 创建时，decoder 天然不可复用、不可缓存。
2. `ComposeDecodeHookFunc` 组合的 6 个 hook（decimal / null.String / time / *time / pgtype.Date / *pgtype.Date），
   对**每一个字段值**都要经 `reflect.Call` 依次执行一遍，即使 99% 的字段与这些类型无关。

其他发现：

- `mitchellh/mapstructure` 仓库已归档（社区迁至 go-viper/mapstructure），依赖本身就应移除。
- `map2struct` 包目前**没有任何单元测试**——直接换引擎没有安全网，必须先补语义快照测试（见 P0）。
- 现有 hook 存在**已证实可远程触发的 panic**：`hook.go` 中 `ToNullableStringHookFunc` /
  `ToDecimalHookFunc` 对非预期输入类型做无保护的 `data.(string)` 断言。
  PoC：模型含 `null.String` 字段时请求体带 `{"bio": 123}`、含 `decimal.Decimal` 字段时带
  `{"balance": true}`，decode 阶段直接 `interface conversion` panic（详见 §7 对抗检查）。

## 2. sonic 的可借鉴点与跨平台现实

sonic v1.15.1（已是 cosy 经 gin 的间接依赖）的构建约束：

```go
//go:build (amd64 && go1.17 && !go1.27) || (arm64 && go1.20 && !go1.27)
```

sonic 的加速来自三件事：

1. **每类型一次的 JIT**：首次遇到某 struct 时把字段布局编译成机器码，缓存复用（`Pretouch` 可预热）；
2. **SIMD** 扫描原始字节；
3. **绕过 reflect**，按预计算偏移量直接读写内存。

其跨平台策略不是"到处生成机器码"，而是**分层降级**：amd64/arm64 走 JIT，
其他环境整包退化为 `encoding/json`（compat 层）。

关键教训藏在 `!go1.27` 里：sonic 依赖 `linkname` 和自研 loader 把机器码注册进 Go runtime，
因此**每个 Go 新版本都会先编译失败、等上游适配**。自研机器码 JIT 意味着 cosy 要背上这份持续维护成本。

而 sonic 三板斧中贡献大头的 ① 和 ③，**用纯 Go 就能实现**（json-iterator、goccy/go-json 均为先例）：
"闭包编译"——首次遇到类型时用一次反射生成执行计划（每字段一个闭包，内含预计算的
`unsafe.Offsetof` 偏移量和特化的转换函数），之后每次执行零反射。
只使用 `unsafe.Pointer` 稳定 API，全平台、全 Go 版本可用，无 linkname。

## 3. 架构：编译一次，执行 N 次

核心思想：把"每个请求做全量反射"变成"**每个类型/规则编译一次，缓存，执行 N 次**"。
JIT 的本质是编译缓存，机器码只是它的一种后端。

```
┌──────────────────────────────────────────────────────┐
│ 编译层（每类型 / 每规则串只跑一次，结果全局缓存）          │
│   structcodec: reflect.Type  → decodePlan（闭包数组）    │
│   rulecheck:   "required,email" → []checkFn            │
├──────────────────────────────────────────────────────┤
│ 执行后端                                                │
│   Tier 1（默认，全平台）  ：闭包 + unsafe 偏移量           │
│   Tier 2（可选，build tag）：amd64/arm64 机器码（预留）     │
│   Tier 0（兜底 / 对拍）   ：现有 mapstructure + validator  │
└──────────────────────────────────────────────────────┘
```

编译式引擎（Tier 1 等价物）在同一基准下的实测：

| 阶段 | 现状 | 编译式 | 提升 |
|---|---|---|---|
| bytes → map | 1,970 ns | 656 ns（sonic，自带 compat 降级） | 3× |
| 校验 | 1,148 ns | 44 ns，**0 alloc** | 26× |
| map → struct | 25,090 ns | 419 ns | **60×** |
| **合计** | ~28 µs / 151 allocs | **~1.1 µs / 25 allocs** | **~25×** |

### 3.1 组件一：`internal/structcodec` —— 替换 mapstructure（收益最大）

- 首次遇到类型 `T` 时扫描字段：json tag、`unsafe.Offsetof`、embedded squash，
  生成 `map[string]fieldDecoder`，以 `reflect.Type` 为 key 全局缓存（sync.Map）。
- 弱类型转换在编译时按 `(目标类型, 源 kind)` 选定特化转换函数
  （string→int、float64→bool 等各为独立小函数），运行时无类型判断分支树。
- 现有 6 个 DecodeHook 改为**类型注册表**：编译时发现字段类型是
  `decimal.Decimal` / `null.String` / `time.Time` / `pgtype.Date`（及指针形式）即直接绑定
  对应转换器；普通字段零 hook 开销。注册表保持公开，下游可注册自定义类型
  （对应现在修改 `hook.go` 的扩展能力）。
- 行为修正（有意变更，写入 changelog）：hook 中无保护类型断言导致的 panic
  改为返回 error。

### 3.2 组件二：`internal/rulecheck` —— 替换 `ValidateMap`

- rules 是每请求重建的 `gin.H`（`SetValidRules`），因此缓存 key 用**规则字符串本身**
  （`"required,email"` → 编译好的 `[]checkFn`），全局缓存，跨请求跨端点复用。
- 为 cosy 生态高频 tag 提供零反射实现：
  required、omitempty、email、url、date、safety_text、max、min、oneof、hostname_port、dive 等。
- **未知 tag 降级到 `v.Var()`**：validator/v10 不从依赖中删除，只是被热路径绕过。
  下游任意 tag、`RegisterValidation` 注册的自定义校验器全部照常工作。
- `db_unique` 本就在校验管线外单独处理，不受影响。

### 3.3 组件三：bytes → map 采用 sonic（已定稿，不自研）

- sonic 已在依赖树中；`sonic.Unmarshal` 在不支持的平台自动退化为 encoding/json；
  数字默认解析为 float64，与现状一致。
- 在 cosy 内直接调用（不要求用户给 gin 加 build tag），行为可控。
- **配置纪律**：必须使用 `ConfigDefault`（已确认其 `CopyString=true`，解码出的字符串
  为拷贝而非引用请求缓冲区）。禁止为性能切换 `ConfigFastest` 或关闭 `CopyString`——
  否则一旦未来引入 buffer pool，将造成跨请求内存别名（数据泄露），见 §7。

### 3.4 预热

`cosy.RegisterModels`（`db.go`）是现成的预热入口——boot 时对已注册模型预编译
decodePlan，等价于 sonic 的 `Pretouch`，消除首请求编译延迟。

### 3.5 明确不做的事

- **不自研机器码 JIT（Tier 2 暂缓）**。map→struct 阶段输入是已装箱的 `any`，
  闭包版 419ns 的大头是 map 遍历与 interface 拆箱，机器码省不掉这些，
  预估收益 < 2×，却要承担 sonic 式的逐 Go 版本适配维护债。
  架构上保留 `decodePlan` 中间表示，未来若有真实场景，
  以 build tag 隔离新增 codegen 后端即可，Tier 1 即天然 fallback。
- **不自研 JSON 解析器 / 不做单遍融合解析**（已定稿）。bytes→map 交给 sonic 保证可靠性；
  且管线语义要求 BeforeDecode hooks 可以在校验后、解码前修改 `c.Payload`，
  `Payload gin.H` 是公开 API，map 必须真实存在，decode 必须留在 hooks 之后。

## 4. 兼容性与测试策略

1. **P0 先补语义快照测试**：针对**现实现**编写语义矩阵测试并固化为金标准——
   弱类型转换全组合（string↔number↔bool、空串→零值）、embedded squash、
   时间三种输入形态（string / 毫秒时间戳 float64 / int64）、decimal（float64 / string）、
   null.String、pgtype.Date 及指针变体。
2. **对拍（differential testing）**：过渡期保留 mapstructure 路径，
   fuzz 随机生成 `map[string]any` + 目标 struct，断言新旧引擎输出一致。
3. **公开 API 一个都不动**：`map2struct.WeakDecode` 签名不变、
   `c.Payload` / `SetValidRules` / 错误格式（`gin.H{key: rule}`）不变、
   `GetValidator()` 继续返回真 validator 实例。现有 validate_test、api_test、
   integration test 原样通过。
4. **逃生舱**：保留一至两个版本的 `COSY_LEGACY_DECODER` 开关，
   线上出现语义差异可即时切回旧引擎。

## 5. 落地路线图

| 阶段 | 内容 | 预期收益 | 安全验收（详见 §7） |
|---|---|---|---|
| P0 | map2struct 语义矩阵测试 + 基准测试基线入库 | 安全网 | 对抗语料库建立（含 §7 PoC）；fuzz targets 搭建 |
| P1 | `internal/structcodec`，`WeakDecode` 换芯 | 管线 ~89% 开销消失（60×） | 类型混淆矩阵 A、数值边界 B 全通过；差分 fuzz 无 crash；`-race` 通过 |
| P2 | `internal/rulecheck`，validate 换芯（validator 保留为 fallback） | 校验 26×，0 alloc | mass-assignment 不变量 I1 测试；规则引擎对抗 F 通过 |
| P3 | bytes→map 切 sonic + `RegisterModels` 预热 + body 大小限制 | 解析 3×；堵住 F2 | 深度/规模对抗 C 通过；sonic 配置纪律 lint |
| P4（可选） | Tier 2 机器码后端（预留，暂缓） | 边际收益，按需评估 | — |

P1 是收益/风险比最高的一步：单独完成即可把典型写请求的框架侧 CPU 开销从 ~28µs 压到 ~4µs。

### 目标包结构

```
cosy/
  internal/structcodec/    # 编译式 map→struct 引擎（Tier 1）
    compile.go             # reflect.Type → decodePlan
    convert.go             # 弱类型转换矩阵（特化函数）
    registry.go            # 特殊类型注册表（decimal / null.* / time / pgtype）
    cache.go               # 以 reflect.Type 为 key 的计划缓存
  internal/rulecheck/      # 规则编译器
    compile.go             # "required,email" → []checkFn
    builtin.go             # 高频 tag 的零反射实现
    fallback.go            # 未知 tag → validator.Var
  map2struct/              # 门面不变：WeakDecode 换内部实现，公开类型注册入口
```

## 6. 风险清单

| 风险 | 应对 |
|---|---|
| mapstructure 弱类型语义长尾（`"1"/"t"`→bool、数字→string、空串→零值、slice 提升等） | P0 语义矩阵测试固化现状，逐条复刻，不靠印象 |
| unsafe 偏移量写内存的正确性 | fuzz + `-race` + 全类型覆盖；不触碰 runtime 内部，Go 升级零风险（对比 sonic 的 `!go1.27`） |
| 下游 tag 生态不可枚举 | 未知 tag 无条件走 validator fallback，兼容性生命线 |
| hook panic → error 属行为变更 | 有意修正（且是安全修复，PoC 已证实可远程触发，见 §7 F1），changelog 注明 |
| 首请求编译延迟 | `RegisterModels` 预热 + 懒编译兜底 |

## 7. 安全对抗检查

自研引擎直接处理**不可信的外部输入**（HTTP 请求体），且使用 `unsafe` 绕过 Go 的类型/边界安全网，
因此安全对抗不是可选项而是验收门槛。本节把攻击面拆成"必须证伪的不变量"与"必须通过的对抗语料"，
每条都要落成可复跑的测试进对应阶段的验收（见 §5 路线图）。

### 7.0 威胁模型

- **攻击者能力**：完全控制请求体字节（JSON 结构、类型、数值、键名、嵌套深度、体积），
  以及 `SetValidRules` 覆盖不到的额外键。
- **信任边界**：模型结构体、`rules`、`columnMapping` 由开发者定义，视为可信；请求体一律不可信。
- **要防护的资产**：进程可用性（不被单请求打挂）、内存安全（unsafe 不越界/不错位）、
  数据隔离（不跨请求泄露）、写入范围（不被写入未授权字段）。

### 7.1 已证实的现网问题（重构必须一并修掉，不得原样搬运）

| 编号 | 问题 | PoC | 处置 |
|---|---|---|---|
| **F1** | decode hook 类型混淆 panic：`null.String` 字段收到非字符串、`decimal.Decimal` 字段收到 bool，`data.(string)` 直接 panic | 本次已实测：`{"bio":123}` → `interface conversion: interface {} is float64, not string`；`{"balance":true}` → `...is bool, not string` | 新引擎所有类型转换器**禁止裸类型断言**，一律走"检查型转换 + 返回 error"；补进 §7.2 矩阵 A |
| **F2** | 无请求体大小限制：`ShouldBindJSON` 前未见 `MaxBytesReader`/`ContentLength` 卡口 | 超大 body 直接进 sonic 解析并在内存里展开成 map | P3 在 bind 前加可配置 `http.MaxBytesReader`（默认上限 + 可覆盖）；与 §7.2 对抗 C 联动 |
| **F3** | `safety_text` 每次调用 `regexp.MustCompile` 两个正则（`valid/safety_text.go`）：既是性能问题也放大 CPU 打击面 | 高频命中该 tag 的字段会反复编译正则 | 正则提到包级 `var` 预编译一次；改写为 rulecheck 内置检查时同样只编译一次 |

> 注：`safety_text` 现有两个正则经审查为线性匹配、无嵌套量词，**无 ReDoS 灾难回溯**；
> `db_unique` 走 GORM 参数化（`db.Or(column, value)`），列名来自开发者定义的白名单，
> **无 SQL 注入**。这两点属于"已检查确认安全"，重构时保持现状即可，不要在改写中引入字符串拼接。

### 7.2 对抗语料库（P0 建立，P1/P2/P3 分别验收）

以下每组都要有对应的表驱动测试与 fuzz target。判定标准统一为：
**要么正确解码、要么返回错误——永不 panic、永不越界、永不静默写错字段。**

- **A. 类型混淆矩阵（P1，unsafe 正确性核心）**
  对每个目标字段类型（`string/int/uint/float/bool/time.Time/*time.Time/decimal.Decimal/null.String/pgtype.Date/slice/嵌套 struct`）
  逐一喂入 7 种 JSON 值（string、number、bool、null、array、object、缺失）。
  重点验证 unsafe 写入的字段偏移与类型宽度绝不错配（写 `bool` 的地方来了 `[]any` 不能按指针宽度乱写）。
- **B. 数值边界（P1）**
  `1e309`(Inf)、`NaN`、`-0`、超出 int64 的整数、`1e-400`(下溢)、高精度小数（decimal 精度/舍入）、
  时间戳负值与溢出。断言与 mapstructure/cast 现状一致或明确记为有意改进。
- **C. 深度与规模（P3，DoS）**
  深层嵌套 `[[[[...]]]]`（栈溢出/递归爆栈）、超宽 map（百万键）、超长 slice、超大 body。
  验收：body 上限生效、递归有深度上限（sonic 侧 + 自研 map 遍历侧）、单请求内存/CPU 有界。
- **D. Unicode 与字符串安全（P1）**
  非法 UTF-8、超长 key、含 NUL、同形/组合字符、超大 key 触发 map rehash。
  校验 sonic `CopyString=true` 语义下 decode 出的 string 不再引用请求缓冲区。
- **E. 键空间攻击（P1/P2）**
  重复键（`{"a":1,"a":2}` 取值语义须与 sonic 一致）、大小写差异键、
  未在 rules/模型中声明的额外键（应被丢弃，见不变量 I1）、`__proto__` 等特殊键名（Go 无原型链，确认无副作用）。
- **F. 规则引擎对抗（P2）**
  超长 rule 字符串、畸形 rule（`"min="`、`"oneof="` 空枚举、未闭合参数）、
  未知 tag 走 validator fallback 的正确性、rule 缓存 key 的注入（确保规则字符串不会被拼进任何可执行上下文）。

### 7.3 必须持续成立的安全不变量

| 编号 | 不变量 | 为什么 | 如何验收 |
|---|---|---|---|
| **I1** | **写入范围收敛**：只有同时出现在 `rules` 且存在于模型的键才允许写入 struct | 防 mass-assignment / 越权改字段（如客户端偷传 `is_admin`、`balance`）。现有 `validate()` 已做键过滤，重构后必须**保持等价或更严**，绝不能因为解码更快而放宽 | 专项测试：payload 带模型有但 rules 无的字段，断言解码后该字段为零值 |
| **I2** | **内存安全**：任意输入下 unsafe 写入不越界、不错位、不写只读内存 | unsafe 是本方案唯一能造成内存损坏的地方 | 全部对抗语料在 `-race` 下跑；`go test -fuzz` 长跑；CI 加 `-gcflags=-d=checkptr` |
| **I3** | **无跨请求别名**：解码结果不得引用请求缓冲区或跨请求共享的可变内存 | 防止一个请求的数据出现在另一个响应里 | 保持 sonic `CopyString=true`；禁用 buffer 复用直到有测试证明安全；§3.3 配置纪律 lint |
| **I4** | **失败即拒绝**：任何转换/校验失败返回 error 并中止该请求，不产生"半解码"的脏 struct 落库 | 部分写入的模型进 GORM 会造成数据污染 | 注入中途失败的 payload，断言不进入 `GormAction` |
| **I5** | **可用性有界**：单个请求无法导致 panic、无法导致非线性 CPU/内存 | 参考 F1（panic）、F2/C（资源耗尽） | A–F 语料 + 顶层 `recover` 兜底（defense-in-depth，不作为唯一防线） |
| **I6** | **降级路径同样安全**：validator fallback、sonic 在非 amd64/arm64 的 encoding/json 降级，安全属性不弱于快路径 | 防"快路径安全、慢路径有洞"的不对称 | 对抗语料在 `COSY_LEGACY_DECODER=1` 与非 JIT 构建下各跑一遍 |

### 7.4 差分对抗（新旧引擎对拍）

过渡期保留 mapstructure/validator 旧路径，用 fuzz 生成的对抗语料**同时喂新旧两条路径**：

- 两者都成功 → 断言输出 struct 逐字段深等；
- 一者报错一者成功 → 进人工清单，判定属"有意改进"（如 F1 的 panic→error）还是"回归 bug"；
- **绝不允许**新引擎在旧引擎拒绝的输入上"成功"写入了额外字段（mass-assignment 回归）。

差分测试是把 §7.2 语料自动放大的手段，而非替代——A/E/F 里的定向用例仍要手写断言。

### 7.5 依赖与供应链

- sonic 版本纳入 `go.mod` 锁定 + Dependabot 关注其安全公告；关注其 `!go1.27` 约束在 Go 升级时的降级行为，
  确保降级到 encoding/json 时 §7.3 不变量仍成立（对应 I6）。
- 移除 `mitchellh/mapstructure`（已归档、不再收安全补丁）是本次重构的安全收益之一。
- CI 加 `govulncheck`，把上述 fuzz target 纳入定期长跑（非仅 PR 时短跑）。

## 8. 基准复现

基准代码位于探索阶段 scratchpad（`bench/bench_test.go`），P0 阶段将整理进仓库
（建议 `internal/structcodec/bench_test.go`）作为性能基线，包含：

- `BenchmarkStdJSONToMap` / `BenchmarkSonicJSONToMap`
- `BenchmarkMapstructureWeakDecode` / `BenchmarkCompiledClosureDecode`
- `BenchmarkValidatorValidateMap` / `BenchmarkCompiledValidate`
- `BenchmarkSonicBytesToStruct`（bytes→struct 上界参考）

## 9. 结论

参考 sonic 的正确姿势不是复刻它的机器码生成，而是复刻它的架构——
**"每类型编译一次 + 分层降级"**。纯 Go 闭包编译后端即可获得约 25× 的整体提升
（瓶颈 mapstructure 为 60×），全平台无条件可用、不绑 Go 版本；
机器码后端作为预留可选层，待有真实场景再评估。
